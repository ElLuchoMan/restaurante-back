package push

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
)

// fakeQS es un pushQuerySeter en memoria que registra los filtros recibidos.
type fakeQS struct {
	filters  []string
	limit    int
	offset   int64
	order    string
	count    int64
	countErr error
	allErr   error
	fill     func(res interface{})
}

func (q *fakeQS) All(res interface{}, _ ...string) (int64, error) {
	if q.allErr != nil {
		return 0, q.allErr
	}
	if q.fill != nil {
		q.fill(res)
	}
	return 0, nil
}
func (q *fakeQS) Filter(expr string, _ ...interface{}) pushQuerySeter {
	q.filters = append(q.filters, expr)
	return q
}
func (q *fakeQS) OrderBy(e ...string) pushQuerySeter { q.order = strings.Join(e, ","); return q }
func (q *fakeQS) Limit(l int) pushQuerySeter         { q.limit = l; return q }
func (q *fakeQS) Offset(o int64) pushQuerySeter      { q.offset = o; return q }
func (q *fakeQS) Count() (int64, error)              { return q.count, q.countErr }
func (q *fakeQS) One(interface{}) error              { return nil }

// fakeOrm es un pushOrmer en memoria.
type fakeOrm struct {
	qs        *fakeQS
	readErr   error
	read      func(md interface{})
	deleteErr error
	deleted   bool
}

func (o *fakeOrm) QueryTable(interface{}) pushQuerySeter { return o.qs }
func (o *fakeOrm) Insert(interface{}) (int64, error)     { return 1, nil }
func (o *fakeOrm) Update(interface{}, ...string) (int64, error) {
	return 1, nil
}
func (o *fakeOrm) Read(md interface{}, _ ...string) error {
	if o.readErr != nil {
		return o.readErr
	}
	if o.read != nil {
		o.read(md)
	}
	return nil
}
func (o *fakeOrm) Delete(interface{}, ...string) (int64, error) {
	o.deleted = o.deleteErr == nil
	return 1, o.deleteErr
}

// useOrm sustituye pushOrmNew durante el test.
func useOrm(t *testing.T, o *fakeOrm) {
	t.Helper()
	prev := pushOrmNew
	pushOrmNew = func() pushOrmer { return o }
	t.Cleanup(func() { pushOrmNew = prev })
}

// fakeSvc implementa services.PushServiceInterface con funciones configurables.
type fakeSvc struct {
	registrar func(*models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error)
	visto     func(int64) error
	actualiza func(int64, []byte) (*models.PushDispositivo, error)
	topics    func(int64, []string) error
	envio     func(*models.RegistrarEnvioRequest) (*models.PushEnvio, error)
	enviar    func(*models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error)
}

func (s *fakeSvc) RegistrarDispositivo(_ stdcontext.Context, r *models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
	return s.registrar(r)
}
func (s *fakeSvc) ActualizarUltimaVista(_ stdcontext.Context, id int64) error { return s.visto(id) }
func (s *fakeSvc) ActualizarDispositivo(_ stdcontext.Context, id int64, b []byte) (*models.PushDispositivo, error) {
	return s.actualiza(id, b)
}
func (s *fakeSvc) ActualizarTopicsDispositivo(_ stdcontext.Context, id int64, t []string) error {
	return s.topics(id, t)
}
func (s *fakeSvc) RegistrarEnvio(_ stdcontext.Context, r *models.RegistrarEnvioRequest) (*models.PushEnvio, error) {
	return s.envio(r)
}
func (s *fakeSvc) EnviarNotificacion(r *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) {
	return s.enviar(r)
}
func (s *fakeSvc) ValidarRegistroDispositivo(*models.RegistrarDispositivoRequest) error { return nil }

// useSvc sustituye la fábrica de servicios durante el test.
func useSvc(t *testing.T, s *fakeSvc) {
	t.Helper()
	prevSvc, prevOrm := newPushService, newServiceOrm
	newPushService = func(orm.Ormer) services.PushServiceInterface { return s }
	newServiceOrm = func() orm.Ormer { return nil }
	t.Cleanup(func() { newPushService, newServiceOrm = prevSvc, prevOrm })
}

// do ejecuta el handler sobre una petición y devuelve la respuesta decodificada.
func do(t *testing.T, method, target, body string, h func(*PushController)) (*httptest.ResponseRecorder, models.ApiResponse, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	ctx := beecontext.NewContext()
	ctx.Reset(rec, req)
	ctx.Input.RequestBody = []byte(body)
	c := &PushController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	h(c)

	var resp models.ApiResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("respuesta no es JSON: %v (%s)", err, rec.Body.String())
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &raw)
	if resp.Code != rec.Code {
		t.Fatalf("code del cuerpo (%d) != status HTTP (%d)", resp.Code, rec.Code)
	}
	return rec, resp, raw
}

func ptr[T any](v T) *T { return &v }
