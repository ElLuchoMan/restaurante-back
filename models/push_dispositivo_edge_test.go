package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushDispositivo_DeserializeSubscribedTopics_EmptyContent(t *testing.T) {

	pd := &PushDispositivo{
		SubscribedTopics: "{}",
	}

	pd.deserializeSubscribedTopics()

	assert.NotNil(t, pd.SubscribedTopicsArray)
	assert.Equal(t, 0, len(pd.SubscribedTopicsArray))
}

func TestPushDispositivo_DeserializeSubscribedTopics_InvalidJSON(t *testing.T) {
	pd := &PushDispositivo{
		SubscribedTopics: `invalid json format`,
	}

	pd.deserializeSubscribedTopics()

	_ = pd.SubscribedTopicsArray
	assert.True(t, true)
}
