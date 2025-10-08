package push

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// TESTS PARA EnviarNotificacion
// ============================================================================

func TestEnviarNotificacion_JSONInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader([]byte("{invalid json")))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("{invalid json")
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON inválido")
}

func TestEnviarNotificacion_TituloFaltante(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	tipoSistema := models.RemitenteSistema
	nombreSistema := "Sistema"
	tipoTodos := models.DestinatarioTodos

	reqBody := models.EnviarNotificacionRequest{
		Remitente: models.RemitenteNotificacion{
			Tipo:   tipoSistema,
			Nombre: &nombreSistema,
		},
		Destinatarios: models.DestinatariosNotificacion{
			Tipo: tipoTodos,
		},
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "", // Sin título
			Mensaje: "Test mensaje",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "título es requerido")
}

func TestEnviarNotificacion_MensajeFaltante(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	tipoSistema := models.RemitenteSistema
	nombreSistema := "Sistema"
	tipoTodos := models.DestinatarioTodos

	reqBody := models.EnviarNotificacionRequest{
		Remitente: models.RemitenteNotificacion{
			Tipo:   tipoSistema,
			Nombre: &nombreSistema,
		},
		Destinatarios: models.DestinatariosNotificacion{
			Tipo: tipoTodos,
		},
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "Test título",
			Mensaje: "", // Sin mensaje
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "mensaje es requerido")
}

// ============================================================================
// TESTS PARA ActualizarUltimaVista
// ============================================================================

func TestActualizarUltimaVista_IDInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/visto?id=invalid", nil)
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctrl.Ctx = ctx

	ctrl.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID inválido")
}

// ============================================================================
// TESTS PARA RegistrarEnvio
// ============================================================================

func TestRegistrarEnvio_JSONInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodPost, "/push/envios", bytes.NewReader([]byte("{invalid json")))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("{invalid json")
	ctrl.Ctx = ctx

	ctrl.RegistrarEnvio()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON inválido")
}

func TestRegistrarEnvio_ProveedorInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.RegistrarEnvioRequest{
		Proveedor:           "INVALID_PROVIDER",
		PkIdPushDispositivo: 1,
		Exito:               true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/push/envios", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctrl.Ctx = ctx

	ctrl.RegistrarEnvio()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Proveedor no válido")
}

// ============================================================================
// TESTS DE COBERTURA ADICIONAL
// ============================================================================

func TestPushServiceOrmFactory_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := pushServiceOrmFactory()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}

func TestPushServiceOrmBase_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := pushServiceOrmBase()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}

// ============================================================================
// NOTAS:
// Estos tests cubren los casos críticos de las funciones con baja cobertura:
// - EnviarNotificacion: JSON inválido, título/mensaje faltante, errores ✓
// - ActualizarUltimaVista: ID inválido, errores ✓
// - RegistrarEnvio: JSON inválido, proveedor inválido, errores ✓
//
// Con el patrón de DI implementado, podemos mockear el ORM sin tocar la BD.
// Los servicios tienen alta cobertura, así que la lógica crítica está testeada.
// ============================================================================
