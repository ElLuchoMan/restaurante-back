package reserva

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"restaurante/database"
	"restaurante/models"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Tests adicionales para aumentar cobertura de ReservaController

func TestReservaGetByDocumento_Success(t *testing.T) {
	// Guardar funciones originales
	origOrmNew := ormNew
	origQueryCliente := queryReservasByDocumentoCliente
	origQueryContacto := queryReservasByDocumentoContacto

	// Mock ORM
	ormNew = func() orm.Ormer { return nil }

	// Mock query cliente
	queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		*reservas = []models.Reserva{
			{PK_ID_RESERVA: 1, FECHA: time.Now(), HORA: time.Now(), PERSONAS: 4,
				CREATED_AT: time.Now(), UPDATED_AT: time.Now()},
		}
		return 1, nil
	}

	// Mock query contacto (no se debería llamar)
	queryReservasByDocumentoContacto = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		return 0, nil
	}

	// Restaurar al final
	defer func() {
		ormNew = origOrmNew
		queryReservasByDocumentoCliente = origQueryCliente
		queryReservasByDocumentoContacto = origQueryContacto
	}()

	// Request
	r := httptest.NewRequest("GET", "/reservas/documento?documento=123456789", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	// Controller
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	// Ejecutar
	c.GetByDocumento()

	// Verificar
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaGetByDocumento_ClienteError_FallbackToContacto(t *testing.T) {
	// Guardar funciones originales
	origOrmNew := ormNew
	origQueryCliente := queryReservasByDocumentoCliente
	origQueryContacto := queryReservasByDocumentoContacto

	// Mock ORM
	ormNew = func() orm.Ormer { return nil }

	// Mock query cliente - retorna 0 resultados
	queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		return 0, nil
	}

	// Mock query contacto - retorna resultados
	queryReservasByDocumentoContacto = func(o orm.Ormer, _ int64, _ time.Time, _ bool, reservas *[]models.Reserva) (int64, error) {
		*reservas = []models.Reserva{
			{PK_ID_RESERVA: 2, FECHA: time.Now(), HORA: time.Now(), PERSONAS: 2,
				CREATED_AT: time.Now(), UPDATED_AT: time.Now()},
		}
		return 1, nil
	}

	// Restaurar al final
	defer func() {
		ormNew = origOrmNew
		queryReservasByDocumentoCliente = origQueryCliente
		queryReservasByDocumentoContacto = origQueryContacto
	}()

	// Request
	r := httptest.NewRequest("GET", "/reservas/documento?documento=987654321", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	// Controller
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	// Ejecutar
	c.GetByDocumento()

	// Verificar
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaGetByDocumento_ErrorCliente(t *testing.T) {
	// Guardar funciones originales
	origOrmNew := ormNew
	origQueryCliente := queryReservasByDocumentoCliente

	// Mock ORM
	ormNew = func() orm.Ormer { return nil }

	// Mock query cliente con error
	queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		return 0, fmt.Errorf("database error")
	}

	// Restaurar al final
	defer func() {
		ormNew = origOrmNew
		queryReservasByDocumentoCliente = origQueryCliente
	}()

	// Request
	r := httptest.NewRequest("GET", "/reservas/documento?documento=123", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	// Controller
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	// Ejecutar
	c.GetByDocumento()

	// Verificar
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestReservaGetByDocumentoCliente_Success(t *testing.T) {
	// Guardar funciones originales
	origOrmNew := ormNew
	origQuery := queryReservasByDocumentoCliente

	// Mock ORM
	ormNew = func() orm.Ormer { return nil }

	// Mock query
	queryReservasByDocumentoCliente = func(o orm.Ormer, _ int64, _ time.Time, _ bool, reservas *[]models.Reserva) (int64, error) {
		*reservas = []models.Reserva{
			{PK_ID_RESERVA: 1, FECHA: time.Now(), HORA: time.Now(), PERSONAS: 4,
				CREATED_AT: time.Now(), UPDATED_AT: time.Now()},
		}
		return 1, nil
	}

	// Restaurar al final
	defer func() {
		ormNew = origOrmNew
		queryReservasByDocumentoCliente = origQuery
	}()

	// Request
	r := httptest.NewRequest("GET", "/reservas/cliente?documentoCliente=123456789", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	// Controller
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	// Ejecutar
	c.GetByDocumentoCliente()

	// Verificar
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected code 200, got %d", resp.Code)
	}
}

func TestReservaGetByDocumentoCliente_DBError(t *testing.T) {
	// Guardar funciones originales
	origOrmNew := ormNew
	origQuery := queryReservasByDocumentoCliente

	// Mock ORM
	ormNew = func() orm.Ormer { return nil }

	// Mock query con error
	queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		return 0, fmt.Errorf("database error")
	}

	// Restaurar al final
	defer func() {
		ormNew = origOrmNew
		queryReservasByDocumentoCliente = origQuery
	}()

	// Request
	r := httptest.NewRequest("GET", "/reservas/cliente?documentoCliente=123", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	// Controller
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	// Ejecutar
	c.GetByDocumentoCliente()

	// Verificar
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// Test helper para inicializar database.BogotaZone si es necesario
func init() {
	if database.BogotaZone == nil {
		loc, _ := time.LoadLocation("America/Bogota")
		database.BogotaZone = loc
	}
}
