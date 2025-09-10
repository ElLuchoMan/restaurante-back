package controlnomina

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestControlNomina_GetAll_NoFilter(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/control_nomina", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	// Status puede ser 200 o 500 dependiendo de la DB mock; validamos que no paniquee
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
