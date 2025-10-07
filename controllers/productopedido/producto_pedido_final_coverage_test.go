package productopedido

import (
	"testing"

	"restaurante/models"
)

// TestComputeDeltas_NullProductoID cubre el caso de PKIDProducto == nil (línea 53)
func TestComputeDeltas_NullProductoID(t *testing.T) {
	// Crear detalles con producto NULL
	actuales := []models.DetallePedido{
		{
			PKIDProducto: nil, // Producto NULL - debería ser ignorado
			Cantidad:     5,
		},
		{
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 100},
			Cantidad:     3,
		},
	}

	nuevos := map[int64]int{
		100: 5, // Aumentar de 3 a 5
		200: 2, // Nuevo producto
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	// Verificar que el producto con ID 100 tiene delta +2
	if deltas[100] != 2 {
		t.Errorf("Expected delta[100] = 2, got %d", deltas[100])
	}

	// Verificar que el producto 200 (nuevo) tiene delta +2
	if deltas[200] != 2 {
		t.Errorf("Expected delta[200] = 2, got %d", deltas[200])
	}

	// Verificar que need solo tiene productos positivos
	if need[100] != 2 {
		t.Errorf("Expected need[100] = 2, got %d", need[100])
	}
	if need[200] != 2 {
		t.Errorf("Expected need[200] = 2, got %d", need[200])
	}

	// El producto NULL no debería aparecer en deltas ni need
	if _, exists := deltas[0]; exists {
		t.Error("Expected producto with ID 0 (from NULL) to not exist in deltas")
	}
}

// TestComputeDeltas_ZeroDelta cubre el caso de delta == 0 (línea 432 else implícito)
func TestComputeDeltas_ZeroDelta(t *testing.T) {
	// Crear detalles actuales
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

	// Nuevos con las mismas cantidades (delta == 0)
	nuevos := map[int64]int{
		100: 5, // Sin cambio
		200: 3, // Sin cambio
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	// Verificar que ambos deltas son 0
	if deltas[100] != 0 {
		t.Errorf("Expected delta[100] = 0, got %d", deltas[100])
	}
	if deltas[200] != 0 {
		t.Errorf("Expected delta[200] = 0, got %d", deltas[200])
	}

	// Verificar que need está vacío (no hay necesidades positivas)
	if len(need) != 0 {
		t.Errorf("Expected need to be empty, got %d items", len(need))
	}
}

// TestComputeDeltas_MixedWithNull cubre múltiples casos incluyendo NULL y deltas variados
func TestComputeDeltas_MixedWithNull(t *testing.T) {
	actuales := []models.DetallePedido{
		{PKIDProducto: nil, Cantidad: 10},                                // NULL - ignorado
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 1}, Cantidad: 5}, // Producto 1: 5
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 2}, Cantidad: 3}, // Producto 2: 3
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 3}, Cantidad: 2}, // Producto 3: 2
	}

	nuevos := map[int64]int{
		1: 5,  // Sin cambio (delta = 0)
		2: 5,  // Aumentar (delta = +2)
		4: 10, // Nuevo (delta = +10)
		// Producto 3 no está en nuevos (delta = -2, se elimina)
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	// Verificaciones
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

	// Need solo debe contener productos con delta > 0
	if len(need) != 2 {
		t.Errorf("Expected need to have 2 items, got %d", len(need))
	}
	if need[2] != 2 {
		t.Errorf("Expected need[2] = 2, got %d", need[2])
	}
	if need[4] != 10 {
		t.Errorf("Expected need[4] = 10, got %d", need[4])
	}

	// Verificar que el producto NULL no afectó los resultados
	if _, exists := deltas[0]; exists {
		t.Error("Expected producto with ID 0 (from NULL) to not exist in deltas")
	}
}

// TestComputeDeltas_AllNull cubre el caso extremo de todos los productos NULL
func TestComputeDeltas_AllNull(t *testing.T) {
	actuales := []models.DetallePedido{
		{PKIDProducto: nil, Cantidad: 5},
		{PKIDProducto: nil, Cantidad: 10},
	}

	nuevos := map[int64]int{
		100: 5,
	}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	// Solo el producto 100 debería aparecer (nuevo)
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
