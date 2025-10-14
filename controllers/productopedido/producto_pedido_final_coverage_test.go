package productopedido

import (
	"testing"

	"restaurante/models"
)

func TestComputeDeltas_NullProductoID(t *testing.T) {

	actuales := []models.DetallePedido{
		{
			PKIDProducto: nil,
			Cantidad:     5,
		},
		{
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 100},
			Cantidad:     3,
		},
	}

	nuevos := map[int64]int{
		100: 5,
		200: 2,
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	if deltas[100] != 2 {
		t.Errorf("Expected delta[100] = 2, got %d", deltas[100])
	}

	if deltas[200] != 2 {
		t.Errorf("Expected delta[200] = 2, got %d", deltas[200])
	}

	if need[100] != 2 {
		t.Errorf("Expected need[100] = 2, got %d", need[100])
	}
	if need[200] != 2 {
		t.Errorf("Expected need[200] = 2, got %d", need[200])
	}

	if _, exists := deltas[0]; exists {
		t.Error("Expected producto with ID 0 (from NULL) to not exist in deltas")
	}
}

func TestComputeDeltas_ZeroDelta(t *testing.T) {

	actuales := []models.DetallePedido{
		{
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 100},
			Cantidad:     5,
		},
		{
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 200},
			Cantidad:     3,
		},
	}

	nuevos := map[int64]int{
		100: 5,
		200: 3,
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	if deltas[100] != 0 {
		t.Errorf("Expected delta[100] = 0, got %d", deltas[100])
	}
	if deltas[200] != 0 {
		t.Errorf("Expected delta[200] = 0, got %d", deltas[200])
	}

	if len(need) != 0 {
		t.Errorf("Expected need to be empty, got %d items", len(need))
	}
}

func TestComputeDeltas_MixedWithNull(t *testing.T) {
	actuales := []models.DetallePedido{
		{PKIDProducto: nil, Cantidad: 10},
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 1}, Cantidad: 5},
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 2}, Cantidad: 3},
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 3}, Cantidad: 2},
	}

	nuevos := map[int64]int{
		1: 5,
		2: 5,
		4: 10,
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	if deltas[1] != 0 {
		t.Errorf("Expected delta[1] = 0, got %d", deltas[1])
	}
	if deltas[2] != 2 {
		t.Errorf("Expected delta[2] = 2, got %d", deltas[2])
	}
	if deltas[3] != -2 {
		t.Errorf("Expected delta[3] = -2, got %d", deltas[3])
	}
	if deltas[4] != 10 {
		t.Errorf("Expected delta[4] = 10, got %d", deltas[4])
	}

	if len(need) != 2 {
		t.Errorf("Expected need to have 2 items, got %d", len(need))
	}
	if need[2] != 2 {
		t.Errorf("Expected need[2] = 2, got %d", need[2])
	}
	if need[4] != 10 {
		t.Errorf("Expected need[4] = 10, got %d", need[4])
	}

	if _, exists := deltas[0]; exists {
		t.Error("Expected producto with ID 0 (from NULL) to not exist in deltas")
	}
}

func TestComputeDeltas_AllNull(t *testing.T) {
	actuales := []models.DetallePedido{
		{PKIDProducto: nil, Cantidad: 5},
		{PKIDProducto: nil, Cantidad: 10},
	}

	nuevos := map[int64]int{
		100: 5,
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	if len(deltas) != 1 {
		t.Errorf("Expected 1 delta, got %d", len(deltas))
	}
	if deltas[100] != 5 {
		t.Errorf("Expected delta[100] = 5, got %d", deltas[100])
	}
	if need[100] != 5 {
		t.Errorf("Expected need[100] = 5, got %d", need[100])
	}
}
