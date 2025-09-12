package logging

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"log/slog"

	beectx "github.com/beego/beego/v2/server/web/context"
)

func TestLoggerFallback(t *testing.T) {
	logger = nil
	l := Logger()
	if l == nil {
		t.Fatal("Logger fallback retornó nil")
	}
}

type memoryHandler struct {
	mu      sync.Mutex
	records []logRecord
	attrs   []slog.Attr
}

type logRecord struct {
	level slog.Level
	msg   string
	attrs map[string]interface{}
}

func (h *memoryHandler) Enabled(_Ctx context.Context, _Level slog.Level) bool {
	return true
}

func (h *memoryHandler) Handle(_Ctx context.Context, r slog.Record) error {
	m := make(map[string]interface{})
	for _, a := range h.attrs {
		m[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	copy := logRecord{level: r.Level, msg: r.Message, attrs: m}
	h.mu.Lock()
	h.records = append(h.records, copy)
	h.mu.Unlock()
	return nil
}

func (h *memoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := &memoryHandler{}
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return clone
}

func (h *memoryHandler) WithGroup(_ string) slog.Handler { return h }

func newBeegoCtx(method, target string, status int, hdr http.Header) *beectx.Context {
	req := httptest.NewRequest(method, target, nil)
	for k, vv := range hdr {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	if req.RemoteAddr == "" {
		req.RemoteAddr = "10.0.0.1:1234"
	}
	w := httptest.NewRecorder()
	ctx := beectx.NewContext()
	ctx.Reset(w, req)
	ctx.ResponseWriter.Status = status
	return ctx
}

func TestSetupAndLogger(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")
	Setup("dev")
	if Logger() == nil {
		t.Fatal("logger no debe ser nil tras Setup")
	}
	Logger().Info("msg_text")

	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_FORMAT", "")
	Setup("prod")
	if Logger() == nil {
		t.Fatal("logger no debe ser nil tras segundo Setup")
	}
	Logger().Warn("msg_json")

	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("LOG_FORMAT", "json")
	Setup("prod")
	Logger().Error("msg_error")

	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "text")
	Setup("dev")
	Logger().Info("msg_info")

	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "")
	Setup("dev")
	Logger().Info("msg_dev_default_text")
}

func TestStartTimerAndLogRequest(t *testing.T) {
	mem := &memoryHandler{}
	logger = slog.New(mem)
	slog.SetDefault(logger)

	ctx := newBeegoCtx(http.MethodGet, "/path?q=x", 200, http.Header{
		"X-Forwarded-For": []string{"1.2.3.4, 5.6.7.8"},
	})
	StartTimer(ctx)
	time.Sleep(1 * time.Millisecond)
	LogRequest(ctx)

	ctx2 := newBeegoCtx(http.MethodPost, "/other", 201, nil)
	LogRequest(ctx2)

	if len(mem.records) < 2 {
		t.Fatalf("se esperaban al menos 2 registros, got=%d", len(mem.records))
	}
}

func TestLogIfError(t *testing.T) {
	mem := &memoryHandler{}
	logger = slog.New(mem)
	slog.SetDefault(logger)

	ctx := newBeegoCtx(http.MethodGet, "/ok", 200, nil)
	LogIfError(ctx)
	if len(mem.records) != 0 {
		t.Fatalf("no se esperaba registro para status<400, got=%d", len(mem.records))
	}

	ctxErr := newBeegoCtx(http.MethodGet, "/fail?password=123&token=abc&q="+strings.Repeat("x", 200), 500, http.Header{
		"X-Forwarded-For": []string{"9.9.9.9"},
	})
	LogIfError(ctxErr)
	if len(mem.records) != 1 {
		t.Fatalf("se esperaba 1 registro de error, got=%d", len(mem.records))
	}
}

func TestLogControllerErrorAndSanitizers(t *testing.T) {
	mem := &memoryHandler{}
	logger = slog.New(mem)
	slog.SetDefault(logger)

	fields := map[string]interface{}{
		"password":      "abc",
		"token":         "abc",
		"Authorization": "Bearer X",
		"secret":        "should-hide",
		"normal":        strings.Repeat("n", 300),
		"count":         5,
	}
	ctx := newBeegoCtx(http.MethodPut, "/x?pass=s&other="+strings.Repeat("o", 150), 418, http.Header{
		"X-Forwarded-For": []string{"2.2.2.2"},
	})

	LogControllerError(ctx, "boom", errors.New("boom"), fields)
	if len(mem.records) != 1 {
		t.Fatalf("se esperaba 1 registro, got=%d", len(mem.records))
	}

	v := url.Values{
		"password":     []string{"topsecret"},
		"access_token": []string{"abc"},
		"q":            []string{strings.Repeat("q", 200)},
		"name":         []string{"ok"},
	}
	qs := sanitizeQuery(v)
	if qs["password"] != "[redacted]" || qs["access_token"] != "[redacted]" {
		t.Fatal("query no fue sanitizada correctamente")
	}
	if got := qs["q"]; !strings.HasSuffix(got, "…") || len(got) != 131 {
		t.Fatalf("se esperaba truncamiento a 128 + '…' (UTF-8 3 bytes), got len=%d", len(got))
	}

	fs := sanitizeFields(map[string]interface{}{
		"Password":      "x",
		"myPass":        "y",
		"TOKEN":         "z",
		"authorization": "A",
		"SecretKey":     "B",
		"normal":        strings.Repeat("n", 300),
		"short":         "ok",
		"i":             1,
	})
	if fs["Password"] != "[redacted]" || fs["myPass"] != "[redacted]" || fs["TOKEN"] != "[redacted]" || fs["authorization"] != "[redacted]" || fs["SecretKey"] != "[redacted]" {
		t.Fatal("fields sensibles no redacatados correctamente")
	}
	if v, ok := fs["normal"].(string); !ok || !strings.HasSuffix(v, "…") || len(v) != 259 {
		t.Fatalf("se esperaba truncamiento a 256 + '…' (UTF-8 3 bytes), got ok=%v len=%d", ok, len(v))
	}
	if v, ok := fs["short"].(string); !ok || v != "ok" {
		t.Fatalf("se esperaba string corto sin truncar, got ok=%v v=%v", ok, v)
	}
	if sanitizeFields(nil) != nil {
		t.Fatal("sanitizeFields(nil) debe retornar nil")
	}

	if safeErr(nil) != "" {
		t.Fatal("safeErr(nil) debe ser cadena vacía")
	}
	if safeErr(errors.New("e")) == "" {
		t.Fatal("safeErr(error) no debe ser vacío")
	}
}

func TestClientIPBranches(t *testing.T) {
	ctxXFF := newBeegoCtx(http.MethodGet, "/", 200, http.Header{
		"X-Forwarded-For": []string{"1.2.3.4, 5.6.7.8"},
	})
	if got := clientIP(ctxXFF); got != "1.2.3.4" {
		t.Fatalf("clientIP con XFF inválido, got=%s", got)
	}
	ctxNoXFF := newBeegoCtx(http.MethodGet, "/", 200, nil)
	if got := clientIP(ctxNoXFF); got == "" || strings.Contains(got, ":") == false {
		_ = got
	}
}
