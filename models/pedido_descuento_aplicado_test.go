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

	pda.serializeDetalle()

	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "tipo")
	assert.Contains(t, pda.Detalle, "porcentaje")
}

func TestPedidoDescuentoAplicado_SerializeDetalle_Empty(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		DetalleObj: json.RawMessage{},
	}

	pda.serializeDetalle()

	assert.Equal(t, "", pda.Detalle)
}

func TestPedidoDescuentoAplicado_DeserializeDetalle(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: `{"tipo":"fijo","valor":10000}`,
	}

	pda.deserializeDetalle()

	assert.NotNil(t, pda.DetalleObj)
	assert.Contains(t, string(pda.DetalleObj), "tipo")
	assert.Contains(t, string(pda.DetalleObj), "fijo")
}

func TestPedidoDescuentoAplicado_DeserializeDetalle_Empty(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: "",
	}

	pda.deserializeDetalle()

	assert.Nil(t, pda.DetalleObj)
}

func TestPedidoDescuentoAplicado_BeforeInsert(t *testing.T) {
	detalleJSON := json.RawMessage(`{"descuento":"cupon"}`)
	pda := &PedidoDescuentoAplicado{
		DetalleObj: detalleJSON,
	}

	pda.BeforeInsert()

	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "cupon")
}

func TestPedidoDescuentoAplicado_BeforeUpdate(t *testing.T) {
	detalleJSON := json.RawMessage(`{"descuento":"oferta"}`)
	pda := &PedidoDescuentoAplicado{
		DetalleObj: detalleJSON,
	}

	pda.BeforeUpdate()

	assert.NotEmpty(t, pda.Detalle)
	assert.Contains(t, pda.Detalle, "oferta")
}

func TestPedidoDescuentoAplicado_AfterLoad(t *testing.T) {
	pda := &PedidoDescuentoAplicado{
		Detalle: `{"tipo":"combinado","valor":15000}`,
	}

	pda.AfterLoad()

	assert.NotNil(t, pda.DetalleObj)
	assert.Contains(t, string(pda.DetalleObj), "combinado")
}

func TestPedidoDescuentoAplicado_JSONSerialization(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}
	detalleJSON := json.RawMessage(`{"info":"test"}`)
	created := time.Date(2024, time.August, 17, 12, 30, 45, 0, loc)
	pda := &PedidoDescuentoAplicado{
		PkIdPedidoDescuento: 1,
		MontoDescuento:      8000,
		DetalleObj:          detalleJSON,
		CreatedAt:           created,
	}

	jsonData, err := json.Marshal(pda)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["pedidoDescuentoId"])
	assert.Equal(t, float64(8000), jsonMap["montoDescuento"])
	assert.Equal(t, "17-08-2024 12:30:45", jsonMap["createdAt"])

	detalle, exists := jsonMap["detalle"]
	assert.True(t, exists)
	assert.NotNil(t, detalle)
}
