package test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"

	"testing"

	"github.com/beego/beego/v2/core/logs"

	_ "restaurante/routers"

	beego "github.com/beego/beego/v2/server/web"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.BConfig.RunMode = "test"
	beego.TestBeegoInit(apppath)
}

func TestBeego(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	_ = logs.SetLogger(logs.AdapterConsole)

	Convey("Subject: Test Root Endpoint\n", t, func() {
		Convey("Status Code Should Be 404", func() {
			So(w.Code, ShouldEqual, 404)
		})
		Convey("The Result Should Not Be Empty", func() {
			So(w.Body.Len(), ShouldBeGreaterThan, 0)
		})
	})

	_ = _coverSelf()
}
