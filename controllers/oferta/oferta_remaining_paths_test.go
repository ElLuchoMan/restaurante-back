package oferta

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ofertaPayload(extra map[string]interface{}) map[string]interface{} {
	p := map[string]interface{}{
		"titulo":         "Oferta Test",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 25,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
		"diasSemana":     []string{"Lunes"},
		"restauranteId":  1,
	}
	for k, v := range extra {
		p[k] = v
	}
	return p
}

func setBody(ctx *context.Context, method, target string, payload interface{}) {
	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body
}

func respOf(t *testing.T, rec *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var r models.ApiResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	return r
}

func TestOfertaDefaults_ConstructoresReales(t *testing.T) {
	assert.NotNil(t, realOfertOrmNew())
	assert.NotNil(t, realNewOfertaService(nil))
}

func TestOfertaPost_FechaInicioInvalida_Mensaje(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()
	setBody(ctx, "POST", "/ofertas", ofertaPayload(map[string]interface{}{"fechaInicio": "no-es-fecha"}))

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, respOf(t, recorder).Message, "Fecha de inicio inválida")
}

func TestOfertaPost_HorarioValido(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()
	setBody(ctx, "POST", "/ofertas", ofertaPayload(map[string]interface{}{"horaInicio": "08:00", "horaFin": "10:30:15"}))

	var got *models.Oferta
	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.Oferta")).Run(func(args mock.Arguments) {
		got = args.Get(0).(*models.Oferta)
	}).Return(int64(1), nil)

	controller.Post()

	assert.Equal(t, http.StatusCreated, recorder.Code)
	require.NotNil(t, got)
	require.NotNil(t, got.HoraInicio)
	require.NotNil(t, got.HoraFin)
	assert.Equal(t, 8, got.HoraInicio.Hour())
	assert.Equal(t, 10, got.HoraFin.Hour())
	assert.Equal(t, 30, got.HoraFin.Minute())
	assert.Equal(t, 15, got.HoraFin.Second())
}

func TestOfertaPost_ReglaDeNegocioInvalida(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()
	setBody(ctx, "POST", "/ofertas", ofertaPayload(map[string]interface{}{"fechaInicio": "2025-12-31", "fechaFin": "2025-01-01"}))

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	resp := respOf(t, recorder)
	assert.Equal(t, "Error de validación", resp.Message)
	assert.Contains(t, resp.Cause, "fecha de fin")
	mockOfertOrmer.AssertNotCalled(t, "Insert", mock.Anything)
}

func TestOfertaPut_HorarioValidoYErrores(t *testing.T) {
	newPut := func(payload map[string]interface{}) (*OfertaController, *httptest.ResponseRecorder) {
		controller, recorder, ctx := setupOfertaTest()
		setBody(ctx, "PUT", "/ofertas?id=1", payload)
		mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil)
		return controller, recorder
	}

	t.Run("horario valido", func(t *testing.T) {
		controller, recorder := newPut(ofertaPayload(map[string]interface{}{"horaInicio": "08:00", "horaFin": "10:00"}))
		var got *models.Oferta
		mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
			got = args.Get(0).(*models.Oferta)
		}).Return(int64(1), nil)

		controller.Put()

		assert.Equal(t, http.StatusOK, recorder.Code)
		require.NotNil(t, got)
		require.NotNil(t, got.HoraFin)
		assert.Equal(t, 10, got.HoraFin.Hour())
	})

	t.Run("regla de negocio invalida", func(t *testing.T) {
		controller, recorder := newPut(ofertaPayload(map[string]interface{}{"fechaInicio": "2025-12-31", "fechaFin": "2025-01-01"}))

		controller.Put()

		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		assert.Contains(t, respOf(t, recorder).Cause, "fecha de fin")
		mockOfertOrmer.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("error al actualizar", func(t *testing.T) {
		controller, recorder := newPut(ofertaPayload(nil))
		mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(int64(0), errors.New("fallo update"))

		controller.Put()

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		resp := respOf(t, recorder)
		assert.Equal(t, "Error al actualizar oferta", resp.Message)
		assert.Equal(t, "fallo update", resp.Cause)
	})
}

func TestOfertaDelete_ErrorDeLectura(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()
	ctx.Request = httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(errors.New("fallo lectura"))

	controller.Delete()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := respOf(t, recorder)
	assert.Equal(t, "Error interno del servidor", resp.Message)
	assert.Equal(t, "fallo lectura", resp.Cause)
}

func TestObtenerOfertasActivas_ParametrosOpcionalesValidos(t *testing.T) {
	origSvc, origFactory := newOfertaService, ofertaServiceOrmFactory
	t.Cleanup(func() { newOfertaService, ofertaServiceOrmFactory = origSvc, origFactory })

	svc := new(MockOfertaServiceInterface)
	var gotFecha, gotHora *time.Time
	var gotProducto *int64
	svc.On("ObtenerOfertasActivas", mock.Anything, int64(3), mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			gotFecha, _ = args.Get(2).(*time.Time)
			gotHora, _ = args.Get(3).(*time.Time)
			gotProducto, _ = args.Get(4).(*int64)
		}).Return([]*models.OfertaActivaResponse{{OfertaId: 1, Titulo: "O"}}, nil)
	newOfertaService = func(orm.Ormer) services.OfertaServiceInterface { return svc }
	ofertaServiceOrmFactory = func() orm.Ormer { return nil }

	controller, recorder, ctx := setupOfertaTest()
	ctx.Request = httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=3&fecha=2025-03-04&hora=09:15&producto_id=77", nil)
	controller.Ctx = ctx

	controller.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, gotFecha)
	require.NotNil(t, gotHora)
	require.NotNil(t, gotProducto)
	assert.Equal(t, 4, gotFecha.Day())
	assert.Equal(t, 9, gotHora.Hour())
	assert.Equal(t, 15, gotHora.Minute())
	assert.Equal(t, int64(77), *gotProducto)
}

func TestOfertaAsociarProducto_ErroresDeInsert(t *testing.T) {
	cases := []struct {
		name       string
		insertErr  error
		wantStatus int
		wantMsg    string
	}{
		{"duplicado", errors.New("duplicate key value violates unique constraint"), http.StatusConflict, "El producto ya está asociado a esta oferta"},
		{"unique sqlite", errors.New("UNIQUE constraint failed"), http.StatusConflict, "El producto ya está asociado a esta oferta"},
		{"otro error", errors.New("conexion perdida"), http.StatusInternalServerError, "Error al asociar producto"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller, recorder, ctx := setupOfertaTest()
			setBody(ctx, "POST", "/ofertas/productos?id=1", map[string]interface{}{"productoId": 1})
			mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil).Once()
			mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Return(nil).Once()
			mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.OfertaProducto")).Return(int64(0), tc.insertErr)

			controller.AsociarProducto()

			assert.Equal(t, tc.wantStatus, recorder.Code)
			assert.Equal(t, tc.wantMsg, respOf(t, recorder).Message)
		})
	}
}
