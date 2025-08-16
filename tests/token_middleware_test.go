package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	. "github.com/smartystreets/goconvey/convey"
)

// TestProtectedEndpointWithoutToken ensures that endpoints protected by the token middleware return 401 when no token is provided.
func TestProtectedEndpointWithoutToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("GET /restaurante/v1/clientes without token should be unauthorized", t, func() {
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

// TestOptionsBypassesToken verifies that OPTIONS requests bypass token validation to allow CORS preflight.
func TestOptionsBypassesToken(t *testing.T) {
	r, _ := http.NewRequest("OPTIONS", "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("OPTIONS request should bypass token validation", t, func() {
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

// TestPublicPostWithoutToken confirms that public POST endpoints can be accessed without a token and return a validation error instead.
func TestPublicPostWithoutToken(t *testing.T) {
	r, _ := http.NewRequest("POST", "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("POST /restaurante/v1/clientes without token should not return unauthorized", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}
