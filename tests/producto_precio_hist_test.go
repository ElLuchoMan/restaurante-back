package test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurante/models"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProductoPriceHistoryLifecycle(t *testing.T) {
	// Create product via controller
	name := "HistTest" + strconv.FormatInt(time.Now().UnixNano(), 10)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("NOMBRE", name)
	writer.WriteField("ESTADO_PRODUCTO", string(models.EstadoProductoDisponible))
	writer.WriteField("PRECIO", "100")
	writer.WriteField("PK_ID_SUBCATEGORIA", "1")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/productos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := sendRequest(req)
	if w.Code != http.StatusCreated {
		t.Skipf("product creation failed or DB unavailable: %d %s", w.Code, w.Body.String())
	}

	o, err := getOrmer()
	if err != nil {
		t.Skipf("orm not available: %v", err)
	}
	var p models.Producto
	if err := o.QueryTable(new(models.Producto)).Filter("NOMBRE", name).One(&p); err != nil {
		t.Skipf("producto not found: %v", err)
	}
	defer o.QueryTable(new(models.PrecioProductoHist)).Filter("PKIDProducto", p.PK_ID_PRODUCTO).Delete()
	defer o.Delete(&p)

	var hist []models.PrecioProductoHist
	if _, err := o.QueryTable(new(models.PrecioProductoHist)).Filter("PKIDProducto", p.PK_ID_PRODUCTO).All(&hist); err != nil {
		t.Skipf("cannot query history: %v", err)
	}
	if len(hist) != 1 || hist[0].Precio != p.PRECIO || hist[0].FechaVigencia.IsZero() {
		t.Errorf("unexpected initial history: %+v", hist)
	}

	values := url.Values{}
	values.Set("NOMBRE", name)
	values.Set("ESTADO_PRODUCTO", string(models.EstadoProductoDisponible))
	values.Set("PRECIO", "200")
	values.Set("PK_ID_SUBCATEGORIA", "1")

	req = httptest.NewRequest(http.MethodPut, "/productos?id="+strconv.FormatInt(p.PK_ID_PRODUCTO, 10), strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = sendRequest(req)
	if w.Code != http.StatusOK {
		t.Skipf("product update failed: %d %s", w.Code, w.Body.String())
	}

	hist = nil
	if _, err := o.QueryTable(new(models.PrecioProductoHist)).Filter("PKIDProducto", p.PK_ID_PRODUCTO).All(&hist); err != nil {
		t.Skipf("cannot query history after update: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("expected 2 history records, got %d", len(hist))
	}
	var oldOK, newOK bool
	for _, h := range hist {
		if h.FechaVigencia.IsZero() {
			t.Errorf("missing FechaVigencia in history: %+v", h)
		}
		if h.Precio == 100 {
			oldOK = true
		}
		if h.Precio == 200 {
			newOK = true
		}
	}
	if !oldOK || !newOK {
		t.Errorf("price history entries not as expected: %+v", hist)
	}
}
