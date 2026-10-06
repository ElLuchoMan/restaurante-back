package routers

import (
	"net/http"
	"testing"
)

// Los datos personales de reservas y contactos solo los ve el personal.
func TestReservasListadosSoloPersonal(t *testing.T) {
	rutas := []string{
		"/restaurante/v1/reservas",
		"/restaurante/v1/reservas/parameter",
		"/restaurante/v1/reservas/documento",
		"/restaurante/v1/reserva_contacto",
		"/restaurante/v1/reserva_contacto/search",
	}
	for _, ruta := range rutas {
		if w := serveAuth(http.MethodGet, ruta, ""); w.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s sin token: esperado 401, obtenido %d", ruta, w.Code)
		}
		if w := serveAuth(http.MethodGet, ruta, tokenConRol(t, "Cliente", "")); w.Code != http.StatusForbidden {
			t.Fatalf("GET %s cliente: esperado 403, obtenido %d", ruta, w.Code)
		}
	}
	// un trabajador pasa el filtro y llega al controlador, que valida (400) antes de tocar la BD
	for _, ruta := range []string{
		"/restaurante/v1/reservas/documento",
		"/restaurante/v1/reserva_contacto/search",
	} {
		if w := serveAuth(http.MethodGet, ruta, tokenConRol(t, "Mesero", "")); w.Code != http.StatusBadRequest {
			t.Fatalf("GET %s mesero: esperado 400 del controlador, obtenido %d %s", ruta, w.Code, w.Body.String())
		}
	}
}

// search/cliente/put/delete exigen token (el controlador limita a cada Cliente a lo suyo).
func TestReservasConTokenObligatorio(t *testing.T) {
	rutas := []struct{ method, path string }{
		{http.MethodGet, "/restaurante/v1/reservas/search?id=1"},
		{http.MethodGet, "/restaurante/v1/reservas/cliente?documentoCliente=1"},
		{http.MethodPut, "/restaurante/v1/reservas?id=1"},
		{http.MethodDelete, "/restaurante/v1/reservas?id=1"},
	}
	for _, r := range rutas {
		if w := serveAuth(r.method, r.path, ""); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s sin token: esperado 401, obtenido %d", r.method, r.path, w.Code)
		}
		if w := serveAuth(r.method, r.path, "no-es-un-jwt"); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s token inválido: esperado 401, obtenido %d", r.method, r.path, w.Code)
		}
	}
	// un Cliente con documento distinto recibe 403 en /cliente antes de tocar la BD
	if w := serveAuth(http.MethodGet, "/restaurante/v1/reservas/cliente?documentoCliente=99", tokenConRol(t, "Cliente", "")); w.Code != http.StatusForbidden {
		t.Fatalf("cliente con documento ajeno: esperado 403, obtenido %d %s", w.Code, w.Body.String())
	}
}

// Públicos: crear reserva (con o sin cuenta) y la consulta de invitado.
func TestReservasRutasPublicas(t *testing.T) {
	if w := serveAuth(http.MethodPost, "/restaurante/v1/reservas", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("POST /reservas sin token debe llegar al controlador (400), obtenido %d %s", w.Code, w.Body.String())
	}
	if w := serveAuth(http.MethodGet, "/restaurante/v1/reservas/consulta", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("GET /reservas/consulta sin token debe llegar al controlador (400), obtenido %d %s", w.Code, w.Body.String())
	}
}
