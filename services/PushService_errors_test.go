package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
)

func strp(s string) *string { return &s }
func i64p(i int64) *int64   { return &i }

func TestPushService_TiposDeError(t *testing.T) {
	assert.Equal(t, "v", (&ValidationError{Msg: "v"}).Error())
	assert.Equal(t, "n", (&NotFoundError{Msg: "n"}).Error())
	assert.Equal(t, "c", (&ConflictError{Msg: "c"}).Error())

	assert.True(t, IsValidationError(fmt.Errorf("x: %w", validationf("a %d", 1))))
	assert.False(t, IsValidationError(errors.New("otro")))
	assert.True(t, IsNotFoundError(fmt.Errorf("x: %w", notFoundf("a"))))
	assert.False(t, IsNotFoundError(errors.New("otro")))
	assert.True(t, IsConflictError(fmt.Errorf("x: %w", &ConflictError{Msg: "a"})))
	assert.False(t, IsConflictError(errors.New("otro")))
}

func TestClasificarDB(t *testing.T) {
	assert.True(t, IsConflictError(clasificarDB(errors.New(`pq: duplicate key value violates unique constraint "x"`), "m")))
	assert.True(t, IsNotFoundError(clasificarDB(errors.New(`pq: violates foreign key constraint "fk"`), "m")))
	err := clasificarDB(errors.New("boom"), "mensaje")
	assert.False(t, IsConflictError(err))
	assert.False(t, IsNotFoundError(err))
	assert.Contains(t, err.Error(), "mensaje: boom")
}

func TestPushService_ValidarRegistroDispositivo_Errores(t *testing.T) {
	s := &PushService{}
	web := func() *models.RegistrarDispositivoRequest {
		return &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaWeb, Endpoint: strp("e"), P256dh: strp("p"), Auth: strp("a"), PkDocumentoCliente: i64p(1)}
	}
	and := func() *models.RegistrarDispositivoRequest {
		return &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoTrabajador: i64p(1)}
	}

	r := web()
	r.PkDocumentoTrabajador = i64p(2)
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "exactamente uno")

	r = web()
	r.PkDocumentoCliente = i64p(0)
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "positivo")

	r = and()
	r.PkDocumentoTrabajador = i64p(-3)
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "positivo")

	r = web()
	r.Auth = strp("  ")
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "WEB se requieren")

	r = web()
	r.FcmToken = strp("t")
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "no se debe especificar fcmToken")

	r = and()
	r.FcmToken = nil
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "se requiere fcmToken")

	r = and()
	r.Endpoint = strp("e")
	assert.ErrorContains(t, s.ValidarRegistroDispositivo(r), "no se deben especificar")

	r = and()
	r.Plataforma = "PALOMA"
	err := s.ValidarRegistroDispositivo(r)
	assert.ErrorContains(t, err, "plataforma no válida")
	assert.True(t, IsValidationError(err))
}

func TestPushService_RegistrarDispositivo_PropietarioYBusqueda(t *testing.T) {
	ctx := context.Background()
	cli := &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoCliente: i64p(1)}
	trab := &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoTrabajador: i64p(2)}

	// cliente inexistente
	s := NewPushService(&mockPushOrmer{readFn: func(interface{}, ...string) error { return orm.ErrNoRows }})
	_, _, err := s.RegistrarDispositivo(ctx, cli)
	assert.True(t, IsNotFoundError(err))
	assert.ErrorContains(t, err, "cliente no encontrado")

	// error leyendo cliente
	s = NewPushService(&mockPushOrmer{readFn: func(interface{}, ...string) error { return errors.New("db") }})
	_, _, err = s.RegistrarDispositivo(ctx, cli)
	assert.False(t, IsNotFoundError(err))
	assert.ErrorContains(t, err, "error al buscar cliente")

	// trabajador inexistente
	s = NewPushService(&mockPushOrmer{readFn: func(interface{}, ...string) error { return orm.ErrNoRows }})
	_, _, err = s.RegistrarDispositivo(ctx, trab)
	assert.True(t, IsNotFoundError(err))
	assert.ErrorContains(t, err, "trabajador no encontrado")

	// error leyendo trabajador
	s = NewPushService(&mockPushOrmer{readFn: func(interface{}, ...string) error { return errors.New("db") }})
	_, _, err = s.RegistrarDispositivo(ctx, trab)
	assert.ErrorContains(t, err, "error al buscar trabajador")

	// error buscando por token
	s = NewPushService(&mockPushOrmer{queryTableFn: func(string) orm.QuerySeter {
		return &mockPushQuerySeter{filterFn: func(string, ...interface{}) orm.QuerySeter {
			return &mockPushQuerySeter{oneFn: func(interface{}, ...string) error { return errors.New("db") }}
		}}
	}})
	_, _, err = s.RegistrarDispositivo(ctx, cli)
	assert.ErrorContains(t, err, "error al buscar dispositivo")
}

