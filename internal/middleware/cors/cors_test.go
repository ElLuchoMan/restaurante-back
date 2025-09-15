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
