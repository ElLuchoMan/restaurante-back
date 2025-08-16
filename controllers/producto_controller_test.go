package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

func setupProductoController(t *testing.T, data []byte) *web.Controller {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if data != nil {
		part, err := writer.CreateFormFile("IMAGEN", "test.png")
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(data); err != nil {
			t.Fatalf("write data: %v", err)
		}
	}
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, req)
	c := &web.Controller{}
	c.Ctx = ctx
	return c
}

func TestHandleImageUploadMissingFile(t *testing.T) {
	c := setupProductoController(t, nil)
	img, err := handleImageUpload(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if img != "" {
		t.Errorf("expected empty string, got %s", img)
	}
}

func TestHandleImageUploadTooLarge(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 1024*1024+1)
	c := setupProductoController(t, data)
	img, err := handleImageUpload(c)
	if err == nil || !strings.Contains(err.Error(), "1MB") {
		t.Fatalf("expected size error, got %v", err)
	}
	if img != "" {
		t.Errorf("expected empty string on error")
	}
}

func TestHandleImageUploadSuccess(t *testing.T) {
	c := setupProductoController(t, []byte("hello"))
	img, err := handleImageUpload(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == "" {
		t.Errorf("expected base64 image string")
	}
}

func int64Ptr(i int64) *int64 { return &i }

func TestValidateProducto(t *testing.T) {
	valid := &models.Producto{NOMBRE: "A", PRECIO: 10, ESTADO_PRODUCTO: "DISPONIBLE"}
	if err := validateProducto(valid); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []models.Producto{
		{PRECIO: 10, ESTADO_PRODUCTO: "DISPONIBLE"},                                      // missing name
		{NOMBRE: "B", PRECIO: 0, ESTADO_PRODUCTO: "DISPONIBLE"},                          // invalid price
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
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener productos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
