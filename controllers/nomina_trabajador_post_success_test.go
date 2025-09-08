package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestNominaTrabajadorPostSuccess(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) {
				incs := dst.(*[]models.Incidencia)
				*incs = append(*incs,
					models.Incidencia{MONTO: 100, RESTA: true},
					models.Incidencia{MONTO: 50, RESTA: false},
				)
				return 2, nil
			}},
			"*models.Trabajador": {one: func(dst interface{}, _ ...string) error {
				tr := dst.(*models.Trabajador)
				tr.SUELDO = 1000
				return nil
			}},
			"*models.Nomina": {one: func(dst interface{}, _ ...string) error {
				n := dst.(*models.Nomina)
				n.FECHA = time.Now()
				n.PK_ID_NOMINA = 1
				return nil
			}},
			"*models.NominaTrabajador": {exist: false},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
