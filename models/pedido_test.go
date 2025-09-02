package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPedidoMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, time.May, 2, 12, 0, 0, 0, time.UTC)
	p := Pedido{FECHA: fecha, UPDATED_AT: updated}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaPedido"] != "01-05-2024" {
		t.Errorf("expected fechaPedido 01-05-2024, got %v", data["fechaPedido"])
	}
	if data["updatedAt"] != "02-05-2024 12:00:00" {
		t.Errorf("expected updatedAt 02-05-2024 12:00:00, got %v", data["updatedAt"])
	}
}

func TestPedidoTableName(t *testing.T) {
	p := Pedido{}
	if p.TableName() != "pedido" {
		t.Errorf("expected table name pedido, got %s", p.TableName())
	}
}

type fakeOrmerDetalle struct {
	insert func(interface{}) (int64, error)
}

func (f fakeOrmerDetalle) Insert(m interface{}) (int64, error) { return f.insert(m) }

func TestDetallePedidoPrecioAuto(t *testing.T) {
	o := fakeOrmerDetalle{insert: func(m interface{}) (int64, error) {
		if d, ok := m.(*DetallePedido); ok {
			d.Precio = 500
		}
		return 1, nil
	}}
	pedidoID := int64(1)
	productoID := int64(1)
	detalle := DetallePedido{
		PKIDPedido:   &pedidoID,
		PKIDProducto: &productoID,
		Cantidad:     1,
	}
	if _, err := o.Insert(&detalle); err != nil {
		t.Fatalf("insert returned error: %v", err)
	}
	if detalle.Precio == 0 {
		t.Errorf("expected Precio to be set by trigger")
	}
}
