package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetodoPago_TableName(t *testing.T) {
	metodoPago := &MetodoPago{}
	assert.Equal(t, "metodo_pago", metodoPago.TableName())
}

func TestMetodoPago_JSONSerialization(t *testing.T) {
	metodoPago := &MetodoPago{
		PK_ID_METODO_PAGO: 1,
		TIPO:              "TARJETA",
		DETALLE:           "Visa terminada en 1234",
	}

	jsonData, err := json.Marshal(metodoPago)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["metodoPagoId"])
	assert.Equal(t, "TARJETA", jsonMap["tipo"])
	assert.Equal(t, "Visa terminada en 1234", jsonMap["detalle"])

	var metodoPagoDeserialized MetodoPago
	err = json.Unmarshal(jsonData, &metodoPagoDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, metodoPago.TIPO, metodoPagoDeserialized.TIPO)
	assert.Equal(t, metodoPago.DETALLE, metodoPagoDeserialized.DETALLE)
}
