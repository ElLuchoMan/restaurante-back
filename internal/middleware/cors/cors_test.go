package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestPreflightOptionsReturns204AndHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/restaurante/v1/login", nil)
	r.Header.Set("Origin", "capacitor://localhost")
	r.Header.Set("Access-Control-Request-Method", "POST")
	r.Header.Set("Access-Control-Request-Headers", "content-type, x-correlation-id")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "capacitor://localhost" {
		t.Fatalf("missing or wrong A-C-Allow-Origin header: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing A-C-Allow-Methods")
	}
	if resp.Header.Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("missing A-C-Allow-Headers")
	}
	if resp.Header.Get("Access-Control-Expose-Headers") == "" {
		t.Fatalf("missing A-C-Expose-Headers")
	}
}

func TestGetIncludesCorsHeadersForAllowedOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/login", nil)
	r.Header.Set("Origin", "https://localhost")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "https://localhost" {
		t.Fatalf("missing or wrong A-C-Allow-Origin header")
	}
	if resp.Header.Get("Access-Control-Expose-Headers") == "" {
		t.Fatalf("missing A-C-Expose-Headers")
	}
}

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{
			name:    "Origin allowed",
			origin:  "http://localhost",
			allowed: []string{"http://localhost", "https://example.com"},
			want:    true,
		},
		{
			name:    "Origin not allowed",
			origin:  "http://evil.com",
			allowed: []string{"http://localhost", "https://example.com"},
			want:    false,
		},
		{
			name:    "Empty allowed list",
			origin:  "http://localhost",
			allowed: []string{},
			want:    false,
		},
		{
			name:    "Case sensitive match",
			origin:  "http://localhost",
			allowed: []string{"HTTP://LOCALHOST"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedOrigin(tt.origin, tt.allowed); got != tt.want {
				t.Errorf("isAllowedOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/login", nil)
	// No se establece el header Origin

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// No debe haber headers CORS si no hay Origin
	if resp.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("should not have A-C-Allow-Origin when no Origin header")
	}
}

func TestCORS_UnallowedOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/login", nil)
	r.Header.Set("Origin", "http://evil.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// No debe permitir origenes no autorizados
	if resp.Header.Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Fatalf("should not allow unauthorized origin")
	}
}

func TestCORS_CaseInsensitiveOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/login", nil)
	r.Header.Set("Origin", "HTTP://LOCALHOST")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// El origen en mayúsculas se normaliza a minúsculas y coincide con "http://localhost"
	// El header devuelto es el origen original tal como vino
	if resp.Header.Get("Access-Control-Allow-Origin") != "HTTP://LOCALHOST" && resp.Header.Get("Access-Control-Allow-Origin") != "http://localhost" {
		t.Fatalf("should allow case-insensitive origin, got: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DefaultAllowedOrigins(t *testing.T) {
	allowedOrigins := []string{
		"capacitor://localhost",
		"http://localhost",
		"http://localhost:4200",
		"https://localhost",
	}

	for _, origin := range allowedOrigins {
		t.Run(origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.Header.Set("Origin", origin)

			ctx := context.NewContext()
			ctx.Reset(w, r)
			CORS()(ctx)

			resp := w.Result()
			if resp.Header.Get("Access-Control-Allow-Origin") != origin {
				t.Errorf("expected origin %q to be allowed, got: %q", origin, resp.Header.Get("Access-Control-Allow-Origin"))
			}
		})
	}
}
