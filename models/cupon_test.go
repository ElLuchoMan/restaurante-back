package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCupon_TableName(t *testing.T) {
	cupon := &Cupon{}
	assert.Equal(t, "cupon", cupon.TableName())
}

func TestCupon_JSONSerialization(t *testing.T) {
	maxUsos := 10
	limitePorCliente := 5
	montoMinimo := int64(10000)

	cupon := &Cupon{
		PkIdCupon:        1,
		Codigo:           "TEST2025",
		Scope:            CuponScopeGlobal,
		TipoDescuento:    TipoDescuentoPorcentaje,
		ValorDescuento:   20,
		MaxUsos:          &maxUsos,
		LimitePorCliente: &limitePorCliente,
		MontoMinimo:      &montoMinimo,
		FechaInicio:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Activo:           true,
	}

	// Serializar
	jsonData, err := json.Marshal(cupon)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar campos en JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["cuponId"])
	assert.Equal(t, "TEST2025", jsonMap["codigo"])
	assert.Equal(t, "GLOBAL", jsonMap["scope"])
	assert.Equal(t, float64(20), jsonMap["valorDescuento"])
	assert.Equal(t, true, jsonMap["activo"])

	// Deserializar
	var cuponDeserialized Cupon
	err = json.Unmarshal(jsonData, &cuponDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, cupon.Codigo, cuponDeserialized.Codigo)
	assert.Equal(t, cupon.Scope, cuponDeserialized.Scope)
	assert.Equal(t, cupon.ValorDescuento, cuponDeserialized.ValorDescuento)
}

func TestCupon_DefaultValues(t *testing.T) {
	cupon := &Cupon{
		Codigo:         "DEFAULT",
		Scope:          CuponScopeGlobal,
		TipoDescuento:  TipoDescuentoPorcentaje,
		ValorDescuento: 10,
		FechaInicio:    time.Now(),
		FechaFin:       time.Now().AddDate(0, 1, 0),
	}

	// Por defecto, Activo es false en Go (zero value)
	// pero en la DB se establece como true
	assert.False(t, cupon.Activo) // Go zero value

	// Valores nulos deben ser nil
	assert.Nil(t, cupon.MaxUsos)
	assert.Nil(t, cupon.LimitePorCliente)
	assert.Nil(t, cupon.MontoMinimo)
	assert.Nil(t, cupon.PkIdProducto)
	assert.Nil(t, cupon.PkIdCategoria)
	assert.Nil(t, cupon.PkDocumentoCliente)
}
