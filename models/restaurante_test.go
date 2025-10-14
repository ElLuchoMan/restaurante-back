package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRestaurante_TableName(t *testing.T) {
	restaurante := &Restaurante{}
	assert.Equal(t, "restaurante", restaurante.TableName())
}

func TestRestaurante_JSONSerialization(t *testing.T) {
	horaApertura := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	restaurante := &Restaurante{
		PK_ID_RESTAURANTE:  1,
		NOMBRE_RESTAURANTE: "Restaurante El Buen Sabor",
		HORA_APERTURA:      horaApertura,
	}

	jsonData, err := json.Marshal(restaurante)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), jsonMap["restauranteId"])
	assert.Equal(t, "Restaurante El Buen Sabor", jsonMap["nombreRestaurante"])

	horaStr, _ := jsonMap["horaApertura"].(string)

	assert.Equal(t, "17:52:32", horaStr)
}

func TestRestaurante_CambioHorarioNull(t *testing.T) {
	restaurante := &Restaurante{
		PK_ID_RESTAURANTE:  1,
		NOMBRE_RESTAURANTE: "Restaurante Test",
		HORA_APERTURA:      time.Now(),
	}

	assert.Nil(t, restaurante.PK_ID_CAMBIO_HORARIO)
}
