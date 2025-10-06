package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPedidoDescuentoAplicado_TableName(t *testing.T) {
	pda := &PedidoDescuentoAplicado{}
	assert.Equal(t, "pedido_descuento_aplicado", pda.TableName())
}

func TestPedidoDescuentoAplicado_SerializeDetalle(t *testing.T) {
	detalleJSON := json.RawMessage(`{"tipo":"porcentaje","valor":20}`)
	pda := &PedidoDescuentoAplicado{
		PkIdPedidoDescuento: 1,
		MontoDescuento:      5000,
		DetalleObj:          detalleJSON,
		CreatedAt:           time.Now(),
	}

	// Serializar detalle
	pda.serializeDetalle()

	// Verificar que Detalle contiene el JSON como string
	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "tipo")
	assert.Contains(t, pda.Detalle, "porcentaje")
}

func TestPedidoDescuentoAplicado_SerializeDetalle_Empty(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		DetalleObj: json.RawMessage{},
	}

	pda.serializeDetalle()

	// Detalle debe estar vacío
	assert.Equal(t, "", pda.Detalle)
}

func TestPedidoDescuentoAplicado_DeserializeDetalle(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: `{"tipo":"fijo","valor":10000}`,
	}

	// Deserializar detalle
	pda.deserializeDetalle()

	// Verificar que DetalleObj contiene el JSON
	assert.NotNil(t, pda.DetalleObj)
	assert.Contains(t, string(pda.DetalleObj), "tipo")
	assert.Contains(t, string(pda.DetalleObj), "fijo")
}

func TestPedidoDescuentoAplicado_DeserializeDetalle_Empty(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: "",
	}

	pda.deserializeDetalle()

	// DetalleObj debe ser nil
	assert.Nil(t, pda.DetalleObj)
}

func TestPedidoDescuentoAplicado_BeforeInsert(t *testing.T) {
	detalleJSON := json.RawMessage(`{"descuento":"cupon"}`)
	pda := &PedidoDescuentoAplicado{
		DetalleObj: detalleJSON,
	}

	// Llamar hook
	pda.BeforeInsert()

	// Verificar que se serializó
	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "cupon")
}

func TestPedidoDescuentoAplicado_BeforeUpdate(t *testing.T) {
	detalleJSON := json.RawMessage(`{"descuento":"oferta"}`)
	pda := &PedidoDescuentoAplicado{
		DetalleObj: detalleJSON,
	}

	// Llamar hook
	pda.BeforeUpdate()

	// Verificar que se serializó
	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "oferta")
}

func TestPedidoDescuentoAplicado_AfterLoad(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: `{"tipo":"combinado","valor":15000}`,
	}

	// Llamar hook
	pda.AfterLoad()

	// Verificar que se deserializó
	assert.NotNil(t, pda.DetalleObj)
	assert.Contains(t, string(pda.DetalleObj), "combinado")
}

func TestPedidoDescuentoAplicado_JSONSerialization(t *testing.T) {
	detalleJSON := json.RawMessage(`{"info":"test"}`)
	pda := &PedidoDescuentoAplicado{
		PkIdPedidoDescuento: 1,
		MontoDescuento:      8000,
		DetalleObj:          detalleJSON,
		CreatedAt:           time.Now(),
	}

	// Serializar
	jsonData, err := json.Marshal(pda)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar campos en JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["pedidoDescuentoId"])
	assert.Equal(t, float64(8000), jsonMap["montoDescuento"])

	// Verificar que detalle está como objeto (no string)
	detalle, exists := jsonMap["detalle"]
	assert.True(t, exists)
	assert.NotNil(t, detalle)
}
