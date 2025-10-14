package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubcategoria_TableName(t *testing.T) {
	subcategoria := &Subcategoria{}
	assert.Equal(t, "subcategoria", subcategoria.TableName())
}

func TestSubcategoria_JSONSerialization(t *testing.T) {
	subcategoria := &Subcategoria{
		PK_ID_SUBCATEGORIA: 1,
		NOMBRE:             "Postres",
	}

	jsonData, err := json.Marshal(subcategoria)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["subcategoriaId"])
	assert.Equal(t, "Postres", jsonMap["nombre"])

	var subcategoriaDeserialized Subcategoria
	err = json.Unmarshal(jsonData, &subcategoriaDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, subcategoria.NOMBRE, subcategoriaDeserialized.NOMBRE)
}
