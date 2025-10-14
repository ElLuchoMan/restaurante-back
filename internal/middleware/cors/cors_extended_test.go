package cors

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

func TestCORS_CustomAllowedOrigins_FromEnv(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, https://api.example.com")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://example.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", resp.Header.Get("Vary"))
}

func TestCORS_CustomAllowedOrigins_NotAllowed(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://evil.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()

	assert.NotEqual(t, "https://evil.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_AllowAll_DevMode(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	web.BConfig.RunMode = "dev"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://any-origin.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
}

func TestCORS_AllowAll_OPTIONS(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	web.BConfig.RunMode = "dev"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.Header.Set("Origin", "http://any-origin.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_EmptyOriginsInEnv(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Setenv("CORS_ALLOWED_ORIGINS", " , , ")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://localhost")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()

	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_MultipleOriginsInEnv(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, HTTPS://API.EXAMPLE.COM, http://localhost:3000")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://example.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://api.example.com")

	ctx = context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp = w.Result()
	assert.Equal(t, "https://api.example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_ProdModeWithoutEnv_UsesDefaults(t *testing.T) {

	originalEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalRunMode := web.BConfig.RunMode

	defer func() {
		if originalEnv != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalEnv)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		web.BConfig.RunMode = originalRunMode
	}()

	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://localhost:4200")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()

	assert.Equal(t, "http://localhost:4200", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", resp.Header.Get("Vary"))
}

func TestCORS_VaryHeader(t *testing.T) {

	originalRunMode := web.BConfig.RunMode
	defer func() {
		web.BConfig.RunMode = originalRunMode
	}()

	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://localhost")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()

	assert.Equal(t, "Origin", resp.Header.Get("Vary"))
}
