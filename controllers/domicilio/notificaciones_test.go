package domicilio

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"

	"restaurante/internal/notify"
)

// grabador sustituye a notify.Default y guarda los eventos recibidos.
type grabador struct{ eventos []notify.Evento }

func (g *grabador) Notificar(ev notify.Evento) { g.eventos = append(g.eventos, ev) }

func grabar(t *testing.T) *grabador {
	t.Helper()
	prev := notify.Default
	g := &grabador{}
	notify.Default = g
	t.Cleanup(func() { notify.Default = prev })
	return g
}

// leyendo hace que la 1.ª lectura del domicilio devuelva `antes` y las
// siguientes `despues` (la relectura tras el UPDATE).
func leyendo(antes, despues string) {
	lecturas := 0
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if !strings.Contains(q, `FROM "domicilio"`) {
			return rowsOf(nil), nil
		}
		lecturas++
		if lecturas == 1 {
			return rowsOf(domCols, domRow(antes, antes == "ENTREGADO", nil)), nil
		}
		return rowsOf(domCols, domRow(despues, despues == "ENTREGADO", nil)), nil
	}
}

func TestPutNotificaSoloLaTransicionAEntregado(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	u := func(c *DomicilioController) { c.Put() }

	leyendo("EN_CAMINO", "ENTREGADO")
	call(t, http.MethodPut, "/domicilios?id=3", `{"estado":"ENTREGADO"}`, u, http.StatusOK)
	if len(g.eventos) != 1 || g.eventos[0].Tipo != notify.DomicilioEntregado || g.eventos[0].DomicilioID != 3 {
		t.Fatalf("debía notificar la entrega: %+v", g.eventos)
	}

	// ya estaba entregado: no se repite el aviso
	g.eventos = nil
	leyendo("ENTREGADO", "ENTREGADO")
	call(t, http.MethodPut, "/domicilios?id=3", `{"observaciones":"x"}`, u, http.StatusOK)
	// otro cambio de estado distinto de ENTREGADO
	leyendo("PENDIENTE", "EN_CAMINO")
	call(t, http.MethodPut, "/domicilios?id=3", `{"estado":"EN_CAMINO"}`, u, http.StatusOK)
	if len(g.eventos) != 0 {
		t.Fatalf("no debía notificar: %+v", g.eventos)
	}
}

func TestAsignarNotificaAlCliente(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	a := func(c *DomicilioController) { c.AsignarDomiciliario() }
	url := "/domicilios/asignar?domicilio_id=3&trabajador_id=77"

	serve(trabajadorCount(1), domOK("EN_CAMINO", false, int64(77)))
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	if len(g.eventos) != 1 || g.eventos[0].Tipo != notify.DomicilioAsignado || g.eventos[0].DomicilioID != 3 {
		t.Fatalf("debía notificar la asignación: %+v", g.eventos)
	}

	// ya asignado -> 409 y sin aviso
	g.eventos = nil
	fakeAffected = 0
	call(t, http.MethodPost, url, "", a, http.StatusConflict)
	fakeAffected = 1
	if len(g.eventos) != 0 {
		t.Fatalf("un 409 no debe notificar: %+v", g.eventos)
	}
}
