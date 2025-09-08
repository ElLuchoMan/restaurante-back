package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestIncidenciaDeleteSuccess(t *testing.T) {
	orig := MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockExec = orig })

	r := httptest.NewRequest(http.MethodDelete, "/incidencias?id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
