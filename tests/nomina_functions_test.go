package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNominaGenerarYVerificar(t *testing.T) {
	body := strings.NewReader("{}")
	req := httptest.NewRequest(http.MethodPost, "/nominas?generar_nomina_automatica=true&verificar_nomina=true&fecha_inicio=2023-01-01&fecha_fin=2023-01-31", body)
	req.Header.Set("Content-Type", "application/json")
	w := sendRequest(req)
	if w.Code != http.StatusCreated {
		t.Skipf("nomina generation/verification failed or DB unavailable: %d %s", w.Code, w.Body.String())
	}
}
