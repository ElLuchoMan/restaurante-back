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
       req := httptest.NewRequest(
               http.MethodPost,
               "/nominas?generar_nomina_automatica=true&verificar_nomina=true&p_nomina_id=1&p_fecha=2024-01-01",
               body,
       )
       req.Header.Set("Content-Type", "application/json")
       w := sendRequest(req)
       if w.Code != http.StatusCreated {
               t.Fatalf("nomina generation/verification failed or DB unavailable: %d %s", w.Code, w.Body.String())
       }

       var resp models.ApiResponse
       if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
               t.Fatalf("failed to unmarshal response: %v", err)
       }

       if resp.Message != "Funciones de nómina ejecutadas correctamente" {
               t.Errorf("unexpected response message: %s", resp.Message)
       }

       dataMap, ok := resp.Data.(map[string]interface{})
       if !ok {
               t.Fatalf("expected response data to be a map, got %T", resp.Data)
       }
       nominaPayload, ok := dataMap["nomina"]
       if !ok {
               t.Fatalf("expected 'nomina' field in response data")
       }

       dataBytes, err := json.Marshal(nominaPayload)
       if err != nil {
               t.Fatalf("failed to marshal nomina payload: %v", err)
       }
       var nomina models.Nomina
       if err := json.Unmarshal(dataBytes, &nomina); err != nil {
               t.Fatalf("failed to unmarshal nomina: %v", err)
       }
       if nomina.PK_ID_NOMINA == 0 {
               t.Errorf("expected valid nomina ID, got %d", nomina.PK_ID_NOMINA)
       }
}
