package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Foo@Example.COM  ")
	if got != "foo@example.com" {
		t.Errorf("expected foo@example.com, got %s", got)
	}
}

func TestIsUniqueEmailErr(t *testing.T) {
	uniqueErr := errors.New("pq: duplicate key value violates unique constraint \"uq_cliente_correo\"")
	if !isUniqueEmailErr(uniqueErr) {
		t.Errorf("expected true for unique email error")
	}
	otherErr := errors.New("some other error")
	if isUniqueEmailErr(otherErr) {
		t.Errorf("expected false for non unique email error")
	}
}

func TestClienteGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener clientes") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
