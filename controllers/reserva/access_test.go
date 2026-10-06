package reserva

import (
	"net/http"
	"testing"
)

func filtro(method, target, auth string) int {
	ctx, w := newCtx(method, target, "")
	if auth != "" {
		ctx.Request.Header.Set("Authorization", auth)
	}
	AccessFilter(ctx)
	return w.Code
}

func TestAccessFilterListadosSoloPersonal(t *testing.T) {
	for _, ruta := range []string{
		"/restaurante/v1/reservas", "/restaurante/v1/reservas/", "/restaurante/v1/reservas?x=1",
		"/restaurante/v1/reservas/parameter", "/restaurante/v1/reservas/documento",
	} {
		if got := filtro(http.MethodGet, ruta, ""); got != http.StatusUnauthorized {
			t.Fatalf("GET %s sin token: %d", ruta, got)
		}
		if got := filtro(http.MethodGet, ruta, tokenDe(rolCliente, 9)); got != http.StatusForbidden {
			t.Fatalf("GET %s cliente: %d", ruta, got)
		}
		if got := filtro(http.MethodGet, ruta, tokenDe(rolMesero, 1)); got != http.StatusOK {
			t.Fatalf("GET %s personal: %d", ruta, got)
		}
	}
}

func TestAccessFilterRutasConToken(t *testing.T) {
	for _, c := range []struct{ method, ruta string }{
		{http.MethodGet, "/restaurante/v1/reservas/search?id=1"},
		{http.MethodGet, "/restaurante/v1/reservas/cliente"},
		{http.MethodPut, "/restaurante/v1/reservas?id=1"},
		{http.MethodDelete, "/restaurante/v1/reservas?id=1"},
	} {
		if got := filtro(c.method, c.ruta, ""); got != http.StatusUnauthorized {
			t.Fatalf("%s %s sin token: %d", c.method, c.ruta, got)
		}
		// con token de Cliente pasa el filtro: el controlador limita a lo suyo
		if got := filtro(c.method, c.ruta, tokenDe(rolCliente, 9)); got != http.StatusOK {
			t.Fatalf("%s %s cliente: %d", c.method, c.ruta, got)
		}
	}
}

func TestAccessFilterRutasPublicas(t *testing.T) {
	if got := filtro(http.MethodPost, "/restaurante/v1/reservas", ""); got != http.StatusOK {
		t.Fatalf("POST /reservas debe ser público: %d", got)
	}
	if got := filtro(http.MethodGet, "/restaurante/v1/reservas/consulta?reservaId=1&contacto=2", ""); got != http.StatusOK {
		t.Fatalf("GET /reservas/consulta debe ser público: %d", got)
	}
}
