package cupon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para CuponController - alcanzar 95%+ cobertura

func TestCuponPost_FechaFinBeforeFechaInicio(t *testing.T) {
	controller, recorder, ctx := setupTest()

	body := map[string]interface{}{
		"codigo":          "TEST2025",
		"descripcion":     "Test",
		"scope":           "producto",
		"tipoDescuento":   "porcentaje",
		"valorDescuento":  float64(10),
		"fechaInicio":     "2025-01-20",
		"fechaFin":        "2025-01-10", // Antes de fechaInicio
		"restauranteId":   float64(1),
		"limiteTotalUsos": float64(100),
	}
	bodyBytes, _ := json.Marshal(body)
	ctx.Input.RequestBody = bodyBytes

	controller.Post()

	// Debería retornar error de validación
	assert.NotEqual(t, http.StatusCreated, recorder.Code)
}

func TestCuponPut_NoChanges(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Body vacío - sin cambios
	bodyBytes := []byte("{}")
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam("id", "1")

	// Mock
	existingCupon := &models.Cupon{
		PK_ID_CUPON: 1,
		Codigo:      "EXISTING",
		Activo:      true,
	}

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = *existingCupon
	})

	// Sin cambios, no debería llamar Update
	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(int64(0), nil)

	controller.Put()

	// Debería aceptar sin cambios
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestCuponDelete_NotFound(t *testing.T) {
	controller, recorder, ctx := setupTest()
	ctx.Input.SetParam("id", "999")

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(999)).Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(orm.ErrNoRows)

	controller.Delete()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestCuponListarRedenciones_EmptyResult(t *testing.T) {
	controller, recorder, ctx := setupTest()
	ctx.Input.SetParam("cupon_id", "1")

	// Mock cupon existe
	existingCupon := &models.Cupon{
		PK_ID_CUPON: 1,
		Codigo:      "TEST",
		Activo:      true,
	}

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = *existingCupon
	})

	// Mock redenciones vacías
	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("Count").Return(int64(0), nil)
	mockQS.On("OrderBy", []string{"-fecha_redencion"}).Return(mockQS)
	mockQS.On("Limit", 10).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Return(int64(0), nil)

	controller.ListarRedenciones()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestCuponGetAll_WithMultipleFilters(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Filtros múltiples
	ctx.Input.SetParam("activo", "true")
	ctx.Input.SetParam("scope", "producto")
	ctx.Input.SetParam("limit", "20")
	ctx.Input.SetParam("page", "2")

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "activo", true).Return(mockQS)
	mockQS.On("Filter", "scope", "producto").Return(mockQS)
	mockQS.On("Count").Return(int64(50), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(20)).Return(mockQS) // page 2 = offset 20
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Return(int64(20), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestCuponPost_LimiteTotalUsosZero(t *testing.T) {
	controller, recorder, ctx := setupTest()

	fechaInicio := time.Now().AddDate(0, 0, -1)
	fechaFin := time.Now().AddDate(0, 0, 7)

	body := map[string]interface{}{
		"codigo":          "UNLIMITED",
		"descripcion":     "Cupón sin límite",
		"scope":           "pedido",
		"tipoDescuento":   "monto_fijo",
		"valorDescuento":  float64(1000),
		"fechaInicio":     fechaInicio.Format("2006-01-02"),
		"fechaFin":        fechaFin.Format("2006-01-02"),
		"restauranteId":   float64(1),
		"limiteTotalUsos": float64(0), // Sin límite
	}
	bodyBytes, _ := json.Marshal(body)
	ctx.Input.RequestBody = bodyBytes

	mockOrmer.On("Insert", mock.AnythingOfType("*models.Cupon")).Return(int64(123), nil)

	controller.Post()

	// Debería aceptar 0 como "sin límite"
	if recorder.Code == http.StatusCreated || recorder.Code == http.StatusOK {
		// Success
	} else {
		t.Errorf("Expected success, got %d", recorder.Code)
	}
}

func TestCuponPut_UpdateOnlyFechas(t *testing.T) {
	controller, recorder, ctx := setupTest()

	existingCupon := &models.Cupon{
		PK_ID_CUPON:    1,
		Codigo:         "EXISTING",
		ValorDescuento: 15,
		Activo:         true,
	}

	nuevaFechaInicio := time.Now().AddDate(0, 0, 5)
	nuevaFechaFin := time.Now().AddDate(0, 0, 15)

	body := map[string]interface{}{
		"fechaInicio": nuevaFechaInicio.Format("2006-01-02"),
		"fechaFin":    nuevaFechaFin.Format("2006-01-02"),
	}
	bodyBytes, _ := json.Marshal(body)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam("id", "1")

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = *existingCupon
	})

	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string{"FechaInicio", "FechaFin"}).Return(int64(1), nil)

	controller.Put()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestCuponGetById_WithLoadRelated(t *testing.T) {
	controller, recorder, ctx := setupTest()
	ctx.Input.SetParam("id", "1")

	cupon := &models.Cupon{
		PK_ID_CUPON:     1,
		Codigo:          "TEST",
		Activo:          true,
		PKIDRestaurante: &models.Restaurante{PK_ID_RESTAURANTE: 1, NOMBRE: "Test Restaurant"},
	}

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("RelatedSel", "PKIDRestaurante").Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = *cupon
	})

	controller.GetById()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestCuponPost_ValorDescuentoNegativo(t *testing.T) {
	controller, recorder, ctx := setupTest()

	body := map[string]interface{}{
		"codigo":          "INVALID",
		"descripcion":     "Test",
		"scope":           "producto",
		"tipoDescuento":   "porcentaje",
		"valorDescuento":  float64(-10), // Negativo
		"fechaInicio":     "2025-01-20",
		"fechaFin":        "2025-01-30",
		"restauranteId":   float64(1),
		"limiteTotalUsos": float64(100),
	}
	bodyBytes, _ := json.Marshal(body)
	ctx.Input.RequestBody = bodyBytes

	controller.Post()

	// Debería retornar error de validación
	assert.NotEqual(t, http.StatusCreated, recorder.Code)
}

func TestCuponListarRedenciones_WithPagination(t *testing.T) {
	controller, recorder, ctx := setupTest()
	ctx.Input.SetParam("cupon_id", "1")
	ctx.Input.SetParam("page", "2")
	ctx.Input.SetParam("limit", "5")

	existingCupon := &models.Cupon{
		PK_ID_CUPON: 1,
		Codigo:      "TEST",
		Activo:      true,
	}

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("One", &models.Cupon{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = *existingCupon
	})

	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", int64(1)).Return(mockQS)
	mockQS.On("Count").Return(int64(15), nil)
	mockQS.On("OrderBy", []string{"-fecha_redencion"}).Return(mockQS)
	mockQS.On("Limit", 5).Return(mockQS)
	mockQS.On("Offset", int64(5)).Return(mockQS) // page 2 = offset 5
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Return(int64(5), nil)

	controller.ListarRedenciones()

	assert.Equal(t, http.StatusOK, recorder.Code)
}
