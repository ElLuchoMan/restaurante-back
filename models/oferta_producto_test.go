package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOfertaProducto_TableName(t *testing.T) {
	op := &OfertaProducto{}
	assert.Equal(t, "oferta_producto", op.TableName())
}

func TestOfertaProducto_TableUnique(t *testing.T) {
	op := &OfertaProducto{}
	unique := op.TableUnique()

	assert.NotNil(t, unique)
	assert.Equal(t, 1, len(unique))
	assert.Equal(t, 2, len(unique[0]))
	assert.Equal(t, "PkIdOferta", unique[0][0])
	assert.Equal(t, "PkIdProducto", unique[0][1])
}

func TestOfertaProducto_JSONSerialization(t *testing.T) {
	op := &OfertaProducto{}

	// Serializar
	jsonData, err := json.Marshal(op)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserializar
	var opDeserialized OfertaProducto
	err = json.Unmarshal(jsonData, &opDeserialized)
	assert.NoError(t, err)
}
