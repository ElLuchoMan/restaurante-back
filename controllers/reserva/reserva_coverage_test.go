package reserva

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func newReservaCtrl(method, url, body string) (*ReservaController, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &ReservaController{}
	c.Init(ctx, "ReservaController", "x", nil)
	return c, w
}

func TestCreateOrFindContacto_DocumentoExistente(t *testing.T) {
	orig := queryReservaContactoByDocumento
	t.Cleanup(func() { queryReservaContactoByDocumento = orig })
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		rc.PKIDContacto = 7
		rc.NombreCompleto = "Existente"
		return nil
	}

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoContacto": float64(100)})
	if err != nil || got == nil || got.PKIDContacto != 7 || got.NombreCompleto != "Existente" {
		t.Fatalf("resultado inesperado: %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_DocumentoErrorConsulta(t *testing.T) {
	orig := queryReservaContactoByDocumento
	t.Cleanup(func() { queryReservaContactoByDocumento = orig })
	boom := errors.New("db caida")
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error { return boom }

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoContacto": float64(100)})
	if got != nil || !errors.Is(err, boom) {
		t.Fatalf("esperado error de consulta, obtenido %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_DocumentoInsertError(t *testing.T) {
	origQ, origI := queryReservaContactoByDocumento, insertReservaContacto
	t.Cleanup(func() { queryReservaContactoByDocumento = origQ; insertReservaContacto = origI })
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}
	boom := errors.New("insert fallo")
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) { return 0, boom }

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{
		"documentoContacto": float64(100), "nombreCompleto": "Nuevo", "telefono": "3001112222",
	})
	if got != nil || !errors.Is(err, boom) {
		t.Fatalf("esperado error de insercion, obtenido %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_ClienteExistente(t *testing.T) {
	orig := queryReservaContactoByCliente
	t.Cleanup(func() { queryReservaContactoByCliente = orig })
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		rc.PKIDContacto = 9
		return nil
	}

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoCliente": float64(55)})
	if err != nil || got == nil || got.PKIDContacto != 9 {
		t.Fatalf("resultado inesperado: %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_ClienteErrorConsulta(t *testing.T) {
	orig := queryReservaContactoByCliente
	t.Cleanup(func() { queryReservaContactoByCliente = orig })
	boom := errors.New("db caida")
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error { return boom }

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoCliente": float64(55)})
	if got != nil || !errors.Is(err, boom) {
		t.Fatalf("esperado error de consulta, obtenido %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_ClienteNoEncontrado(t *testing.T) {
	origQ, origR := queryReservaContactoByCliente, readCliente
	t.Cleanup(func() { queryReservaContactoByCliente = origQ; readCliente = origR })
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error { return orm.ErrNoRows }

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoCliente": float64(55)})
	if got != nil || err == nil || !strings.Contains(err.Error(), "cliente no encontrado") || !errors.Is(err, orm.ErrNoRows) {
		t.Fatalf("esperado cliente no encontrado, obtenido %+v, %v", got, err)
	}
}

func stubClienteNuevo(t *testing.T, insertErr error) {
	t.Helper()
	origQ, origR, origI := queryReservaContactoByCliente, readCliente, insertReservaContacto
	t.Cleanup(func() {
		queryReservaContactoByCliente = origQ
		readCliente = origR
		insertReservaContacto = origI
	})
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error {
		c.NOMBRE, c.APELLIDO, c.TELEFONO = "Ana", "Perez", "3000000000"
		return nil
	}
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		if insertErr != nil {
			return 0, insertErr
		}
		return 21, nil
	}
}

func TestCreateOrFindContacto_ClienteInsertError(t *testing.T) {
	boom := errors.New("insert fallo")
	stubClienteNuevo(t, boom)

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoCliente": float64(55)})
	if got != nil || !errors.Is(err, boom) {
		t.Fatalf("esperado error de insercion, obtenido %+v, %v", got, err)
	}
}

func TestCreateOrFindContacto_ClienteNuevoOK(t *testing.T) {
	stubClienteNuevo(t, nil)

	got, err := createOrFindReservaContacto(nil, map[string]interface{}{"documentoCliente": float64(55)})
	if err != nil || got == nil {
		t.Fatalf("error inesperado: %+v, %v", got, err)
	}
	if got.PKIDContacto != 21 || got.NombreCompleto != "Ana Perez" || got.Telefono == nil || *got.Telefono != "3000000000" {
		t.Fatalf("contacto inesperado: %+v", got)
	}
}

func TestReservaPut_IDCeroSinErrorDeParseo(t *testing.T) {
	c, w := newReservaCtrl(http.MethodPut, "/reservas?id=0", "{}")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtenido %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "distinto de cero") {
		t.Errorf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestReservaPut_IDAusente(t *testing.T) {
	c, w := newReservaCtrl(http.MethodPut, "/reservas", "{}")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtenido %d", w.Code)
	}
}

func stubPutBase(t *testing.T) {
	t.Helper()
	origRead, origUpd := readReserva, updateReserva
	t.Cleanup(func() { readReserva = origRead; updateReserva = origUpd })
	readReserva = func(o orm.Ormer, r *models.Reserva) error { return nil }
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) { return 1, nil }
}

func TestReservaPut_ContactoPorDocumentoContactoOK(t *testing.T) {
	stubPutBase(t)
	orig := queryReservaContactoByDocumento
	t.Cleanup(func() { queryReservaContactoByDocumento = orig })
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		rc.PKIDContacto = 33
		return nil
	}
	var updated *models.Reserva
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		updated = r
		return 1, nil
	}

	c, w := newReservaCtrl(http.MethodPut, "/reservas?id=1", `{"documentoContacto":123}`)
	c.Put()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Reserva actualizada correctamente") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
	if updated == nil || updated.PK_ID_CONTACTO == nil || updated.PK_ID_CONTACTO.PKIDContacto != 33 {
		t.Fatalf("contacto no asignado: %+v", updated)
	}
}

func TestReservaPut_ContactoPorDocumentoClienteOK(t *testing.T) {
	stubPutBase(t)
	orig := queryReservaContactoByCliente
	t.Cleanup(func() { queryReservaContactoByCliente = orig })
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		rc.PKIDContacto = 44
		return nil
	}
	var updated *models.Reserva
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		updated = r
		return 1, nil
	}

	c, w := newReservaCtrl(http.MethodPut, "/reservas?id=1", `{"documentoCliente":456}`)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d: %s", w.Code, w.Body.String())
	}
	if updated == nil || updated.PK_ID_CONTACTO == nil || updated.PK_ID_CONTACTO.PKIDContacto != 44 {
		t.Fatalf("contacto no asignado: %+v", updated)
	}
}

