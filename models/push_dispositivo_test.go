package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPushDispositivo_TableName(t *testing.T) {
	pd := &PushDispositivo{}
	assert.Equal(t, "push_dispositivo", pd.TableName())
}

func TestPushDispositivo_SerializeSubscribedTopics(t *testing.T) {
	pd := &PushDispositivo{
		Plataforma:            PlataformaWeb,
		SubscribedTopicsArray: []string{"ofertas", "noticias", "promociones"},
	}

	pd.serializeSubscribedTopics()

	// Verificar que se serializó como array de PostgreSQL
	assert.NotEmpty(t, pd.SubscribedTopics)
	assert.Contains(t, pd.SubscribedTopics, "ofertas")
	assert.Contains(t, pd.SubscribedTopics, "noticias")
	assert.True(t, pd.SubscribedTopics[0] == '{')
	assert.True(t, pd.SubscribedTopics[len(pd.SubscribedTopics)-1] == '}')
}

func TestPushDispositivo_SerializeSubscribedTopics_Empty(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopicsArray: []string{},
	}

	pd.serializeSubscribedTopics()

	assert.Equal(t, "", pd.SubscribedTopics)
}

func TestPushDispositivo_SerializeSubscribedTopics_WithQuotes(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopicsArray: []string{"topic with \"quotes\"", "normal topic"},
	}

	pd.serializeSubscribedTopics()

	// Verificar que las comillas se escaparon correctamente
	assert.NotEmpty(t, pd.SubscribedTopics)
	assert.Contains(t, pd.SubscribedTopics, `""`)
}

func TestPushDispositivo_DeserializeSubscribedTopics(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: `{"ofertas","noticias","promociones"}`,
	}

	pd.deserializeSubscribedTopics()

	assert.Equal(t, 3, len(pd.SubscribedTopicsArray))
	assert.Equal(t, "ofertas", pd.SubscribedTopicsArray[0])
	assert.Equal(t, "noticias", pd.SubscribedTopicsArray[1])
	assert.Equal(t, "promociones", pd.SubscribedTopicsArray[2])
}

func TestPushDispositivo_DeserializeSubscribedTopics_Empty(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: "",
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 0, len(pd.SubscribedTopicsArray))
}

func TestPushDispositivo_DeserializeSubscribedTopics_EmptyArray(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: "{}",
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 0, len(pd.SubscribedTopicsArray))
}

func TestPushDispositivo_DeserializeSubscribedTopics_WithQuotes(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: `{"topic with ""quotes""","normal topic"}`,
	}

	pd.deserializeSubscribedTopics()

	assert.Equal(t, 2, len(pd.SubscribedTopicsArray))
	assert.Equal(t, `topic with "quotes"`, pd.SubscribedTopicsArray[0])
	assert.Equal(t, "normal topic", pd.SubscribedTopicsArray[1])
}

func TestPushDispositivo_DeserializeSubscribedTopics_JSONFallback(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: `["ofertas","noticias"]`,
	}

	pd.deserializeSubscribedTopics()

	// Debe usar fallback de JSON
	assert.Equal(t, 2, len(pd.SubscribedTopicsArray))
	assert.Equal(t, "ofertas", pd.SubscribedTopicsArray[0])
	assert.Equal(t, "noticias", pd.SubscribedTopicsArray[1])
}

func TestPushDispositivo_BeforeInsert(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopicsArray: []string{"topic1", "topic2"},
	}

	pd.BeforeInsert()

	// Verificar que se serializó
	assert.NotEmpty(t, pd.SubscribedTopics)
	assert.Contains(t, pd.SubscribedTopics, "topic1")
}

func TestPushDispositivo_BeforeUpdate(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopicsArray: []string{"updated1", "updated2"},
	}

	pd.BeforeUpdate()

	// Verificar que se serializó
	assert.NotEmpty(t, pd.SubscribedTopics)
	assert.Contains(t, pd.SubscribedTopics, "updated1")
}

func TestPushDispositivo_AfterLoad(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: `{"loaded1","loaded2"}`,
	}

	pd.AfterLoad()

	// Verificar que se deserializó
	assert.Equal(t, 2, len(pd.SubscribedTopicsArray))
	assert.Equal(t, "loaded1", pd.SubscribedTopicsArray[0])
}

func TestPushDispositivo_JSONSerialization(t *testing.T) {
	endpoint := "https://test.push.com"
	pd := &PushDispositivo{
		PkIdPushDispositivo:   1,
		Plataforma:            PlataformaWeb,
		Endpoint:              &endpoint,
		Enabled:               true,
		SubscribedTopicsArray: []string{"news", "sports"},
		CreatedAt:             time.Now(),
	}

	// Serializar
	jsonData, err := json.Marshal(pd)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verificar campos en JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["pushDispositivoId"])
	assert.Equal(t, "WEB", jsonMap["plataforma"])
	assert.Equal(t, true, jsonMap["enabled"])

	// Verificar que subscribedTopics está como array
	topics, ok := jsonMap["subscribedTopics"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(topics))
}
