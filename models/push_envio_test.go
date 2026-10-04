package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPushEnvio_TableName(t *testing.T) {
	pe := &PushEnvio{}
	assert.Equal(t, "push_envio", pe.TableName())
}

func TestPushEnvio_SerializeData(t *testing.T) {
	dataJSON := json.RawMessage(`{"title":"Test","body":"Message"}`)
	pe := &PushEnvio{
		Proveedor: ProveedorWebPush,
		DataObj:   dataJSON,
	}

	pe.serializeData()

	assert.NotEmpty(t, pe.Data)
	assert.Contains(t, pe.Data, "title")
	assert.Contains(t, pe.Data, "Test")
}

func TestPushEnvio_SerializeData_Empty(t *testing.T) {
	pe := &PushEnvio{
		DataObj: json.RawMessage{},
	}

	pe.serializeData()

	assert.Equal(t, "", pe.Data)
}

func TestPushEnvio_DeserializeData(t *testing.T) {
	pe := &PushEnvio{
		Data: `{"title":"Notification","body":"Content"}`,
	}

	pe.deserializeData()

	assert.NotNil(t, pe.DataObj)
	assert.Contains(t, string(pe.DataObj), "title")
	assert.Contains(t, string(pe.DataObj), "Notification")
}

func TestPushEnvio_DeserializeData_Empty(t *testing.T) {
	pe := &PushEnvio{
		Data: "",
	}

	pe.deserializeData()

	assert.Nil(t, pe.DataObj)
}

func TestPushEnvio_BeforeInsert(t *testing.T) {
	dataJSON := json.RawMessage(`{"test":"data"}`)
	pe := &PushEnvio{
		DataObj: dataJSON,
	}

	pe.BeforeInsert()

	assert.NotEmpty(t, pe.Data)
	assert.Contains(t, pe.Data, "test")
}

func TestPushEnvio_BeforeUpdate(t *testing.T) {
	dataJSON := json.RawMessage(`{"updated":"data"}`)
	pe := &PushEnvio{
		DataObj: dataJSON,
	}

	pe.BeforeUpdate()

	assert.NotEmpty(t, pe.Data)
	assert.Contains(t, pe.Data, "updated")
}

func TestPushEnvio_AfterLoad(t *testing.T) {
	pe := &PushEnvio{
		Data: `{"loaded":"data"}`,
	}

	pe.AfterLoad()

	assert.NotNil(t, pe.DataObj)
	assert.Contains(t, string(pe.DataObj), "loaded")
}

func TestPushEnvio_JSONSerialization(t *testing.T) {
	dataJSON := json.RawMessage(`{"info":"test"}`)
	statusCode := 200
	errorCode := "none"

	pe := &PushEnvio{
		PkIdPushEnvio: 1,
		Proveedor:     ProveedorWebPush,
		DataObj:       dataJSON,
		Exito:         true,
		StatusCode:    &statusCode,
		ErrorCode:     &errorCode,
		SentAt:        time.Now(),
	}

	jsonData, err := json.Marshal(pe)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["pushEnvioId"])
	assert.Equal(t, "WEB_PUSH", jsonMap["proveedor"])
	assert.Equal(t, true, jsonMap["exito"])
	assert.Equal(t, float64(200), jsonMap["statusCode"])

	data, exists := jsonMap["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)
}

func TestPushEnvio_NullableFields(t *testing.T) {
	pe := &PushEnvio{
		PkIdPushEnvio: 1,
		Proveedor:     ProveedorFCM,
		Exito:         false,
		SentAt:        time.Now(),
	}

	assert.Nil(t, pe.StatusCode)
	assert.Nil(t, pe.ErrorCode)

	jsonData, err := json.Marshal(pe)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	_, exists := jsonMap["statusCode"]
	assert.False(t, exists)
}
