package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// TEST DE COBERTURA PARA deserializeSubscribedTopics
// Caso específico: comillas escapadas (línea 88)
// ============================================================================

func TestPushDispositivo_DeserializeSubscribedTopics_ConComillasEscapadas(t *testing.T) {
	// Este test cubre la línea 88 donde se desescapan las comillas dobles
	// Formato PostgreSQL con comillas escapadas: {"topic ""quoted"""}
	pd := &PushDispositivo{
		SubscribedTopics: `{"topic ""with quotes""","normal"}`,
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 2, len(pd.SubscribedTopicsArray))
	assert.Equal(t, `topic "with quotes"`, pd.SubscribedTopicsArray[0])
	assert.Equal(t, "normal", pd.SubscribedTopicsArray[1])
}

func TestPushDispositivo_DeserializeSubscribedTopics_SoloComillasEscapadas(t *testing.T) {
	// Test adicional para asegurar cobertura completa del unescape
	pd := &PushDispositivo{
		SubscribedTopics: `{""""}`, // Solo comillas dobles escapadas
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 1, len(pd.SubscribedTopicsArray))
	assert.Equal(t, `"`, pd.SubscribedTopicsArray[0])
}

func TestPushDispositivo_DeserializeSubscribedTopics_MultiplesComillasEscapadas(t *testing.T) {
	// Test para múltiples comillas escapadas consecutivas
	pd := &PushDispositivo{
		SubscribedTopics: `{"text""with""""multiple""quotes"}`,
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 1, len(pd.SubscribedTopicsArray))
	assert.Equal(t, `text"with""multiple"quotes`, pd.SubscribedTopicsArray[0])
}

func TestPushDispositivo_DeserializeSubscribedTopics_SinComillas(t *testing.T) {
	// Test para valores sin comillas (línea 90, caso else implícito)
	pd := &PushDispositivo{
		SubscribedTopics: `{topic_sin_comillas,otro_topic}`,
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 2, len(pd.SubscribedTopicsArray))
	assert.Equal(t, "topic_sin_comillas", pd.SubscribedTopicsArray[0])
	assert.Equal(t, "otro_topic", pd.SubscribedTopicsArray[1])
}

func TestPushDispositivo_DeserializeSubscribedTopics_MixtoComillasYSin(t *testing.T) {
	// Test con mezcla de valores con y sin comillas
	pd := &PushDispositivo{
		SubscribedTopics: `{"con_comillas",sin_comillas,"otro ""escaped"""}`,
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 3, len(pd.SubscribedTopicsArray))
	assert.Equal(t, "con_comillas", pd.SubscribedTopicsArray[0])
	assert.Equal(t, "sin_comillas", pd.SubscribedTopicsArray[1])
	assert.Equal(t, `otro "escaped"`, pd.SubscribedTopicsArray[2])
}