func seteroExistente(d *models.PushDispositivo) func(string) orm.QuerySeter {
	return func(string) orm.QuerySeter {
		return &mockPushQuerySeter{filterFn: func(string, ...interface{}) orm.QuerySeter {
			return &mockPushQuerySeter{oneFn: func(c interface{}, _ ...string) error {
				*(c.(*models.PushDispositivo)) = *d
				return nil
			}}
		}}
	}
}

func TestPushService_RegistrarDispositivo_ReasignaYConflictos(t *testing.T) {
	ctx := context.Background()

	// existente: pasa de trabajador a cliente y se reactiva; created=false
	existente := &models.PushDispositivo{PkIdPushDispositivo: 9, Plataforma: models.PlataformaAndroid, PkDocumentoTrabajador: &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 5}, SubscribedTopics: `{"a"}`}
	s := NewPushService(&mockPushOrmer{queryTableFn: seteroExistente(existente)})
	d, created, err := s.RegistrarDispositivo(ctx, &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoCliente: i64p(7)})
	assert.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, int64(7), d.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
	assert.Nil(t, d.PkDocumentoTrabajador)
	assert.True(t, d.Enabled)

	// nuevo: created=true
	s = NewPushService(&mockPushOrmer{})
	d, created, err = s.RegistrarDispositivo(ctx, &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoCliente: i64p(7)})
	assert.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, int64(7), d.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)

	dup := errors.New(`pq: duplicate key value violates unique constraint "uq_push_fcm"`)
	// carrera en insert -> 409
	s = NewPushService(&mockPushOrmer{insertFn: func(interface{}) (int64, error) { return 0, dup }})
	_, _, err = s.RegistrarDispositivo(ctx, &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoTrabajador: i64p(7)})
	assert.True(t, IsConflictError(err))

	// update -> 409
	s = NewPushService(&mockPushOrmer{queryTableFn: seteroExistente(&models.PushDispositivo{}), updateFn: func(interface{}, ...string) (int64, error) { return 0, dup }})
	_, _, err = s.RegistrarDispositivo(ctx, &models.RegistrarDispositivoRequest{Plataforma: models.PlataformaAndroid, FcmToken: strp("t"), PkDocumentoTrabajador: i64p(7)})
	assert.True(t, IsConflictError(err))
}

