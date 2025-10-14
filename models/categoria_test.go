package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategoria_TableName(t *testing.T) {
	categoria := &Categoria{}
	assert.Equal(t, "categoria", categoria.TableName())
}

func TestCategoria_JSONSerialization(t *testing.T) {
	categoria := &Categoria{
		PK_ID_CATEGORIA: 1,
		NOMBRE:          "Bebidas",
	}

	jsonData, err := json.Marshal(categoria)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["categoriaId"])
	assert.Equal(t, "Bebidas", jsonMap["nombre"])

	var categoriaDeserialized Categoria
	err = json.Unmarshal(jsonData, &categoriaDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, categoria.NOMBRE, categoriaDeserialized.NOMBRE)
	assert.Equal(t, categoria.PK_ID_CATEGORIA, categoriaDeserialized.PK_ID_CATEGORIA)
}
