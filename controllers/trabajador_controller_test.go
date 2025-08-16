package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	context2 "github.com/beego/beego/v2/server/web/context"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	hashed, err := hashPassword("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hashed == "secret" {
		t.Errorf("hashed password should differ from original")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte("secret")); err != nil {
		t.Errorf("password does not match: %v", err)
	}
}

func TestValidateDates(t *testing.T) {
	ingreso := time.Now()
	retiroAntes := ingreso.Add(-time.Hour)
	if err := validateDates(&ingreso, &retiroAntes); err == nil {
		t.Errorf("expected error for retiro before ingreso")
	}
	retiroDespues := ingreso.Add(time.Hour)
	if err := validateDates(&ingreso, &retiroDespues); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := validateDates(nil, nil); err != nil {
		t.Errorf("expected nil for nil dates, got %v", err)
	}
}

func TestTrabajadorGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	w := httptest.NewRecorder()
	ctx := context2.NewContext()
	ctx.Reset(w, r)
	c := TrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener trabajadores") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
