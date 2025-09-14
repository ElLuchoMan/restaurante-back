package login

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestValidateToken_Missing(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestValidateToken_BearerNoSpace(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	r.Header.Set("Authorization", "invalidtoken")
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	r.Header.Set("Authorization", "Bearer invalid")
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestValidateToken_PublicRouteWithSlash(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/productos/", nil)
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestValidateToken_RestaurantesPublicGet(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes", nil)
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestValidateToken_RestaurantesSearchPublicGet(t *testing.T) {
	ctx := context.NewContext()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search", nil)
	ctx.Reset(w, r)
	ValidateToken(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