func TestReservaPut_ContactoPorDocumentoClienteError(t *testing.T) {
	stubPutBase(t)
	orig := queryReservaContactoByCliente
	t.Cleanup(func() { queryReservaContactoByCliente = orig })
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		return errors.New("db caida")
	}

	c, w := newReservaCtrl(http.MethodPut, "/reservas?id=1", `{"documentoCliente":456}`)
	c.Put()
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "Error al procesar contacto") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}

func TestReservaGetByDocumento_ErrorConsultaCliente(t *testing.T) {
	orig := queryReservasByDocumentoCliente
	t.Cleanup(func() { queryReservasByDocumentoCliente = orig })
	queryReservasByDocumentoCliente = func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		return 0, errors.New("fallo cliente")
	}

	c, w := newReservaCtrl(http.MethodGet, "/reservas/documento?documento=10", "")
	c.GetByDocumento()
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "fallo cliente") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}

func TestReservaGetByDocumento_ErrorConsultaContacto(t *testing.T) {
	origC, origT := queryReservasByDocumentoCliente, queryReservasByDocumentoContacto
	t.Cleanup(func() { queryReservasByDocumentoCliente = origC; queryReservasByDocumentoContacto = origT })
	queryReservasByDocumentoCliente = func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		return 0, nil
	}
	queryReservasByDocumentoContacto = func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		return 0, errors.New("fallo contacto")
	}

	c, w := newReservaCtrl(http.MethodGet, "/reservas/documento?documento=10&fecha=2024-05-01", "")
	c.GetByDocumento()
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "fallo contacto") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}

func TestReservaGetByDocumentoCliente_ConFechaYError(t *testing.T) {
	orig := queryReservasByDocumentoCliente
	t.Cleanup(func() { queryReservasByDocumentoCliente = orig })
	var gotUseFecha bool
	var gotFecha time.Time
	queryReservasByDocumentoCliente = func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		gotUseFecha, gotFecha = useFecha, f
		return 0, errors.New("fallo db")
	}

	c, w := newReservaCtrl(http.MethodGet, "/reservas/cliente?documentoCliente=10&fecha=2024-05-01", "")
	c.GetByDocumentoCliente()
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "fallo db") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
	if !gotUseFecha || gotFecha.Format("2006-01-02") != "2024-05-01" {
		t.Fatalf("la fecha no se propago: %v %v", gotUseFecha, gotFecha)
	}
}

func TestReservaPut_ReservaNoEncontrada(t *testing.T) {
	orig := readReserva
	t.Cleanup(func() { readReserva = orig })
	readReserva = func(o orm.Ormer, r *models.Reserva) error { return orm.ErrNoRows }

	c, w := newReservaCtrl(http.MethodPut, "/reservas?id=5", "{}")
	c.Put()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Reserva no encontrada") || !strings.Contains(w.Body.String(), "404") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}

func TestReservaGetByDocumento_SinResultados(t *testing.T) {
	origC, origT := queryReservasByDocumentoCliente, queryReservasByDocumentoContacto
	t.Cleanup(func() { queryReservasByDocumentoCliente = origC; queryReservasByDocumentoContacto = origT })
	empty := func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		return 0, nil
	}
	queryReservasByDocumentoCliente = empty
	queryReservasByDocumentoContacto = empty

	c, w := newReservaCtrl(http.MethodGet, "/reservas/documento?documento=10", "")
	c.GetByDocumento()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "No se encontraron reservas para este documento") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}

func TestReservaGetByDocumentoCliente_SinResultados(t *testing.T) {
	orig := queryReservasByDocumentoCliente
	t.Cleanup(func() { queryReservasByDocumentoCliente = orig })
	queryReservasByDocumentoCliente = func(o orm.Ormer, doc int64, f time.Time, useFecha bool, rs *[]models.Reserva) (int64, error) {
		return 0, nil
	}

	c, w := newReservaCtrl(http.MethodGet, "/reservas/cliente?documentoCliente=10", "")
	c.GetByDocumentoCliente()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "No se encontraron reservas para este cliente") {
		t.Fatalf("respuesta inesperada: %d %s", w.Code, w.Body.String())
	}
}
