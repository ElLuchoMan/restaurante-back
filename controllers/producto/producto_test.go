package producto

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/database"

	beecontext "github.com/beego/beego/v2/server/web/context"
)

var cols = []string{"pk_id_producto", "nombre", "calorias", "descripcion", "precio", "estado_producto", "imagen", "cantidad", "pk_id_subcategoria"}

type sqlLog struct{ stmts []string }

func (l *sqlLog) has(sub string) bool {
	for _, s := range l.stmts {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// found programa un producto existente (id 4, DISPONIBLE, subcategoría 2) y
// registra todas las sentencias de escritura; lastArgs guarda los argumentos del UPDATE de producto.
func found(estado string) (*sqlLog, *[]driver.NamedValue) {
	l := &sqlLog{}
	var upd []driver.NamedValue
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(4), "Bandeja", int64(850), "Plato", int64(25000), estado, []byte("img"), int64(10), int64(2)}), nil
	}
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		l.stmts = append(l.stmts, q)
		if strings.HasPrefix(q, `UPDATE "producto"`) {
			upd = a
		}
		return fakeResult{}, nil
	}
	return l, &upd
}

func boom() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
}

func execErrOn(prefix, msg string) {
	prev := fakeExec
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(q, prefix) {
			return nil, errors.New(msg)
		}
		if prev != nil {
			return prev(q, a)
		}
		return fakeResult{}, nil
	}
}

