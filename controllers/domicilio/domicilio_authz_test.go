package domicilio

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

const (
	bodyDomicilio = `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31"%s}`
	clienteSQL    = "FROM pedido p\nJOIN cliente"
)

func domBody(extra string) string { return strings.Replace(bodyDomicilio, "%s", extra, 1) }

func clienteDelDomicilio(doc int64) rt {
	return rt{clienteSQL, func() driver.Rows { return rowsOf(clienteCols, []driver.Value{doc, "Juan", "Pérez"}) }}
}

func TestSinTokenEs401(t *testing.T) {
	defer resetFake()
	sinToken(t)
	call(t, http.MethodGet, "/domicilios", "", func(c *DomicilioController) { c.GetAll() }, http.StatusUnauthorized)
	call(t, http.MethodGet, "/domicilios/search?id=3", "", func(c *DomicilioController) { c.GetById() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/domicilios", domBody(""), func(c *DomicilioController) { c.Post() }, http.StatusUnauthorized)
	call(t, http.MethodPut, "/domicilios?id=3", `{}`, func(c *DomicilioController) { c.Put() }, http.StatusUnauthorized)
	call(t, http.MethodDelete, "/domicilios?id=3", "", func(c *DomicilioController) { c.Delete() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/domicilios/asignar?domicilio_id=3&trabajador_id=77", "", func(c *DomicilioController) { c.AsignarDomiciliario() }, http.StatusUnauthorized)
}

func TestClienteSoloCreaYLeeLoSuyo(t *testing.T) {
	defer resetFake()
	serve(domOK("PENDIENTE", false, nil))
	como(t, rolClienteT, 1001)
	call(t, http.MethodGet, "/domicilios", "", func(c *DomicilioController) { c.GetAll() }, http.StatusForbidden)
	call(t, http.MethodPut, "/domicilios?id=3", `{}`, func(c *DomicilioController) { c.Put() }, http.StatusForbidden)
	call(t, http.MethodDelete, "/domicilios?id=3", "", func(c *DomicilioController) { c.Delete() }, http.StatusForbidden)
	call(t, http.MethodPost, "/domicilios/asignar?domicilio_id=3&trabajador_id=1001", "", func(c *DomicilioController) { c.AsignarDomiciliario() }, http.StatusForbidden)
}

func TestGetByIdCliente(t *testing.T) {
	defer resetFake()
	g := func(c *DomicilioController) { c.GetById() }
	const url = "/domicilios/search?id=3"
	como(t, rolClienteT, 1001)
	serve(domOK("PENDIENTE", false, nil), clienteDelDomicilio(1001))
	b := call(t, http.MethodGet, url, "", g, http.StatusOK)
	contains(t, b, `"cliente":{"documento":1001`)
	// domicilio de otro cliente
	serve(domOK("PENDIENTE", false, nil), clienteDelDomicilio(2002))
	if b := call(t, http.MethodGet, url, "", g, http.StatusNotFound); !strings.Contains(b, "Domicilio no encontrado") || strings.Contains(b, "Pérez") {
		t.Fatalf("ajeno debe verse como inexistente: %s", b)
	}
	// domicilio sin pedido
	serve(domOK("PENDIENTE", false, nil))
	call(t, http.MethodGet, url, "", g, http.StatusNotFound)
	// token de cliente sin documento
	como(t, rolClienteT, 0)
	serve(domOK("PENDIENTE", false, nil), clienteDelDomicilio(0))
	call(t, http.MethodGet, url, "", g, http.StatusNotFound)
	// el personal ve cualquiera
	como(t, rolDomi, 9)
	serve(domOK("PENDIENTE", false, nil))
	call(t, http.MethodGet, url, "", g, http.StatusOK)
}

func TestPostCliente(t *testing.T) {
	defer resetFake()
	p := func(c *DomicilioController) { c.Post() }
	serve(domOK("PENDIENTE", false, nil), trabajadorCount(1))
	como(t, rolClienteT, 1001)
	call(t, http.MethodPost, "/domicilios", domBody(""), p, http.StatusCreated)
	call(t, http.MethodPost, "/domicilios", domBody(`,"estadoDomicilio":"pendiente","trabajadorAsignado":0`), p, http.StatusCreated)
	call(t, http.MethodPost, "/domicilios", domBody(`,"estado":"PENDIENTE","trabajadorAsignado":null`), p, http.StatusCreated)
	// no puede asignar un domiciliario ni crearlo ya en camino o entregado
	call(t, http.MethodPost, "/domicilios", domBody(`,"trabajadorAsignado":77`), p, http.StatusForbidden)
	call(t, http.MethodPost, "/domicilios", domBody(`,"estadoDomicilio":"ENTREGADO"`), p, http.StatusForbidden)
	call(t, http.MethodPost, "/domicilios", domBody(`,"estado":"EN_CAMINO"`), p, http.StatusForbidden)
	// el personal sí
	como(t, rolMesero, 5)
	call(t, http.MethodPost, "/domicilios", domBody(`,"estadoDomicilio":"EN_CAMINO","trabajadorAsignado":77`), p, http.StatusCreated)
}

func TestAsignarDomiciliarioPorRol(t *testing.T) {
	defer resetFake()
	a := func(c *DomicilioController) { c.AsignarDomiciliario() }
	const url = "/domicilios/asignar?domicilio_id=3&trabajador_id=77"
	serve(trabajadorCount(1), domOK("EN_CAMINO", false, int64(77)))
	// el Domiciliario solo se asigna a sí mismo
	como(t, rolDomi, 77)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	como(t, rolDomi, 78)
	if b := call(t, http.MethodPost, url, "", a, http.StatusForbidden); !strings.Contains(b, "propio Domiciliario") {
		t.Fatalf("mensaje poco claro: %s", b)
	}
	// el Administrador asigna a cualquiera
	como(t, rolAdmin, 1)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	// los demás roles de personal no reparten domicilios, ni siquiera a sí mismos
	como(t, rolMesero, 77)
	call(t, http.MethodPost, url, "", a, http.StatusForbidden)
}
