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

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}
	now := time.Date(2024, time.August, 17, 12, 30, 45, 0, loc)
	cr := &CuponRedencion{
		PkIdCuponRedencion: 1,
		MontoDescuento:     5000,
		CreatedAt:          now,
	}

	jsonData, err := json.Marshal(cr)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["cuponRedencionId"])
	assert.Equal(t, float64(5000), jsonMap["montoDescuento"])
	assert.Equal(t, "17-08-2024 12:30:45", jsonMap["createdAt"])

	var crMap map[string]interface{}
	err = json.Unmarshal(jsonData, &crMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(cr.MontoDescuento), crMap["montoDescuento"])
}

func TestCuponRedencion_NullableFields(t *testing.T) {
	cr := &CuponRedencion{
		PkIdCuponRedencion: 1,
		MontoDescuento:     5000,
		CreatedAt:          time.Now(),
	}

	assert.Nil(t, cr.PkIdPedido)

	jsonData, err := json.Marshal(cr)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	_, exists := jsonMap["pedidoId"]
	assert.False(t, exists)
}
