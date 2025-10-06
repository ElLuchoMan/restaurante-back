package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCuponRedencion_TableName(t *testing.T) {
	cr := &CuponRedencion{}
	assert.Equal(t, "cupon_redencion", cr.TableName())
}

func TestCuponRedencion_JSONSerialization(t *testing.T) {
	now := time.Now()
	cr := &CuponRedencion{
		PkIdCuponRedencion: 1,
		MontoDescuento:     5000,
		CreatedAt:          now,
	}

	// Serializar
	jsonData, err := json.Marshal(cr)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar campos en JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["cuponRedencionId"])
	assert.Equal(t, float64(5000), jsonMap["montoDescuento"])

	// Deserializar
	var crDeserialized CuponRedencion
	err = json.Unmarshal(jsonData, &crDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, cr.MontoDescuento, crDeserialized.MontoDescuento)
}

func TestCuponRedencion_NullableFields(t *testing.T) {
	cr := &CuponRedencion{
		PkIdCuponRedencion: 1,
		MontoDescuento:     5000,
		CreatedAt:          time.Now(),
	}

	// PkIdPedido debe ser nil por defecto (campo opcional)
	assert.Nil(t, cr.PkIdPedido)

	// Serializar y verificar
	jsonData, err := json.Marshal(cr)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// pedidoId no debe aparecer si es nil (omitempty)
	_, exists := jsonMap["pedidoId"]
	assert.False(t, exists)
}
