package incidencia

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidOrmChangeDoc struct{}

func (fakeIncidOrmChangeDoc) QueryTable(interface{}) orm.QuerySeter        { return nil }
func (fakeIncidOrmChangeDoc) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmChangeDoc) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmChangeDoc) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmChangeDoc) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaPut_ChangeDocumentoTrabajador(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmChangeDoc{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	body := `{"documentoTrabajador":777,"motivo":"x"}`
	r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
