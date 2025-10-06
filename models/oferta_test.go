package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOferta_TableName(t *testing.T) {
	oferta := &Oferta{}
	assert.Equal(t, "oferta", oferta.TableName())
}

func TestOferta_SerializeDiasSemana(t *testing.T) {
	oferta := &Oferta{
		Titulo:          "Oferta Test",
		TipoDescuento:   TipoDescuentoPorcentaje,
		ValorDescuento:  20,
		FechaInicio:     time.Now(),
		FechaFin:        time.Now().AddDate(0, 1, 0),
		DiasSemanaArray: []string{"Lunes", "Martes", "Miércoles"},
	}

	// Llamar a serializeDiasSemana
	oferta.serializeDiasSemana()

	// Verificar que DiasSemana contiene JSON válido
	assert.NotEmpty(t, oferta.DiasSemana)

	var diasArray []string
	err := json.Unmarshal([]byte(oferta.DiasSemana), &diasArray)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(diasArray))
	assert.Equal(t, "Lunes", diasArray[0])
	assert.Equal(t, "Martes", diasArray[1])
	assert.Equal(t, "Miércoles", diasArray[2])
}

func TestOferta_SerializeDiasSemana_Empty(t *testing.T) {
	oferta := &Oferta{
		DiasSemanaArray: []string{},
	}

	oferta.serializeDiasSemana()

	// DiasSemana debe estar vacío cuando el array está vacío
	assert.Equal(t, "", oferta.DiasSemana)
}

func TestOferta_DeserializeDiasSemana(t *testing.T) {
	oferta := &Oferta{
		DiasSemana: `["Lunes","Martes","Miércoles"]`,
	}

	// Llamar a deserializeDiasSemana
	oferta.deserializeDiasSemana()

	// Verificar que DiasSemanaArray se llenó correctamente
	assert.Equal(t, 3, len(oferta.DiasSemanaArray))
	assert.Equal(t, "Lunes", oferta.DiasSemanaArray[0])
	assert.Equal(t, "Martes", oferta.DiasSemanaArray[1])
	assert.Equal(t, "Miércoles", oferta.DiasSemanaArray[2])
}

func TestOferta_DeserializeDiasSemana_Empty(t *testing.T) {
	oferta := &Oferta{
		DiasSemana: "",
	}

	oferta.deserializeDiasSemana()

	// DiasSemanaArray debe estar vacío
	assert.NotNil(t, oferta.DiasSemanaArray)
	assert.Equal(t, 0, len(oferta.DiasSemanaArray))
}

func TestOferta_BeforeInsert(t *testing.T) {
	oferta := &Oferta{
		Titulo:          "Oferta Test",
		TipoDescuento:   TipoDescuentoPorcentaje,
		ValorDescuento:  20,
		FechaInicio:     time.Now(),
		FechaFin:        time.Now().AddDate(0, 1, 0),
		DiasSemanaArray: []string{"Lunes", "Viernes"},
	}

	// Llamar al hook BeforeInsert
	oferta.BeforeInsert()

	// Verificar que se serializó DiasSemana
	assert.NotEmpty(t, oferta.DiasSemana)
	assert.Contains(t, oferta.DiasSemana, "Lunes")
	assert.Contains(t, oferta.DiasSemana, "Viernes")
}

func TestOferta_BeforeUpdate(t *testing.T) {
	oferta := &Oferta{
		Titulo:          "Oferta Test",
		TipoDescuento:   TipoDescuentoPorcentaje,
		ValorDescuento:  20,
		FechaInicio:     time.Now(),
		FechaFin:        time.Now().AddDate(0, 1, 0),
		DiasSemanaArray: []string{"Sábado", "Domingo"},
	}

	// Llamar al hook BeforeUpdate
	oferta.BeforeUpdate()

	// Verificar que se serializó DiasSemana
	assert.NotEmpty(t, oferta.DiasSemana)
	assert.Contains(t, oferta.DiasSemana, "Sábado")
	assert.Contains(t, oferta.DiasSemana, "Domingo")
}

func TestOferta_AfterLoad(t *testing.T) {
	oferta := &Oferta{
		DiasSemana: `["Lunes","Martes","Miércoles","Jueves","Viernes"]`,
	}

	// Llamar al hook AfterLoad
	oferta.AfterLoad()

	// Verificar que se deserializó DiasSemanaArray
	assert.Equal(t, 5, len(oferta.DiasSemanaArray))
	assert.Equal(t, "Lunes", oferta.DiasSemanaArray[0])
	assert.Equal(t, "Viernes", oferta.DiasSemanaArray[4])
}

func TestOferta_JSONSerialization(t *testing.T) {
	oferta := &Oferta{
		PkIdOferta:      1,
		Titulo:          "Oferta Especial",
		TipoDescuento:   TipoDescuentoPorcentaje,
		ValorDescuento:  25,
		FechaInicio:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:        time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		DiasSemanaArray: []string{"Lunes", "Martes"},
		Activo:          true,
	}

	// Serializar
	jsonData, err := json.Marshal(oferta)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar que diasSemana está en el JSON (no DiasSemana que tiene json:"-")
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, "Oferta Especial", jsonMap["titulo"])
	assert.NotNil(t, jsonMap["diasSemana"])

	// Verificar que DiasSemana NO está en el JSON (tiene json:"-")
	_, exists := jsonMap["dias_semana"]
	assert.False(t, exists)
}
