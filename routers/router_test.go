package routers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	os.Setenv("JWT_SECRET", "testsecret")
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

func TestLoginGetRouteNotFound(t *testing.T) {
	r, _ := http.NewRequest("GET", "/restaurante/v1/login", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("Subject: GET /restaurante/v1/login\n", t, func() {
		Convey("Status Code Should Be 404", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestStaticFile(t *testing.T) {
	r, _ := http.NewRequest("GET", "/static/js/reload.min.js", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("Subject: GET static file\n", t, func() {
		Convey("Status Code Should Be 200", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
