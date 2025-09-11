package producto

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func int64Ptr(i int64) *int64 { return &i }

func TestValidateProducto(t *testing.T) {
	valid := &models.Producto{NOMBRE: "A", PRECIO: 10, ESTADO_PRODUCTO: "DISPONIBLE"}
	if err := validateProducto(valid); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []models.Producto{
		{PRECIO: 10, ESTADO_PRODUCTO: "DISPONIBLE"},                                      // missing name
		{NOMBRE: "B", PRECIO: 0, ESTADO_PRODUCTO: "DISPONIBLE"},                          // zero price
		{NOMBRE: "B", PRECIO: -5, ESTADO_PRODUCTO: "DISPONIBLE"},                         // negative price
		{NOMBRE: "B", PRECIO: 10, CALORIAS: int64Ptr(-1), ESTADO_PRODUCTO: "DISPONIBLE"}, // negative calories
		{NOMBRE: "B", PRECIO: 10, ESTADO_PRODUCTO: "OTRO"},                               // invalid estado
	}

	for _, p := range tests {
		if err := validateProducto(&p); err == nil {
			t.Errorf("expected error for producto %+v", p)
		}
	}
}

func TestProductoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryProductosAll
	queryProductosAll = func(o orm.Ormer, onlyActive bool, productos *[]models.Producto) (int64, error) {
		return 0, errors.New("db error")
	}
	t.Cleanup(func() { queryProductosAll = saved })

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener productos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoGetAllSuccessFiltersImage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/productos?includeImage=false&onlyActive=true", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryProductosAll
	queryProductosAll = func(o orm.Ormer, onlyActive bool, productos *[]models.Producto) (int64, error) {
		*productos = []models.Producto{{IMAGEN: "img", ESTADO_PRODUCTO: models.EstadoProductoDisponible}}
		return 1, nil
	}
	t.Cleanup(func() { queryProductosAll = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "img") {
		t.Errorf("expected image stripped when includeImage=false: %s", w.Body.String())
	}
}

func TestProductoGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoGetByIdNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return orm.ErrNoRows }
	t.Cleanup(func() { readProductoFn = savedRead })

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "producto no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoGetByIdOtherError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return errors.New("boom") }
	t.Cleanup(func() { readProductoFn = savedRead })

	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "error al buscar el producto") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPostMissingNombre(t *testing.T) {
	body := bytes.NewBufferString(`{"estadoProducto":"DISPONIBLE","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "nombre") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPostSuccess(t *testing.T) {
	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"DISPONIBLE","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	savedInsertHist := insertPrecioHistFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { insertProductoFn = savedInsert; insertPrecioHistFn = savedInsertHist })

	c.Post()
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 status in multipart post")
	}
}

func TestProductoPostMultipartSuccess(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("nombre", "A")
	_ = mw.WriteField("descripcion", "desc")
	_ = mw.WriteField("precio", "10")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.WriteField("cantidad", "2")
	_ = mw.WriteField("calorias", "100")
	_ = mw.WriteField("subcategoriaId", "1")
	fw, _ := mw.CreateFormFile("imagen", "a.txt")
	_, _ = io.Copy(fw, strings.NewReader("imgdata"))
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/productos", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)

	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	savedInsertHist := insertPrecioHistFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { insertProductoFn = savedInsert; insertPrecioHistFn = savedInsertHist })

	c.Post()
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 status in multipart post")
	}
}

func TestProductoPutMultipartSuccessWithPriceChange(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("nombre", "B")
	_ = mw.WriteField("descripcion", "desc2")
	_ = mw.WriteField("precio", "11")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.WriteField("cantidad", "3")
	_ = mw.WriteField("calorias", "150")
	_ = mw.WriteField("subcategoriaId", "2")
	fw, _ := mw.CreateFormFile("imagen", "b.txt")
	_, _ = io.Copy(fw, strings.NewReader("imgdata2"))
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)

	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	origInsertHist := insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origInsertHist })

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPostInsertError(t *testing.T) {
	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"DISPONIBLE","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 0, errors.New("db") }
	t.Cleanup(func() { insertProductoFn = savedInsert })

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoPostInvalidEstado(t *testing.T) {
	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"OTRO","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPostJSONWithImageBase64(t *testing.T) {
	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"DISPONIBLE","precio":10,"imagen":"aW1nZGF0YQ=="}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	savedInsertHist := insertPrecioHistFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { insertProductoFn = savedInsert; insertPrecioHistFn = savedInsertHist })

	c.Post()
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestProductoPostMultipartMissingNombre(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("precio", "10")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/productos", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProductoPostPrecioHistInsertError(t *testing.T) {
	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"DISPONIBLE","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	savedInsertHist := insertPrecioHistFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 0, errors.New("hist fail") }
	t.Cleanup(func() { insertProductoFn = savedInsert; insertPrecioHistFn = savedInsertHist })

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/productos", strings.NewReader("{invalid"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{invalid")
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPostMultipartInvalidNumericFields(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("nombre", "A")
	_ = mw.WriteField("descripcion", "desc")
	_ = mw.WriteField("precio", "abc")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.WriteField("cantidad", "xyz")
	_ = mw.WriteField("calorias", "abc")
	_ = mw.WriteField("subcategoriaId", "xyz")
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/productos", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPostRawExecError(t *testing.T) {
	savedOrm := ormNewProducto
	ormNewProducto = func() orm.Ormer { return newExecErrOrmer("mockErrExecPost") }
	t.Cleanup(func() { ormNewProducto = savedOrm })

	body := bytes.NewBufferString(`{"nombre":"A","estadoProducto":"DISPONIBLE","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body.Bytes()
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedInsert := insertProductoFn
	savedHist := insertPrecioHistFn
	insertProductoFn = func(o orm.Ormer, p *models.Producto) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { insertProductoFn = savedInsert; insertPrecioHistFn = savedHist })

	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

// PUT cases
func TestProductoPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPutJSONInvalido(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return nil }
	t.Cleanup(func() { readProductoFn = savedRead })

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoPutNoCambios(t *testing.T) {
	body := `{"nombre":"A","precio":10,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusNotModified {
		t.Fatalf("expected status 304, got %d", w.Code)
	}
}

func TestProductoPutUpdateError(t *testing.T) {
	body := `{"nombre":"B","precio":11,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 0, errors.New("db") }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate })

	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoPutHistorialError(t *testing.T) {
	body := `{"nombre":"B","precio":11,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	origInsertHist := insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 0, errors.New("db") }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origInsertHist })

	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoPutSuccess(t *testing.T) {
	body := `{"nombre":"B","precio":11,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	origInsertHist := insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origInsertHist })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoPutNotFound(t *testing.T) {
	body := `{"nombre":"A","precio":10,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=99", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return orm.ErrNoRows }
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Producto no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPutPriceChangeHistSuccess_JSON(t *testing.T) {
	body := `{"nombre":"B","precio":12,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead, origUpdate, origHist := readProductoFn, updateProductoFn, insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origHist })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Producto actualizado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPutPriceChangeHistError_JSON(t *testing.T) {
	body := `{"nombre":"B","precio":12,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead, origUpdate, origHist := readProductoFn, updateProductoFn, insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 0, errors.New("hist error") }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origHist })

	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoPutRawExecError(t *testing.T) {
	savedOrm := ormNewProducto
	ormNewProducto = func() orm.Ormer { return newExecErrOrmer("mockErrExecPut") }
	t.Cleanup(func() { ormNewProducto = savedOrm })

	body := `{"nombre":"B","precio":11,"estadoProducto":"DISPONIBLE"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	origHist := insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origHist })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoPutValidationError_NegativeCalories(t *testing.T) {
	body := `{"nombre":"A","precio":10,"estadoProducto":"DISPONIBLE","calorias":-1}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProductoPutInvalidEstadoJSON(t *testing.T) {
	body := `{"nombre":"A","precio":10,"estadoProducto":"INVALIDO"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProductoPutJSONWithImage(t *testing.T) {
	body := `{"nombre":"B","precio":12,"estadoProducto":"DISPONIBLE","imagen":"aW1n"}`
	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead, origUpdate, origHist := readProductoFn, updateProductoFn, insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origHist })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoPutMultipartNoPriceChange(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("nombre", "A")
	_ = mw.WriteField("descripcion", "desc")
	_ = mw.WriteField("precio", "10")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.WriteField("cantidad", "2")
	_ = mw.WriteField("calorias", "100")
	_ = mw.WriteField("subcategoriaId", "1")
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Request.ParseMultipartForm(32 << 20)

	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	origUpdate := updateProductoFn
	origInsertHist := insertPrecioHistFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) {
		t.Fatalf("history should not be inserted when price does not change")
		return 0, nil
	}
	t.Cleanup(func() { readProductoFn = origRead; updateProductoFn = origUpdate; insertPrecioHistFn = origInsertHist })

	c.Put()
	if w.Code != http.StatusOK && w.Code != http.StatusNotModified {
		t.Fatalf("expected status 200 or 304, got %d", w.Code)
	}
}

func TestProductoPutMultipartMissingFields(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusNotModified {
		t.Fatalf("expected status 304, got %d", w.Code)
	}
}

func TestProductoPutMultipartInvalidNumeric(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("nombre", "A")
	_ = mw.WriteField("precio", "abc")
	_ = mw.WriteField("estadoProducto", "DISPONIBLE")
	_ = mw.WriteField("cantidad", "xyz")
	_ = mw.WriteField("calorias", "abc")
	_ = mw.WriteField("subcategoriaId", "xyz")
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPut, "/productos?id=1", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "A"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = origRead })

	c.Put()
	if w.Code != http.StatusNotModified {
		t.Fatalf("expected status 304, got %d", w.Code)
	}
}

