package metodopago

import (
	"net/http"
	"testing"
)

func TestMutacionesSoloAdministrador(t *testing.T) {
	defer resetFake()
	found()
	p := func(c *MetodoPagoController) { c.Post() }
	u := func(c *MetodoPagoController) { c.Put() }
	d := func(c *MetodoPagoController) { c.Delete() }
	g := func(c *MetodoPagoController) { c.GetAll() }
	s := func(c *MetodoPagoController) { c.GetById() }

	sinToken(t)
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"X"}`, p, http.StatusUnauthorized)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusUnauthorized)
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusUnauthorized)

	for _, rol := range []string{rolClienteT, rolMesero, rolDomi} {
		como(t, rol, 7)
		call(t, http.MethodPost, "/metodos_pago", `{"tipo":"X"}`, p, http.StatusForbidden)
		call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusForbidden)
		call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusForbidden)
		// la lectura queda abierta a cualquier sesión (el carrito del cliente lista los métodos)
		call(t, http.MethodGet, "/metodos_pago", "", g, http.StatusOK)
		call(t, http.MethodGet, "/metodos_pago/search?id=3", "", s, http.StatusOK)
	}

	como(t, rolAdmin, 1)
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"X"}`, p, http.StatusCreated)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusOK)
	fakeQuery = nil
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusOK)
}
