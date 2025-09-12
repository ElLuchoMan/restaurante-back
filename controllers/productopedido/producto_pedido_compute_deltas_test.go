package productopedido

import (
	"reflect"
	"testing"

	"restaurante/models"
)

func TestProductoPedidoComputeDeltas_Mixed(t *testing.T) {
	actuales := []models.DetallePedido{
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 1}, Cantidad: 2},
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 2}, Cantidad: 1},
		{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 3}, Cantidad: 4},
	}
	nuevos := map[int64]int{1: 2, 2: 3, 3: 3, 4: 5}

	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)

	wantDeltas := map[int64]int{1: 0, 2: 2, 3: -1, 4: 5}
	wantNeed := map[int64]int{2: 2, 4: 5}

	if !reflect.DeepEqual(deltas, wantDeltas) {
		t.Fatalf("deltas mismatch: got %v, want %v", deltas, wantDeltas)
	}
	if !reflect.DeepEqual(need, wantNeed) {
		t.Fatalf("need mismatch: got %v, want %v", need, wantNeed)
	}
}
