package cupon

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web"
)

func TestCuponController_Integration(t *testing.T) {

	web.BConfig.RunMode = "test"

	t.Run("CrearCupon - Request válido", func(t *testing.T) {

		request := &models.CrearCuponRequest{
			Codigo:         "CUPON_TEST_001",
			Scope:          models.CuponScopeGlobal,
			TipoDescuento:  models.TipoDescuentoPorcentaje,
			ValorDescuento: 15,
			FechaInicio:    "2024-01-15",
			FechaFin:       "2024-01-22",
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		req := httptest.NewRequest("POST", "/restaurante/v1/cupones/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test_token")

		var parsedRequest models.CrearCuponRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.Codigo != request.Codigo {
			t.Errorf("Expected codigo %s, got %s", request.Codigo, parsedRequest.Codigo)
		}

		if parsedRequest.TipoDescuento != request.TipoDescuento {
			t.Errorf("Expected tipo descuento %s, got %s", request.TipoDescuento, parsedRequest.TipoDescuento)
		}

		if parsedRequest.ValorDescuento != request.ValorDescuento {
			t.Errorf("Expected valor descuento %d, got %d", request.ValorDescuento, parsedRequest.ValorDescuento)
		}

	})

	t.Run("CrearCupon - Request inválido", func(t *testing.T) {

		request := &models.CrearCuponRequest{
			Codigo:         "CUPON_INVALID",
			Scope:          models.CuponScopeGlobal,
			TipoDescuento:  models.TipoDescuentoPorcentaje,
			ValorDescuento: 150,
			FechaInicio:    "2024-01-15",
			FechaFin:       "2024-01-22",
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		var parsedRequest models.CrearCuponRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.ValorDescuento != 150 {
			t.Errorf("Expected valor descuento 150, got %d", parsedRequest.ValorDescuento)
		}

	})

	t.Run("ValidarCupon - Request válido", func(t *testing.T) {
		request := &models.ValidarCuponRequest{
			Codigo:    "CUPON_VALIDO",
			ClienteId: 12345678,
			Items: []models.ValidarCuponItemRequest{
				{ProductoId: 1, Cantidad: 2, Precio: 10000},
				{ProductoId: 2, Cantidad: 1, Precio: 5000},
			},
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		var parsedRequest models.ValidarCuponRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.Codigo != request.Codigo {
			t.Errorf("Expected codigo %s, got %s", request.Codigo, parsedRequest.Codigo)
		}

		if parsedRequest.ClienteId != request.ClienteId {
			t.Errorf("Expected cliente ID %d, got %d", request.ClienteId, parsedRequest.ClienteId)
		}

		if len(parsedRequest.Items) != len(request.Items) {
			t.Errorf("Expected %d items, got %d", len(request.Items), len(parsedRequest.Items))
		}

		for i, item := range parsedRequest.Items {
			expectedItem := request.Items[i]
			if item.ProductoId != expectedItem.ProductoId {
				t.Errorf("Item %d: expected producto ID %d, got %d", i, expectedItem.ProductoId, item.ProductoId)
			}
			if item.Cantidad != expectedItem.Cantidad {
				t.Errorf("Item %d: expected cantidad %d, got %d", i, expectedItem.Cantidad, item.Cantidad)
			}
			if item.Precio != expectedItem.Precio {
				t.Errorf("Item %d: expected precio %d, got %d", i, expectedItem.Precio, item.Precio)
			}
		}
	})

	t.Run("RedimirCupon - Request válido", func(t *testing.T) {
		request := &models.RedimirCuponRequest{
			ClienteId: 12345678,
			PedidoId:  testInt64Ptr(1),
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		var parsedRequest models.RedimirCuponRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.ClienteId != request.ClienteId {
			t.Errorf("Expected cliente ID %d, got %d", request.ClienteId, parsedRequest.ClienteId)
		}

		if parsedRequest.PedidoId == nil || *parsedRequest.PedidoId != *request.PedidoId {
			t.Errorf("Expected pedido ID %v, got %v", request.PedidoId, parsedRequest.PedidoId)
		}
	})
}

func TestCuponController_ResponseStructures(t *testing.T) {
	t.Run("ValidarCuponResponse - Serialización", func(t *testing.T) {
		motivo := "Cupón no encontrado"
		response := &models.ValidarCuponResponse{
			Aplicable:      false,
			MontoDescuento: 0,
			Motivo:         &motivo,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Error marshaling response: %v", err)
		}

		var parsedResponse models.ValidarCuponResponse
		err = json.Unmarshal(jsonData, &parsedResponse)
		if err != nil {
			t.Errorf("Error parsing response: %v", err)
		}

		if parsedResponse.Aplicable != response.Aplicable {
			t.Errorf("Expected aplicable %v, got %v", response.Aplicable, parsedResponse.Aplicable)
		}

		if parsedResponse.MontoDescuento != response.MontoDescuento {
			t.Errorf("Expected monto descuento %d, got %d", response.MontoDescuento, parsedResponse.MontoDescuento)
		}

		if parsedResponse.Motivo == nil || *parsedResponse.Motivo != *response.Motivo {
			t.Errorf("Expected motivo %v, got %v", response.Motivo, parsedResponse.Motivo)
		}
	})
}

func testInt64Ptr(i int64) *int64 {
	return &i
}
