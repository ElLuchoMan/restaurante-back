package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
)

func newCtl() (*web.Controller, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c := &web.Controller{}
	c.Init(beecontext.NewContext(), "T", "A", nil)
	c.Ctx.Reset(w, httptest.NewRequest(http.MethodGet, "/", nil))
	return c, w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var r models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("json: %v (%s)", err, w.Body.String())
	}
	return r
}

func TestSendYFail(t *testing.T) {
	c, w := newCtl()
	Send(c, http.StatusCreated, "ok", map[string]int{"a": 1})
	if w.Code != http.StatusCreated || decode(t, w).Code != http.StatusCreated {
		t.Fatalf("status/code incoherentes: %d %s", w.Code, w.Body.String())
	}

	c, w = newCtl()
	Fail(c, http.StatusNotFound, "no existe", errors.New("boom"))
	r := decode(t, w)
	if w.Code != http.StatusNotFound || r.Code != http.StatusNotFound || r.Cause != "boom" {
		t.Fatalf("respuesta inesperada: %d %+v", w.Code, r)
	}

	c, w = newCtl()
	Fail(c, http.StatusBadRequest, "malo", nil)
	if r := decode(t, w); r.Cause != "" || w.Code != http.StatusBadRequest {
		t.Fatalf("sin cause esperado: %+v", r)
	}
}

func TestList(t *testing.T) {
	var nilSlice []int
	b, _ := json.Marshal(List(nilSlice))
	if string(b) != "[]" {
		t.Fatalf("nil debe serializar []: %s", b)
	}
	if got := List([]int{1}); len(got) != 1 {
		t.Fatalf("no debe alterar slices con datos")
	}
}

type dst struct {
	A string  `json:"a"`
	B *string `json:"b"`
	C int     `json:"c"`
}

func TestDecodeMerge(t *testing.T) {
	b := "viejo"
	d := dst{A: "x", B: &b, C: 3}
	if err := DecodeMerge([]byte(`{"c":9}`), &d, "b"); err != nil || d.A != "x" || d.B == nil || d.C != 9 {
		t.Fatalf("los ausentes deben conservarse: %+v %v", d, err)
	}
	if err := DecodeMerge([]byte(`{"b":null}`), &d, "b"); err != nil || d.B != nil {
		t.Fatalf("null permitido debe limpiar: %+v %v", d, err)
	}
	if err := DecodeMerge([]byte(`{"a":null,"c":null}`), &d, "b"); err == nil {
		t.Fatalf("null en campo no anulable debe fallar")
	}
	if err := DecodeMerge([]byte(`{`), &d); err == nil {
		t.Fatalf("json inválido debe fallar")
	}
	if err := DecodeMerge([]byte(`{"c":"texto"}`), &d); err == nil {
		t.Fatalf("tipo inválido debe fallar")
	}
}

func TestPresentEIsNull(t *testing.T) {
	p, err := Present([]byte(`{"a":1,"b":null}`))
	if err != nil || !p["a"] || !p["b"] || p["c"] {
		t.Fatalf("Present: %v %v", p, err)
	}
	if _, err := Present([]byte(`[`)); err == nil {
		t.Fatalf("debe fallar con json inválido")
	}
	if !IsNull([]byte(`{"b": null}`), "b") || IsNull([]byte(`{"b":1}`), "b") || IsNull([]byte(`{}`), "b") || IsNull([]byte(`x`), "b") {
		t.Fatalf("IsNull incorrecto")
	}
}

func TestPositiveInt64Param(t *testing.T) {
	cases := map[string]bool{"": false, "abc": false, "0": false, "-3": false, "7": true, " 9 ": true}
	for raw, ok := range cases {
		c, _ := newCtl()
		c.Ctx.Request.URL.RawQuery = "id=" + url.QueryEscape(raw)
		v, err := PositiveInt64Param(c, "id")
		if ok && (err != nil || v <= 0) {
			t.Fatalf("%q debía ser válido: %d %v", raw, v, err)
		}
		if !ok && (err == nil || v != 0) {
			t.Fatalf("%q debía fallar: %d %v", raw, v, err)
		}
	}
}

func TestIsPGConflict(t *testing.T) {
	for _, m := range []string{"pq: duplicate key value (23505)", "ERROR: 23503", "violates foreign key constraint", "duplicate key"} {
		if !IsPGConflict(errors.New(m)) {
			t.Fatalf("%q debía ser conflicto", m)
		}
	}
	if IsPGConflict(nil) || IsPGConflict(errors.New("otro")) {
		t.Fatalf("no debía ser conflicto")
	}
	if ErrNotFound == nil {
		t.Fatalf("centinela nulo")
	}
}
