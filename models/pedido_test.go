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
	if p.TableName() != "PEDIDO" {
		t.Errorf("expected table name PEDIDO, got %s", p.TableName())
	}
}