func run(t *testing.T, ctx *beecontext.Context, w *httptest.ResponseRecorder, f func(c *ProductoController), status int) string {
	t.Helper()
	c := &ProductoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func call(t *testing.T, method, target, body string, f func(c *ProductoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	return run(t, ctx, w, f, status)
}

func callForm(t *testing.T, method, target string, fields map[string]string, file []byte, f func(c *ProductoController), status int) string {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	if file != nil {
		fw, _ := mw.CreateFormFile("imagen", "x.png")
		_, _ = fw.Write(file)
	}
	_ = mw.Close()
	r := httptest.NewRequest(method, target, &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := beecontext.NewContext()
	ctx.Reset(w, r)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return run(t, ctx, w, f, status)
}

var (
	getAll  = func(c *ProductoController) { c.GetAll() }
	getByID = func(c *ProductoController) { c.GetById() }
	post    = func(c *ProductoController) { c.Post() }
	put     = func(c *ProductoController) { c.Put() }
	del     = func(c *ProductoController) { c.Delete() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	if b := call(t, http.MethodGet, "/productos", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	found("DISPONIBLE")
	b := call(t, http.MethodGet, "/productos?onlyActive=true", "", getAll, http.StatusOK)
	if strings.Contains(b, `"imagen"`) || !strings.Contains(b, `"subcategoriaId":2`) {
		t.Fatalf("sin includeImage no debe traer imagen: %s", b)
	}
	if b := call(t, http.MethodGet, "/productos?includeImage=true", "", getAll, http.StatusOK); !strings.Contains(b, `"imagen":"aW1n"`) {
		t.Fatalf("includeImage debe traer imagen: %s", b)
	}
	call(t, http.MethodGet, "/productos?includeImage=quizas", "", getAll, http.StatusBadRequest)
	call(t, http.MethodGet, "/productos?onlyActive=quizas", "", getAll, http.StatusBadRequest)
	boom()
	call(t, http.MethodGet, "/productos", "", getAll, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodGet, "/productos/search"+q, "", getByID, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/productos/search?id=4", "", getByID, http.StatusNotFound)
	found("DISPONIBLE")
	if b := call(t, http.MethodGet, "/productos/search?id=4", "", getByID, http.StatusOK); !strings.Contains(b, `"productoId":4`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	boom()
	call(t, http.MethodGet, "/productos/search?id=4", "", getByID, http.StatusInternalServerError)
}

const okBody = `{"nombre":"Bandeja","precio":25000,"estadoProducto":"disponible","cantidad":3,"subcategoriaId":2,"imagen":"data:image/png;base64,aW1n"}`

func TestPostJSON(t *testing.T) {
	defer resetFake()
	for _, b := range []string{"", "nojson", `{"nombre":"x","imagen":"%%%"}`, `{"nombre":"x","imagen":"data:xx"}`} {
		call(t, http.MethodPost, "/productos", b, post, http.StatusBadRequest)
	}
	for _, b := range []string{
		`{}`,
		`{"nombre":" ","precio":1,"estadoProducto":"DISPONIBLE"}`,
		`{"nombre":"x","precio":0,"estadoProducto":"DISPONIBLE"}`,
		`{"nombre":"x","precio":1,"estadoProducto":"OTRO"}`,
		`{"nombre":"x","precio":1,"estadoProducto":"DISPONIBLE","calorias":-1}`,
		`{"nombre":"x","precio":1,"estadoProducto":"DISPONIBLE","cantidad":-1}`,
	} {
		call(t, http.MethodPost, "/productos", b, post, http.StatusBadRequest)
	}

	l, _ := found("DISPONIBLE")
	fakeAffected = 0
	b := call(t, http.MethodPost, "/productos", okBody, post, http.StatusCreated)
	if !strings.Contains(b, `"productoId":7`) || !strings.Contains(b, `"estadoProducto":"DISPONIBLE"`) || !strings.Contains(b, `"imagen":"aW1n"`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	if !l.has(`INSERT INTO "precio_producto_hist"`) {
		t.Fatalf("debe registrar el historial: %v", l.stmts)
	}
	// Si ya hay precio hoy, solo se actualiza.
	l, _ = found("DISPONIBLE")
	fakeAffected = 1
	call(t, http.MethodPost, "/productos", `{"nombre":"x","precio":5,"estadoProducto":"DISPONIBLE"}`, post, http.StatusCreated)
	if l.has(`INSERT INTO "precio_producto_hist"`) {
		t.Fatalf("no debe insertar historial duplicado: %v", l.stmts)
	}

	execErrOn("INSERT INTO", "duplicate key value (23505)")
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusConflict)
	resetFake()
	execErrOn("INSERT INTO", "violates foreign key constraint (23503)")
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusBadRequest)
	resetFake()
	execErrOn("INSERT INTO", "boom")
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusInternalServerError)
	resetFake()
	execErrOn("UPDATE precio_producto_hist", "boom")
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusInternalServerError)
	resetFake()
	fakeAffected = 0
	execErrOn(`INSERT INTO "precio_producto_hist"`, "boom")
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusInternalServerError)
}

func TestPostMultipart(t *testing.T) {
	defer resetFake()
	fields := map[string]string{"nombre": "Bandeja", "descripcion": "Plato", "precio": "25000", "estadoProducto": "disponible", "cantidad": "3", "calorias": "850", "subcategoriaId": "2"}
	b := callForm(t, http.MethodPost, "/productos", fields, []byte("img"), post, http.StatusCreated)
	if !strings.Contains(b, `"imagen":"aW1n"`) || !strings.Contains(b, `"calorias":850`) || !strings.Contains(b, `"subcategoriaId":2`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	callForm(t, http.MethodPost, "/productos", map[string]string{"nombre": "x", "precio": "abc"}, nil, post, http.StatusBadRequest)
	callForm(t, http.MethodPost, "/productos", map[string]string{"nombre": "x"}, nil, post, http.StatusBadRequest) // precio 0

	prev := readAll
	readAll = func(io.Reader) ([]byte, error) { return nil, errors.New("lectura") }
	defer func() { readAll = prev }()
	callForm(t, http.MethodPost, "/productos", fields, []byte("img"), post, http.StatusBadRequest)
}

func TestPut(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPut, "/productos", `{}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/productos?id=4", `{}`, put, http.StatusNotFound)
	boom()
	call(t, http.MethodPut, "/productos?id=4", `{}`, put, http.StatusInternalServerError)

	l, upd := found("DISPONIBLE")
	// PUT sin cambios: 200, sin UPDATE y subcategoriaId conservada.
	b := call(t, http.MethodPut, "/productos?id=4", `{}`, put, http.StatusOK)
	if !strings.Contains(b, `"subcategoriaId":2`) || len(l.stmts) != 0 {
		t.Fatalf("sin cambios debe responder 200 sin escribir: %s %v", b, l.stmts)
	}
	// El mismo JSON del GET (valores idénticos) también es 200 sin escribir.
	same := `{"productoId":4,"nombre":"Bandeja","calorias":850,"descripcion":"Plato","precio":25000,"estadoProducto":"DISPONIBLE","cantidad":10,"subcategoriaId":2}`
	b = call(t, http.MethodPut, "/productos?id=4", strings.Replace(same, `"productoId":4,`, "", 1), put, http.StatusOK)
	if len(l.stmts) != 0 {
		t.Fatalf("no debe escribir: %v", l.stmts)
	}

	for _, body := range []string{
		"", "nojson",
		`{"nombre":null}`, `{"precio":null}`, `{"estadoProducto":null}`, `{"cantidad":null}`,
		`{"subcategoriaId":0}`, `{"subcategoriaId":-3}`,
		`{"imagen":"%%%"}`,
		`{"nombre":" "}`, `{"precio":0}`, `{"estadoProducto":"OTRO"}`, `{"calorias":-5}`, `{"cantidad":-1}`,
		`{"precio":"caro"}`,
	} {
		call(t, http.MethodPut, "/productos?id=4", body, put, http.StatusBadRequest)
	}

	// Cambio de precio: UPDATE de producto y del historial del día.
	l, upd = found("DISPONIBLE")
	b = call(t, http.MethodPut, "/productos?id=4", `{"precio":30000,"estadoProducto":"no_disponible"}`, put, http.StatusOK)
	if !strings.Contains(b, `"precio":30000`) || !strings.Contains(b, `"subcategoriaId":2`) || !l.has(`UPDATE precio_producto_hist`) {
		t.Fatalf("cuerpo/sentencias inesperados %s %v", b, l.stmts)
	}
	_ = upd

	// Cambio de precio sin fila de historial hoy: se inserta.
	l, _ = found("DISPONIBLE")
	fakeAffected = 0
	call(t, http.MethodPut, "/productos?id=4", `{"precio":30001}`, put, http.StatusOK)
	if !l.has(`INSERT INTO "precio_producto_hist"`) {
		t.Fatalf("debe insertar historial: %v", l.stmts)
	}
	fakeAffected = 1

	// Sin cambio de precio no toca el historial; asigna todos los campos.
	l, upd = found("DISPONIBLE")
	b = call(t, http.MethodPut, "/productos?id=4", `{"nombre":"Nueva","calorias":900,"descripcion":"Otra","cantidad":1,"imagen":"data:image/png;base64,aW1n","subcategoriaId":5}`, put, http.StatusOK)
	if l.has("precio_producto_hist") || !strings.Contains(b, `"subcategoriaId":5`) || !strings.Contains(b, `"calorias":900`) {
		t.Fatalf("cuerpo/sentencias inesperados %s %v", b, l.stmts)
	}

	// null limpia los campos anulables.
	l, upd = found("DISPONIBLE")
	b = call(t, http.MethodPut, "/productos?id=4", `{"calorias":null,"descripcion":null,"imagen":null,"subcategoriaId":null}`, put, http.StatusOK)
	if !strings.Contains(b, `"subcategoriaId":null`) || !strings.Contains(b, `"calorias":null`) || strings.Contains(b, `"descripcion"`) {
		t.Fatalf("null debe limpiar: %s", b)
	}
	if got := (*upd)[len(*upd)-2].Value; got != nil {
		t.Fatalf("subcategoria debe enviarse NULL, fue %v", got)
	}

	found("DISPONIBLE")
	execErrOn(`UPDATE "producto"`, "duplicate key value (23505)")
	call(t, http.MethodPut, "/productos?id=4", `{"nombre":"Otro"}`, put, http.StatusConflict)
	execErrOn(`UPDATE "producto"`, "violates foreign key constraint (23503)")
	call(t, http.MethodPut, "/productos?id=4", `{"subcategoriaId":99}`, put, http.StatusBadRequest)
	resetFake()
	found("DISPONIBLE")
	execErrOn(`UPDATE "producto"`, "boom")
	call(t, http.MethodPut, "/productos?id=4", `{"nombre":"Otro"}`, put, http.StatusInternalServerError)
	found("DISPONIBLE")
	execErrOn(`UPDATE precio_producto_hist`, "boom")
	call(t, http.MethodPut, "/productos?id=4", `{"precio":1}`, put, http.StatusInternalServerError)
}

func TestPutMultipart(t *testing.T) {
	defer resetFake()
	found("DISPONIBLE")
	b := callForm(t, http.MethodPut, "/productos?id=4", map[string]string{
		"nombre": "Nueva", "descripcion": "D", "estadoProducto": "no_disponible", "precio": "100", "cantidad": "1", "calorias": "5", "subcategoriaId": "9",
	}, []byte("zzz"), put, http.StatusOK)
	if !strings.Contains(b, `"nombre":"Nueva"`) || !strings.Contains(b, `"subcategoriaId":9`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	// Campos vacíos o ausentes se conservan.
	b = callForm(t, http.MethodPut, "/productos?id=4", map[string]string{"nombre": ""}, nil, put, http.StatusOK)
	if !strings.Contains(b, `"nombre":"Bandeja"`) {
		t.Fatalf("debe conservar: %s", b)
	}
	callForm(t, http.MethodPut, "/productos?id=4", map[string]string{"precio": "abc"}, nil, put, http.StatusBadRequest)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodDelete, "/productos"+q, "", del, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/productos?id=4", "", del, http.StatusNotFound)
	boom()
	call(t, http.MethodDelete, "/productos?id=4", "", del, http.StatusInternalServerError)

	l, _ := found("DISPONIBLE")
	call(t, http.MethodDelete, "/productos?id=4", "", del, http.StatusOK)
	if !l.has(`UPDATE "producto"`) {
		t.Fatalf("debe actualizar el estado: %v", l.stmts)
	}
	execErrOn(`UPDATE "producto"`, "boom")
	call(t, http.MethodDelete, "/productos?id=4", "", del, http.StatusInternalServerError)

	l, _ = found("NO_DISPONIBLE")
	call(t, http.MethodDelete, "/productos?id=4", "", del, http.StatusBadRequest)
	if len(l.stmts) != 0 {
		t.Fatalf("no debe escribir: %v", l.stmts)
	}
}

func TestRegistrarPrecioConZonaBogota(t *testing.T) {
	defer resetFake()
	prev := database.BogotaZone
	database.BogotaZone = time.FixedZone("UTC-5", -5*60*60)
	defer func() { database.BogotaZone = prev }()
	call(t, http.MethodPost, "/productos", okBody, post, http.StatusCreated)
}
