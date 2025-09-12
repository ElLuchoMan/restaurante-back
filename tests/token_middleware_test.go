package test

import (
	"net/http"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProtectedEndpointWithoutToken(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	r, _ := http.NewRequest("GET", "/restaurante/v1/clientes", nil)
	w := sendRequest(r)

	Convey("GET /restaurante/v1/clientes without token should be unauthorized", t, func() {
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

func TestOptionsBypassesToken(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	r, _ := http.NewRequest("OPTIONS", "/restaurante/v1/clientes", nil)
	w := sendRequest(r)

	Convey("OPTIONS request should bypass token validation", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
		So(w.Code, ShouldEqual, http.StatusMethodNotAllowed)
	})
}

func TestPublicPostWithoutToken(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	rBody := strings.NewReader("{")
	r, _ := http.NewRequest("POST", "/restaurante/v1/clientes", rBody)
	r.Header.Set("Content-Type", "application/json")
	w := sendRequest(r)

	Convey("POST /restaurante/v1/clientes without token should not return unauthorized", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
		So(w.Code, ShouldBeIn, []int{http.StatusBadRequest, http.StatusInternalServerError})
	})
}
