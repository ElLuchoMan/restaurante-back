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

// Tests adicionales para aumentar cobertura de CORS middleware

func TestCORS_CustomAllowedOrigins_FromEnv(t *testing.T) {
	// Guardar valores originales
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

	// Configurar origenes custom
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
	// Guardar valores originales
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

	// Configurar origenes custom
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://evil.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// No debe permitir origenes no autorizados
	assert.NotEqual(t, "https://evil.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_AllowAll_DevMode(t *testing.T) {
	// Guardar valores originales
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

	// Limpiar env y configurar modo dev
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	web.BConfig.RunMode = "dev"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://any-origin.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// En modo dev sin CORS_ALLOWED_ORIGINS, debe permitir cualquier origen
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
}

func TestCORS_AllowAll_OPTIONS(t *testing.T) {
	// Guardar valores originales
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

	// Configurar modo dev
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
	// Guardar valores originales
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

	// Configurar con origenes vacíos y espacios
	os.Setenv("CORS_ALLOWED_ORIGINS", " , , ")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://localhost")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// No debe permitir ningún origen si la lista está vacía después de limpiar
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_MultipleOriginsInEnv(t *testing.T) {
	// Guardar valores originales
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

	// Configurar múltiples origenes
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, HTTPS://API.EXAMPLE.COM, http://localhost:3000")
	web.BConfig.RunMode = "prod"

	// Test primer origen (normalizado a lowercase)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://example.com")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))

	// Test segundo origen (uppercase normalizado)
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
	// Guardar valores originales
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

	// Configurar prod sin env var
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	web.BConfig.RunMode = "prod"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "http://localhost:4200")

	ctx := context.NewContext()
	ctx.Reset(w, r)
	CORS()(ctx)

	resp := w.Result()
	// En prod sin env var, debe usar origenes por defecto
	assert.Equal(t, "http://localhost:4200", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", resp.Header.Get("Vary"))
}

func TestCORS_VaryHeader(t *testing.T) {
	// Guardar valores originales
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
	// Vary header debe estar presente cuando no es allow-all
	assert.Equal(t, "Origin", resp.Header.Get("Vary"))
}
