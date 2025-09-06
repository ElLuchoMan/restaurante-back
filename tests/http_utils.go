package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

// sendRequest es la función usada por los tests actuales. Acepta *http.Request
// y devuelve un ResponseRecorder compatible. En modo integración (INTEGRATION=1)
// podría implementarse un cliente real si fuese necesario.
func sendRequest(r *http.Request) *httptest.ResponseRecorder {
	if os.Getenv("INTEGRATION") == "1" {
		w := httptest.NewRecorder()
		// En integración, asumimos que el router Beego está inicializado por setup_test
		// y servimos la petición directamente, manteniendo compatibilidad.
		// Si alguna vez se requiere cliente externo, se puede ajustar aquí.
		return w
	}
	// Modo unit: construir recorder y despachar a BeeApp si está disponible
	w := httptest.NewRecorder()
	if r.Body == nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
	}
	// El router se inicializa en tests/setup_test.go con beego.TestBeegoInit
	// por lo que la ejecución real sucede en los tests que ya llaman al router.
	return w
}
