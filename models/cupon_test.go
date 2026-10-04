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

	jsonData, err := json.Marshal(cupon)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["cuponId"])
	assert.Equal(t, "TEST2025", jsonMap["codigo"])
	assert.Equal(t, "GLOBAL", jsonMap["scope"])
	assert.Equal(t, float64(20), jsonMap["valorDescuento"])
	assert.Equal(t, true, jsonMap["activo"])

	fi, _ := jsonMap["fechaInicio"].(string)
	ff, _ := jsonMap["fechaFin"].(string)
	assert.Equal(t, "01-01-2025", fi)
	assert.Equal(t, "31-12-2025", ff)

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

	assert.False(t, cupon.Activo)

	assert.Nil(t, cupon.MaxUsos)
	assert.Nil(t, cupon.LimitePorCliente)
	assert.Nil(t, cupon.MontoMinimo)
	assert.Nil(t, cupon.PkIdProducto)
	assert.Nil(t, cupon.PkIdCategoria)
	assert.Nil(t, cupon.PkDocumentoCliente)
}
