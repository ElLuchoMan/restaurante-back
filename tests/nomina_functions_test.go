package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"
)

func TestNominaGenerarYVerificar(t *testing.T) {
	body := strings.NewReader("{}")
	req := httptest.NewRequest(http.MethodPost, "/nominas?generar_nomina_automatica=true&verificar_nomina=true", body)
	req.Header.Set("Content-Type", "application/json")
	w := sendRequest(req)
	if w.Code != http.StatusCreated {
		t.Fatalf("nomina generation/verification failed or DB unavailable: %d %s", w.Code, w.Body.String())
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "Nómina creada correctamente" {
		t.Errorf("unexpected response message: %s", resp.Message)
	}

	if resp.Data == nil {
		t.Fatalf("expected response data")
	}
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	var nomina models.Nomina
	if err := json.Unmarshal(dataBytes, &nomina); err != nil {
		t.Fatalf("failed to unmarshal nomina: %v", err)
	}
	if nomina.PK_ID_NOMINA == 0 {
		t.Errorf("expected valid nomina ID, got %d", nomina.PK_ID_NOMINA)
	}
}
