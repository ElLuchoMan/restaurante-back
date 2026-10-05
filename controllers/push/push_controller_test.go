package push

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dispositivoConPassword() *models.PushDispositivo {
	return &models.PushDispositivo{
		PkIdPushDispositivo:   7,
		Plataforma:            models.PlataformaAndroid,
		FcmToken:              ptr("tok"),
		Enabled:               true,
		PkDocumentoCliente:    &models.Cliente{PK_DOCUMENTO_CLIENTE: 1001, PASSWORD: "secreto"},
		PkDocumentoTrabajador: &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 2002, PASSWORD: "secreto"},
		CreatedAt:             time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
	}
}

// ---------- GetAll ----------

func TestGetAll_OKVacioDevuelveListaVacia(t *testing.T) {
	qs := &fakeQS{count: 0}
	useOrm(t, &fakeOrm{qs: qs})
	rec, resp, _ := do(t, http.MethodGet, "/push/dispositivos", "", (*PushController).GetAll)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"data":[]`)
	assert.NotContains(t, rec.Body.String(), "null")
	assert.Equal(t, 20, qs.limit)
	assert.Equal(t, int64(0), qs.offset)
	assert.Equal(t, "-created_at", qs.order)
	assert.NotEmpty(t, resp.Message)
}

func TestGetAll_OKConFiltrosYPaginacion(t *testing.T) {
	qs := &fakeQS{count: 45, fill: func(res interface{}) {
		*(res.(*[]*models.PushDispositivo)) = []*models.PushDispositivo{dispositivoConPassword()}
	}}
	useOrm(t, &fakeOrm{qs: qs})
	rec, _, raw := do(t, http.MethodGet, "/push/dispositivos?cliente_id=1&trabajador_id=2&plataforma=IOS&limit=500&offset=40", "", (*PushController).GetAll)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"pk_documento_cliente", "pk_documento_trabajador", "plataforma"}, qs.filters)
	assert.Equal(t, 100, qs.limit)
	assert.Equal(t, int64(40), qs.offset)

	body := rec.Body.String()
	assert.NotContains(t, body, "secreto")
	assert.NotContains(t, strings.ToLower(body), "password")
	data := raw["data"].(map[string]interface{})
	assert.EqualValues(t, 45, data["total"])
	assert.EqualValues(t, 1, data["totalPages"])
	assert.EqualValues(t, 1, data["page"])
	item := data["data"].([]interface{})[0].(map[string]interface{})
	assert.EqualValues(t, 1001, item["documentoCliente"])
	assert.EqualValues(t, 2002, item["documentoTrabajador"])
	assert.Equal(t, []interface{}{}, item["subscribedTopics"])
}

func TestGetAll_ParametrosInvalidos(t *testing.T) {
	useOrm(t, &fakeOrm{qs: &fakeQS{}})
	for _, q := range []string{
		"cliente_id=abc", "cliente_id=0", "trabajador_id=-1", "trabajador_id=x",
		"plataforma=PALOMA", "limit=0", "limit=x", "offset=-1", "offset=x",
	} {
		rec, _, _ := do(t, http.MethodGet, "/push/dispositivos?"+q, "", (*PushController).GetAll)
		assert.Equal(t, http.StatusBadRequest, rec.Code, q)
	}
}

func TestGetAll_ErroresDeBD(t *testing.T) {
	useOrm(t, &fakeOrm{qs: &fakeQS{countErr: errors.New("db")}})
	rec, _, _ := do(t, http.MethodGet, "/push/dispositivos", "", (*PushController).GetAll)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	useOrm(t, &fakeOrm{qs: &fakeQS{allErr: errors.New("db")}})
	rec, _, _ = do(t, http.MethodGet, "/push/dispositivos", "", (*PushController).GetAll)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- Post ----------

func TestPost_Estados(t *testing.T) {
	casos := []struct {
		nombre  string
		body    string
		created bool
		err     error
		status  int
	}{
		{"creado", `{"plataforma":"ANDROID"}`, true, nil, http.StatusCreated},
		{"actualizado", `{"plataforma":"ANDROID"}`, false, nil, http.StatusOK},
		{"validacion", `{"plataforma":"X"}`, false, &services.ValidationError{Msg: "mal"}, http.StatusBadRequest},
		{"no encontrado", `{}`, false, &services.NotFoundError{Msg: "cliente no encontrado"}, http.StatusNotFound},
		{"conflicto", `{}`, false, &services.ConflictError{Msg: "dup"}, http.StatusConflict},
		{"interno", `{}`, false, errors.New("db"), http.StatusInternalServerError},
		{"json invalido", `{`, false, nil, http.StatusBadRequest},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			useSvc(t, &fakeSvc{registrar: func(*models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
				if c.err != nil {
					return nil, false, c.err
				}
				return dispositivoConPassword(), c.created, nil
			}})
			rec, _, _ := do(t, http.MethodPost, "/push/dispositivos", c.body, (*PushController).Post)
			assert.Equal(t, c.status, rec.Code)
			assert.NotContains(t, rec.Body.String(), "secreto")
		})
	}
}

func TestPost_ErrorInternoNoFiltraDetalleEnMensaje(t *testing.T) {
	useSvc(t, &fakeSvc{registrar: func(*models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
		return nil, false, errors.New("db")
	}})
	_, resp, _ := do(t, http.MethodPost, "/push/dispositivos", `{}`, (*PushController).Post)
	assert.Equal(t, "Error al registrar dispositivo", resp.Message)
}

// ---------- GetById ----------

func TestGetById(t *testing.T) {
	useOrm(t, &fakeOrm{qs: &fakeQS{}, read: func(md interface{}) {
		d := md.(*models.PushDispositivo)
		*d = *dispositivoConPassword()
		d.SubscribedTopics = `{"a","b"}`
	}})
	rec, _, raw := do(t, http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, strings.ToLower(rec.Body.String()), "password")
	data := raw["data"].(map[string]interface{})
	assert.EqualValues(t, 7, data["pushDispositivoId"])
	assert.Equal(t, []interface{}{"a", "b"}, data["subscribedTopics"])

	for _, q := range []string{"", "?id=0", "?id=x"} {
		rec, _, _ = do(t, http.MethodGet, "/push/dispositivos/search"+q, "", (*PushController).GetById)
		assert.Equal(t, http.StatusBadRequest, rec.Code, q)
	}

	useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: orm.ErrNoRows})
	rec, _, _ = do(t, http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: errors.New("db")})
	rec, _, _ = do(t, http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- Put ----------

func TestPut(t *testing.T) {
	var gotID int64
	var gotBody string
	svc := &fakeSvc{actualiza: func(id int64, b []byte) (*models.PushDispositivo, error) {
		gotID, gotBody = id, string(b)
		return dispositivoConPassword(), nil
	}}
	useSvc(t, svc)
	rec, _, _ := do(t, http.MethodPut, "/push/dispositivos?id=7", `{"enabled":false}`, (*PushController).Put)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(7), gotID)
	assert.JSONEq(t, `{"enabled":false}`, gotBody)
	assert.NotContains(t, rec.Body.String(), "secreto")

	rec, _, _ = do(t, http.MethodPut, "/push/dispositivos", `{}`, (*PushController).Put)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	casos := map[error]int{
		&services.ValidationError{Msg: "v"}: http.StatusBadRequest,
		&services.NotFoundError{Msg: "n"}:   http.StatusNotFound,
		errors.New("db"):                    http.StatusInternalServerError,
	}
	for err, status := range casos {
		svc.actualiza = func(int64, []byte) (*models.PushDispositivo, error) { return nil, err }
		rec, _, _ = do(t, http.MethodPut, "/push/dispositivos?id=7", `{}`, (*PushController).Put)
		assert.Equal(t, status, rec.Code, err.Error())
	}
}

// ---------- Delete ----------

func TestDelete(t *testing.T) {
	o := &fakeOrm{qs: &fakeQS{}}
	useOrm(t, o)
	rec, _, _ := do(t, http.MethodDelete, "/push/dispositivos?id=7", "", (*PushController).Delete)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, o.deleted)

	for _, q := range []string{"", "?id=0"} {
		rec, _, _ = do(t, http.MethodDelete, "/push/dispositivos"+q, "", (*PushController).Delete)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: orm.ErrNoRows})
	rec, _, _ = do(t, http.MethodDelete, "/push/dispositivos?id=7", "", (*PushController).Delete)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: errors.New("db")})
	rec, _, _ = do(t, http.MethodDelete, "/push/dispositivos?id=7", "", (*PushController).Delete)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	useOrm(t, &fakeOrm{qs: &fakeQS{}, deleteErr: errors.New("db")})
	rec, _, _ = do(t, http.MethodDelete, "/push/dispositivos?id=7", "", (*PushController).Delete)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- ActualizarUltimaVista ----------

func TestActualizarUltimaVista(t *testing.T) {
	svc := &fakeSvc{visto: func(int64) error { return nil }}
	useSvc(t, svc)
	rec, _, _ := do(t, http.MethodPatch, "/push/dispositivos/visto?id=7", "", (*PushController).ActualizarUltimaVista)
	assert.Equal(t, http.StatusOK, rec.Code)

	rec, _, _ = do(t, http.MethodPatch, "/push/dispositivos/visto", "", (*PushController).ActualizarUltimaVista)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.visto = func(int64) error { return &services.NotFoundError{Msg: "dispositivo no encontrado"} }
	rec, resp, _ := do(t, http.MethodPatch, "/push/dispositivos/visto?id=7", "", (*PushController).ActualizarUltimaVista)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	svc.visto = func(int64) error { return errors.New("db") }
	rec, _, _ = do(t, http.MethodPatch, "/push/dispositivos/visto?id=7", "", (*PushController).ActualizarUltimaVista)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- ActualizarTopics ----------

func TestActualizarTopics(t *testing.T) {
	var got []string
	svc := &fakeSvc{topics: func(_ int64, tp []string) error { got = tp; return nil }}
	useSvc(t, svc)
	do200 := func(body string) {
		rec, _, _ := do(t, http.MethodPatch, "/push/dispositivos/topics?id=7", body, (*PushController).ActualizarTopics)
		assert.Equal(t, http.StatusOK, rec.Code, body)
	}
	do200(`{"subscribedTopics":["a","b"]}`)
	assert.Equal(t, []string{"a", "b"}, got)
	do200(`{"subscribedTopics":[]}`)
	assert.Empty(t, got)

	for _, body := range []string{`{`, `{}`, `{"subscribedTopics":null}`} {
		rec, _, _ := do(t, http.MethodPatch, "/push/dispositivos/topics?id=7", body, (*PushController).ActualizarTopics)
		assert.Equal(t, http.StatusBadRequest, rec.Code, body)
	}
	rec, _, _ := do(t, http.MethodPatch, "/push/dispositivos/topics", `{"subscribedTopics":[]}`, (*PushController).ActualizarTopics)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.topics = func(int64, []string) error { return &services.NotFoundError{Msg: "n"} }
	rec, _, _ = do(t, http.MethodPatch, "/push/dispositivos/topics?id=7", `{"subscribedTopics":[]}`, (*PushController).ActualizarTopics)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	svc.topics = func(int64, []string) error { return errors.New("db") }
	rec, _, _ = do(t, http.MethodPatch, "/push/dispositivos/topics?id=7", `{"subscribedTopics":[]}`, (*PushController).ActualizarTopics)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- EnviarNotificacion ----------

func notif(titulo, mensaje string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"remitente":     map[string]string{"tipo": "SISTEMA"},
		"destinatarios": map[string]string{"tipo": "TODOS"},
		"notificacion":  map[string]string{"titulo": titulo, "mensaje": mensaje},
	})
	return string(b)
}

func TestEnviarNotificacion(t *testing.T) {
	svc := &fakeSvc{enviar: func(*models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) {
		return &models.EnviarNotificacionResponse{DetalleEnvios: []models.DetalleEnvioNotificacion{}}, nil
	}}
	useSvc(t, svc)
	post := func(body string) int {
		rec, _, _ := do(t, http.MethodPost, "/push/enviar", body, (*PushController).EnviarNotificacion)
		return rec.Code
	}
	assert.Equal(t, http.StatusOK, post(notif("hola", "mundo")))
	assert.Equal(t, http.StatusOK, post(notif(strings.Repeat("ñ", 100), strings.Repeat("ñ", 500))))
	assert.Equal(t, http.StatusBadRequest, post(`{`))
	assert.Equal(t, http.StatusBadRequest, post(notif("", "m")))
	assert.Equal(t, http.StatusBadRequest, post(notif(strings.Repeat("a", 101), "m")))
	assert.Equal(t, http.StatusBadRequest, post(notif("t", "")))
	assert.Equal(t, http.StatusBadRequest, post(notif("t", strings.Repeat("a", 501))))

	casos := map[error]int{
		&services.ValidationError{Msg: "v"}: http.StatusBadRequest,
		&services.NotFoundError{Msg: "n"}:   http.StatusNotFound,
		errors.New("db"):                    http.StatusInternalServerError,
	}
	for err, status := range casos {
		svc.enviar = func(*models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) { return nil, err }
		assert.Equal(t, status, post(notif("t", "m")), err.Error())
	}
}

// ---------- ListarEnvios ----------

func TestListarEnvios(t *testing.T) {
	qs := &fakeQS{count: 1, fill: func(res interface{}) {
		*(res.(*[]*models.PushEnvio)) = []*models.PushEnvio{{
			PkIdPushEnvio:       3,
			PkIdPushDispositivo: dispositivoConPassword(),
			Proveedor:           models.ProveedorFCM,
			Data:                `{"a":1}`,
			SentAt:              time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		}}
	}}
	useOrm(t, &fakeOrm{qs: qs})
	rec, _, raw := do(t, http.MethodGet, "/push/envios?dispositivo_id=7&fecha_desde=2026-10-01&fecha_hasta=2026-10-05&limit=10&offset=10", "", (*PushController).ListarEnvios)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"pk_id_push_dispositivo", "sent_at__gte", "sent_at__lte"}, qs.filters)
	assert.Equal(t, 10, qs.limit)
	assert.NotContains(t, strings.ToLower(rec.Body.String()), "password")
	data := raw["data"].(map[string]interface{})
	assert.EqualValues(t, 2, data["page"])
	item := data["data"].([]interface{})[0].(map[string]interface{})
	assert.EqualValues(t, 7, item["pushDispositivoId"], "debe ser el id numérico, no un objeto")
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, item["data"])

	useOrm(t, &fakeOrm{qs: &fakeQS{}})
	rec, _, _ = do(t, http.MethodGet, "/push/envios", "", (*PushController).ListarEnvios)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"data":[]`)

	for _, q := range []string{"dispositivo_id=x", "fecha_desde=ayer", "fecha_hasta=2026-13-45", "limit=0", "offset=-5"} {
		rec, _, _ = do(t, http.MethodGet, "/push/envios?"+q, "", (*PushController).ListarEnvios)
		assert.Equal(t, http.StatusBadRequest, rec.Code, q)
	}

	useOrm(t, &fakeOrm{qs: &fakeQS{countErr: errors.New("db")}})
	rec, _, _ = do(t, http.MethodGet, "/push/envios", "", (*PushController).ListarEnvios)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	useOrm(t, &fakeOrm{qs: &fakeQS{allErr: errors.New("db")}})
	rec, _, _ = do(t, http.MethodGet, "/push/envios", "", (*PushController).ListarEnvios)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ---------- RegistrarEnvio ----------

func TestRegistrarEnvio(t *testing.T) {
	svc := &fakeSvc{envio: func(*models.RegistrarEnvioRequest) (*models.PushEnvio, error) {
		return &models.PushEnvio{PkIdPushEnvio: 1, PkIdPushDispositivo: dispositivoConPassword(), Proveedor: models.ProveedorFCM}, nil
	}}
	useSvc(t, svc)
	rec, _, raw := do(t, http.MethodPost, "/push/envios", `{"pushDispositivoId":7,"proveedor":"FCM","exito":true}`, (*PushController).RegistrarEnvio)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.EqualValues(t, 7, raw["data"].(map[string]interface{})["pushDispositivoId"])
	assert.NotContains(t, rec.Body.String(), "secreto")

	rec, _, _ = do(t, http.MethodPost, "/push/envios", `{`, (*PushController).RegistrarEnvio)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	casos := map[error]int{
		&services.ValidationError{Msg: "v"}: http.StatusBadRequest,
		&services.NotFoundError{Msg: "n"}:   http.StatusNotFound,
		errors.New("db"):                    http.StatusInternalServerError,
	}
	for err, status := range casos {
		svc.envio = func(*models.RegistrarEnvioRequest) (*models.PushEnvio, error) { return nil, err }
		rec, _, _ = do(t, http.MethodPost, "/push/envios", `{}`, (*PushController).RegistrarEnvio)
		assert.Equal(t, status, rec.Code, err.Error())
	}
}

// ---------- Proveedores por defecto y adaptadores ----------

func TestProveedoresPorDefecto(t *testing.T) {
	_, ok := newPushService(newServiceOrm()).(*services.PushService)
	assert.True(t, ok)
	assert.NotNil(t, newServiceOrm())
}