// DELETE cases
func TestProductoDeleteAlreadyDisabled(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.ESTADO_PRODUCTO = models.EstadoProductoNoDisponible
		return nil
	}
	t.Cleanup(func() { readProductoFn = savedRead })

	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoDeleteUpdateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	savedUpdate := updateProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 0, errors.New("db") }
	t.Cleanup(func() { readProductoFn = savedRead; updateProductoFn = savedUpdate })

	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestProductoDeleteSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	savedUpdate := updateProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		return nil
	}
	updateProductoFn = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return 1, nil }
	t.Cleanup(func() { readProductoFn = savedRead; updateProductoFn = savedUpdate })

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos?id=999", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedGet := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return orm.ErrNoRows }
	t.Cleanup(func() { readProductoFn = savedGet })

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "producto no encontrado") && !strings.Contains(strings.ToLower(w.Body.String()), "no encontrado") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoDeleteReadError(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error { return errors.New("db err") }
	t.Cleanup(func() { readProductoFn = savedRead })

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "error al buscar el producto") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoGetByIdSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	savedRead := readProductoFn
	readProductoFn = func(o orm.Ormer, p *models.Producto) error {
		p.NOMBRE = "P"
		p.PRECIO = 10
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
		p.IMAGEN = "img"
		return nil
	}
	t.Cleanup(func() { readProductoFn = savedRead })

	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProductoGetAllIncludeImageTrue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/productos?includeImage=true", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryProductosAll
	queryProductosAll = func(o orm.Ormer, onlyActive bool, productos *[]models.Producto) (int64, error) {
		*productos = []models.Producto{{IMAGEN: "img", ESTADO_PRODUCTO: models.EstadoProductoDisponible}}
		return 1, nil
	}
	t.Cleanup(func() { queryProductosAll = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "\"imagen\":\"") {
		t.Errorf("expected image field present: %s", w.Body.String())
	}
}