func TestPushService_ActualizarDispositivo(t *testing.T) {
	ctx := context.Background()
	base := func() *models.PushDispositivo {
		return &models.PushDispositivo{PkIdPushDispositivo: 3, Plataforma: models.PlataformaWeb, Enabled: true, Locale: strp("es"), TimeZone: strp("UTC"), AppVersion: strp("1"), UserAgent: strp("ua"), SubscribedTopics: `{"x"}`}
	}
	leer := func(d *models.PushDispositivo) func(interface{}, ...string) error {
		return func(m interface{}, _ ...string) error { *(m.(*models.PushDispositivo)) = *d; return nil }
	}

	// merge: solo enabled cambia, el resto se conserva
	var guardado *models.PushDispositivo
	s := NewPushService(&mockPushOrmer{readFn: leer(base()), updateFn: func(m interface{}, cols ...string) (int64, error) {
		guardado = m.(*models.PushDispositivo)
		assert.Contains(t, cols, "Enabled")
		return 1, nil
	}})
	d, err := s.ActualizarDispositivo(ctx, 3, []byte(`{"enabled":false}`))
	assert.NoError(t, err)
	assert.False(t, d.Enabled)
	assert.Equal(t, "es", *d.Locale)
	assert.Equal(t, []string{"x"}, d.SubscribedTopicsArray)
	assert.Same(t, guardado, d)

	// todos los campos, incluyendo topics vacío y null en anulables
	s = NewPushService(&mockPushOrmer{readFn: leer(base())})
	d, err = s.ActualizarDispositivo(ctx, 3, []byte(`{"locale":null,"timeZone":"America/Bogota","appVersion":null,"userAgent":null,"subscribedTopics":[]}`))
	assert.NoError(t, err)
	assert.Nil(t, d.Locale)
	assert.Equal(t, "America/Bogota", *d.TimeZone)
	assert.Nil(t, d.AppVersion)
	assert.Nil(t, d.UserAgent)
	assert.Empty(t, d.SubscribedTopicsArray)
	assert.True(t, d.Enabled)

	// null en campo no anulable / JSON inválido -> validación
	for _, body := range []string{`{"enabled":null}`, `{"subscribedTopics":null}`, `no-json`, `[]`} {
		_, err = s.ActualizarDispositivo(ctx, 3, []byte(body))
		assert.True(t, IsValidationError(err), body)
	}
	// tipo incorrecto
	_, err = s.ActualizarDispositivo(ctx, 3, []byte(`{"enabled":"si"}`))
	assert.True(t, IsValidationError(err))

	// no existe
	s = NewPushService(&mockPushOrmer{})
	_, err = s.ActualizarDispositivo(ctx, 3, []byte(`{}`))
	assert.True(t, IsNotFoundError(err))

	// error de lectura y de actualización
	s = NewPushService(&mockPushOrmer{readFn: func(interface{}, ...string) error { return errors.New("db") }})
	_, err = s.ActualizarDispositivo(ctx, 3, []byte(`{}`))
	assert.ErrorContains(t, err, "error al buscar dispositivo")
	s = NewPushService(&mockPushOrmer{readFn: leer(base()), updateFn: func(interface{}, ...string) (int64, error) { return 0, errors.New("db") }})
	_, err = s.ActualizarDispositivo(ctx, 3, []byte(`{}`))
	assert.ErrorContains(t, err, "error al actualizar dispositivo")
}

func TestPushService_RegistrarEnvio_Validaciones(t *testing.T) {
	ctx := context.Background()
	s := NewPushService(&mockPushOrmer{})
	_, err := s.RegistrarEnvio(ctx, &models.RegistrarEnvioRequest{Proveedor: models.ProveedorFCM})
	assert.True(t, IsValidationError(err))

	// FK violada al insertar -> 404
	s = NewPushService(&mockPushOrmer{
		readFn:   func(interface{}, ...string) error { return nil },
		insertFn: func(interface{}) (int64, error) { return 0, errors.New("violates foreign key constraint") },
	})
	_, err = s.RegistrarEnvio(ctx, &models.RegistrarEnvioRequest{PkIdPushDispositivo: 1, Proveedor: models.ProveedorFCM})
	assert.True(t, IsNotFoundError(err))
}

func TestPushService_ValidarRemitente_ErrorBD(t *testing.T) {
	s := NewPushService(&mockPushOrmer{queryTableFn: func(string) orm.QuerySeter {
		return &mockPushQuerySeter{filterFn: func(string, ...interface{}) orm.QuerySeter {
			return &mockPushQuerySeter{oneFn: func(interface{}, ...string) error { return errors.New("db") }}
		}}
	}})
	err := s.validarRemitente(&models.RemitenteNotificacion{Tipo: models.RemitenteTrabajador, DocumentoTrabajador: i64p(1)})
	assert.ErrorContains(t, err, "error al buscar trabajador")

	err = s.validarRemitente(&models.RemitenteNotificacion{Tipo: "X"})
	assert.True(t, IsValidationError(err))
}

func TestPushService_RegistrarEnvioNotificacion_ErrorInsertSeLoguea(t *testing.T) {
	s := NewPushService(&mockPushOrmer{insertFn: func(interface{}) (int64, error) { return 0, errors.New("db") }})
	assert.NotPanics(t, func() {
		s.registrarEnvioNotificacion(&models.PushDispositivo{Plataforma: models.PlataformaWeb}, &models.ContenidoNotificacion{}, true, nil, nil)
	})
}
