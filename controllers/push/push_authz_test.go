package push

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	docCliente    = int64(1001)
	docTrabajador = int64(2002)
	docAjenoPush  = int64(9999)
)

// dispositivoDe devuelve un dispositivo con credenciales de un solo dueño.
func dispositivoDe(cliente, trabajador *int64) *models.PushDispositivo {
	d := &models.PushDispositivo{
		PkIdPushDispositivo: 7,
		Plataforma:          models.PlataformaWeb,
		Endpoint:            ptr("https://push.example/abc"),
		P256dh:              ptr("p256-secreto"),
		Auth:                ptr("auth-secreto"),
		FcmToken:            ptr("fcm-secreto"),
		Enabled:             true,
	}
	if cliente != nil {
		d.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *cliente}
	}
	if trabajador != nil {
		d.PkDocumentoTrabajador = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: *trabajador}
	}
	return d
}

func usarDispositivo(t *testing.T, d *models.PushDispositivo) {
	t.Helper()
	useOrm(t, &fakeOrm{qs: &fakeQS{}, read: func(md interface{}) { *(md.(*models.PushDispositivo)) = *d }})
}

func sinCredencialesEn(t *testing.T, cuerpo string) {
	t.Helper()
	for _, secreto := range []string{"secreto", "https://push.example", `"endpoint"`, `"p256dh"`, `"auth"`, `"fcmToken"`} {
		assert.NotContains(t, cuerpo, secreto)
	}
}

func svcOK() *fakeSvc {
	return &fakeSvc{
		actualiza: func(int64, []byte) (*models.PushDispositivo, error) { return dispositivoDe(ptr(docCliente), nil), nil },
		visto:     func(int64) error { return nil },
		topics:    func(int64, []string) error { return nil },
	}
}

// endpoints por dispositivo: cada uno recibe el Authorization y devuelve el status.
var porDispositivo = []struct {
	nombre string
	llamar func(t *testing.T, auth string) *httptest.ResponseRecorder
}{
	{"GetById", func(t *testing.T, auth string) *httptest.ResponseRecorder {
		rec, _, _ := doAs(t, auth, http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
		return rec
	}},
	{"Put", func(t *testing.T, auth string) *httptest.ResponseRecorder {
		rec, _, _ := doAs(t, auth, http.MethodPut, "/push/dispositivos?id=7", `{"enabled":false}`, (*PushController).Put)
		return rec
	}},
	{"Delete", func(t *testing.T, auth string) *httptest.ResponseRecorder {
		rec, _, _ := doAs(t, auth, http.MethodDelete, "/push/dispositivos?id=7", "", (*PushController).Delete)
		return rec
	}},
	{"ActualizarUltimaVista", func(t *testing.T, auth string) *httptest.ResponseRecorder {
		rec, _, _ := doAs(t, auth, http.MethodPatch, "/push/dispositivos/visto?id=7", "", (*PushController).ActualizarUltimaVista)
		return rec
	}},
	{"ActualizarTopics", func(t *testing.T, auth string) *httptest.ResponseRecorder {
		rec, _, _ := doAs(t, auth, http.MethodPatch, "/push/dispositivos/topics?id=7", `{"subscribedTopics":["a"]}`, (*PushController).ActualizarTopics)
		return rec
	}},
}

func TestDispositivo_SoloDuenoOAdministrador(t *testing.T) {
	casos := []struct {
		nombre string
		device *models.PushDispositivo
		auth   string
		want   int
	}{
		{"sin token", dispositivoDe(ptr(docCliente), nil), "", http.StatusUnauthorized},
		{"cliente dueño", dispositivoDe(ptr(docCliente), nil), tokenDe(rolCliente, docCliente), http.StatusOK},
		{"cliente ajeno", dispositivoDe(ptr(docCliente), nil), tokenDe(rolCliente, docAjenoPush), http.StatusNotFound},
		{"trabajador dueño", dispositivoDe(nil, ptr(docTrabajador)), tokenDe(rolMesero, docTrabajador), http.StatusOK},
		{"trabajador ajeno", dispositivoDe(nil, ptr(docTrabajador)), tokenDe(rolMesero, docAjenoPush), http.StatusNotFound},
		{"administrador dueño", dispositivoDe(nil, ptr(docTrabajador)), tokenDe(rolAdmin, docTrabajador), http.StatusOK},
		{"administrador de otro", dispositivoDe(ptr(docCliente), nil), tokenDe(rolAdmin, docAjenoPush), http.StatusOK},
		// el mismo número de documento en el otro rol no da acceso
		{"cliente con el documento del trabajador dueño", dispositivoDe(nil, ptr(docTrabajador)), tokenDe(rolCliente, docTrabajador), http.StatusNotFound},
		{"trabajador con el documento del cliente dueño", dispositivoDe(ptr(docCliente), nil), tokenDe(rolMesero, docCliente), http.StatusNotFound},
	}
	for _, ep := range porDispositivo {
		for _, c := range casos {
			t.Run(ep.nombre+"/"+c.nombre, func(t *testing.T) {
				usarDispositivo(t, c.device)
				useSvc(t, svcOK())
				assert.Equal(t, c.want, ep.llamar(t, c.auth).Code)
			})
		}
	}
}

func TestDispositivo_AjenoEsIgualQueInexistente(t *testing.T) {
	for _, ep := range porDispositivo {
		t.Run(ep.nombre, func(t *testing.T) {
			svc := svcOK()
			svc.actualiza = func(int64, []byte) (*models.PushDispositivo, error) {
				t.Fatal("no debe llegar al servicio")
				return nil, nil
			}
			svc.visto = func(int64) error { t.Fatal("no debe llegar al servicio"); return nil }
			svc.topics = func(int64, []string) error { t.Fatal("no debe llegar al servicio"); return nil }
			useSvc(t, svc)
			auth := tokenDe(rolCliente, docAjenoPush)

			useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: orm.ErrNoRows})
			inexistente := ep.llamar(t, auth)
			usarDispositivo(t, dispositivoDe(ptr(docCliente), nil))
			ajeno := ep.llamar(t, auth)

			assert.Equal(t, http.StatusNotFound, inexistente.Code)
			assert.JSONEq(t, inexistente.Body.String(), ajeno.Body.String(), "inexistente y ajeno deben ser indistinguibles")
			assert.Contains(t, ajeno.Body.String(), "Dispositivo no encontrado")
		})
	}
}

