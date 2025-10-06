package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliente_TableName(t *testing.T) {
	cliente := &Cliente{}
	assert.Equal(t, "cliente", cliente.TableName())
}

func TestCliente_JSONSerialization(t *testing.T) {
	observaciones := "Cliente VIP"
	cliente := &Cliente{
		PK_DOCUMENTO_CLIENTE: 123456789,
		NOMBRE:               "Juan",
		APELLIDO:             "Pérez",
		CORREO:               "juan.perez@example.com",
		DIRECCION:            "Calle 123",
		TELEFONO:             "3001234567",
		OBSERVACIONES:        &observaciones,
		PASSWORD:             "hashed_password",
	}

	// Serializar
	jsonData, err := json.Marshal(cliente)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar campos en JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(123456789), jsonMap["documentoCliente"])
	assert.Equal(t, "Juan", jsonMap["nombre"])
	assert.Equal(t, "juan.perez@example.com", jsonMap["correo"])

	// Deserializar
	var clienteDeserialized Cliente
	err = json.Unmarshal(jsonData, &clienteDeserialized)
	assert.NoError(t, err)
	assert.Equal(t, cliente.NOMBRE, clienteDeserialized.NOMBRE)
	assert.Equal(t, cliente.CORREO, clienteDeserialized.CORREO)
}

func TestCliente_ObservacionesNull(t *testing.T) {
	cliente := &Cliente{
		PK_DOCUMENTO_CLIENTE: 123456789,
		NOMBRE:               "Juan",
		APELLIDO:             "Pérez",
		CORREO:               "juan@example.com",
		DIRECCION:            "Calle 123",
		TELEFONO:             "3001234567",
		PASSWORD:             "pass",
	}

	// Observaciones debe ser nil por defecto
	assert.Nil(t, cliente.OBSERVACIONES)

	// Serializar y verificar que observaciones no aparece en JSON
	jsonData, err := json.Marshal(cliente)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// observaciones debe ser nil (null en JSON)
	_, exists := jsonMap["observaciones"]
	assert.True(t, exists) // Existe pero es null
}