func TestDispositivo_ErrorDeLecturaYParametros(t *testing.T) {
	// id inválido sigue siendo 400 para un usuario autenticado, 401 sin token
	rec, _, _ := doAs(t, tokenDe(rolCliente, docCliente), http.MethodGet, "/push/dispositivos/search", "", (*PushController).GetById)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	rec, _, _ = doAs(t, "", http.MethodGet, "/push/dispositivos/search", "", (*PushController).GetById)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	useOrm(t, &fakeOrm{qs: &fakeQS{}, readErr: errors.New("db")})
	rec, _, _ = doAs(t, tokenDe(rolCliente, docCliente), http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDispositivo_RespuestasSinCredenciales(t *testing.T) {
	dueno := tokenDe(rolCliente, docCliente)
	usarDispositivo(t, dispositivoDe(ptr(docCliente), nil))
	svc := svcOK()
	useSvc(t, svc)

	rec, _, _ := doAs(t, dueno, http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	require.Equal(t, http.StatusOK, rec.Code)
	sinCredencialesEn(t, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"documentoCliente":1001`)

	// el administrador tampoco recibe las credenciales
	rec, _, _ = doAs(t, tokenDe(rolAdmin, 1), http.MethodGet, "/push/dispositivos/search?id=7", "", (*PushController).GetById)
	require.Equal(t, http.StatusOK, rec.Code)
	sinCredencialesEn(t, rec.Body.String())

	svc.actualiza = func(int64, []byte) (*models.PushDispositivo, error) { return dispositivoDe(ptr(docCliente), nil), nil }
	rec, _, _ = doAs(t, dueno, http.MethodPut, "/push/dispositivos?id=7", `{"enabled":true}`, (*PushController).Put)
	require.Equal(t, http.StatusOK, rec.Code)
	sinCredencialesEn(t, rec.Body.String())

	// el listado del Administrador
	useOrm(t, &fakeOrm{qs: &fakeQS{count: 1, fill: func(res interface{}) {
		*(res.(*[]*models.PushDispositivo)) = []*models.PushDispositivo{dispositivoDe(ptr(docCliente), nil)}
	}}})
	rec, _, _ = doAs(t, tokenDe(rolAdmin, 1), http.MethodGet, "/push/dispositivos", "", (*PushController).GetAll)
	require.Equal(t, http.StatusOK, rec.Code)
	sinCredencialesEn(t, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"documentoCliente":1001`)
}

func TestSinCredencialesNoMutaElOriginal(t *testing.T) {
	d := dispositivoDe(ptr(docCliente), nil)
	copia := sinCredenciales(d)
	assert.Nil(t, copia.Endpoint)
	assert.Nil(t, copia.P256dh)
	assert.Nil(t, copia.Auth)
	assert.Nil(t, copia.FcmToken)
	assert.NotNil(t, d.FcmToken, "el original conserva sus credenciales")
	assert.Equal(t, d.PkIdPushDispositivo, copia.PkIdPushDispositivo)
}

// ---------- rutas solo Administrador ----------

func TestRutasSoloAdministrador(t *testing.T) {
	useOrm(t, &fakeOrm{qs: &fakeQS{}})
	useSvc(t, &fakeSvc{})
	rutas := []struct {
		nombre string
		llamar func(auth string) int
	}{
		{"GetAll", func(auth string) int {
			rec, _, _ := doAs(t, auth, http.MethodGet, "/push/dispositivos", "", (*PushController).GetAll)
			return rec.Code
		}},
		{"ListarEnvios", func(auth string) int {
			rec, _, _ := doAs(t, auth, http.MethodGet, "/push/envios", "", (*PushController).ListarEnvios)
			return rec.Code
		}},
		{"RegistrarEnvio", func(auth string) int {
			rec, _, _ := doAs(t, auth, http.MethodPost, "/push/envios", `{}`, (*PushController).RegistrarEnvio)
			return rec.Code
		}},
		{"EnviarNotificacion", func(auth string) int {
			rec, _, _ := doAs(t, auth, http.MethodPost, "/push/enviar", notif("t", "m"), (*PushController).EnviarNotificacion)
			return rec.Code
		}},
	}
	for _, r := range rutas {
		t.Run(r.nombre, func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, r.llamar(""))
			assert.Equal(t, http.StatusForbidden, r.llamar(tokenDe(rolCliente, docCliente)))
			assert.Equal(t, http.StatusForbidden, r.llamar(tokenDe(rolMesero, docTrabajador)))
		})
	}
}

// ---------- POST /push/dispositivos: el dueño sale del token ----------

func registrarCon(t *testing.T, auth, body string) (int, *models.RegistrarDispositivoRequest) {
	t.Helper()
	var got *models.RegistrarDispositivoRequest
	useSvc(t, &fakeSvc{registrar: func(r *models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
		got = r
		return dispositivoDe(r.PkDocumentoCliente, r.PkDocumentoTrabajador), true, nil
	}})
	rec, _, _ := doAs(t, auth, http.MethodPost, "/push/dispositivos", body, (*PushController).Post)
	sinCredencialesEn(t, rec.Body.String())
	return rec.Code, got
}

func TestPost_DuenoSaleDelToken(t *testing.T) {
	const web = `"plataforma":"WEB","endpoint":"https://push.example/abc","p256dh":"k","auth":"a"`

	code, got := registrarCon(t, "", `{`+web+`}`)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.Nil(t, got)

	// Cliente: sin dueño en el cuerpo o con el propio
	for _, extra := range []string{``, `,"documentoCliente":1001`} {
		code, got = registrarCon(t, tokenDe(rolCliente, docCliente), `{`+web+extra+`}`)
		require.Equal(t, http.StatusCreated, code, extra)
		require.NotNil(t, got.PkDocumentoCliente)
		assert.Equal(t, docCliente, *got.PkDocumentoCliente)
		assert.Nil(t, got.PkDocumentoTrabajador)
	}

	// Trabajador y Administrador: siempre como documentoTrabajador
	for _, rol := range []string{rolMesero, rolAdmin} {
		for _, extra := range []string{``, `,"documentoTrabajador":2002`} {
			code, got = registrarCon(t, tokenDe(rol, docTrabajador), `{`+web+extra+`}`)
			require.Equal(t, http.StatusCreated, code, rol+extra)
			require.NotNil(t, got.PkDocumentoTrabajador)
			assert.Equal(t, docTrabajador, *got.PkDocumentoTrabajador)
			assert.Nil(t, got.PkDocumentoCliente)
		}
	}
}

func TestPost_DuenoDistintoEs403(t *testing.T) {
	const web = `"plataforma":"WEB","endpoint":"https://push.example/abc","p256dh":"k","auth":"a"`
	casos := []struct{ nombre, auth, extra string }{
		{"cliente a nombre de otro cliente", tokenDe(rolCliente, docCliente), `,"documentoCliente":9999`},
		{"cliente como trabajador", tokenDe(rolCliente, docCliente), `,"documentoTrabajador":1001`},
		{"trabajador a nombre de un cliente", tokenDe(rolMesero, docTrabajador), `,"documentoCliente":2002`},
		{"trabajador a nombre de otro trabajador", tokenDe(rolMesero, docTrabajador), `,"documentoTrabajador":9999`},
		{"administrador a nombre de otro trabajador", tokenDe(rolAdmin, docTrabajador), `,"documentoTrabajador":9999`},
		{"administrador a nombre de un cliente", tokenDe(rolAdmin, docTrabajador), `,"documentoCliente":1001`},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			code, got := registrarCon(t, c.auth, `{`+web+c.extra+`}`)
			assert.Equal(t, http.StatusForbidden, code)
			assert.Nil(t, got, "no debe llegar al servicio")
		})
	}
}

func TestPost_TokenSinDocumentoEs403(t *testing.T) {
	for _, rol := range []string{rolCliente, rolMesero} {
		code, got := registrarCon(t, tokenDe(rol, 0), `{"plataforma":"ANDROID","fcmToken":"x"}`)
		assert.Equal(t, http.StatusForbidden, code, rol)
		assert.Nil(t, got)
	}
}

func TestPost_JSONInvalidoConToken(t *testing.T) {
	code, _ := registrarCon(t, tokenDe(rolCliente, docCliente), `{`)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRespuestaPostNoContieneCredenciales(t *testing.T) {
	useSvc(t, &fakeSvc{registrar: func(*models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
		return dispositivoDe(ptr(docCliente), nil), false, nil
	}})
	rec, _, _ := doAs(t, tokenDe(rolCliente, docCliente), http.MethodPost, "/push/dispositivos", `{"plataforma":"ANDROID","fcmToken":"x"}`, (*PushController).Post)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, strings.Contains(rec.Body.String(), "fcm-secreto"))
	sinCredencialesEn(t, rec.Body.String())
}
