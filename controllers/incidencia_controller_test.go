package controllers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// mockOrmer and mockQuerySeter allow us to simulate database interactions
// without touching a real database.
type mockOrmer struct {
	qs        querySeter
	insertErr error
	readErr   error
	updateErr error
	deleteErr error
}

func (m *mockOrmer) QueryTable(interface{}) querySeter { return m.qs }
func (m *mockOrmer) Insert(interface{}) (int64, error) { return 1, m.insertErr }
func (m *mockOrmer) Read(interface{}, ...string) error { return m.readErr }
func (m *mockOrmer) Update(interface{}, ...string) (int64, error) {
	return 1, m.updateErr
}
func (m *mockOrmer) Delete(interface{}, ...string) (int64, error) {
	return 1, m.deleteErr
}

type mockQuerySeter struct {
	allErr error
	data   []models.Incidencia
}

func (m *mockQuerySeter) Filter(string, interface{}) querySeter { return m }
func (m *mockQuerySeter) All(container interface{}) (int64, error) {
	slice := container.(*[]models.Incidencia)
	if m.data != nil {
		*slice = append(*slice, m.data...)
	}
	if m.allErr != nil {
		if len(*slice) == 0 {
			*slice = append(*slice, models.Incidencia{})
		}
		return int64(len(*slice)), m.allErr
	}
	return int64(len(*slice)), nil
}

// helper to create controller with recorder and request
func newIncidenciaCtx(method, url string, body string) (*IncidenciaController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	b, _ := io.ReadAll(r.Body)
	ctx.Input.RequestBody = b
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return &c, w
}

// TestIncidenciaWrapperCoverage ensures that the real wrapper types are also
// executed at least once so that their lines are included in coverage reports.
func TestIncidenciaWrapperCoverage(t *testing.T) {
	o := newOrmer()
	qs := o.QueryTable(new(models.Incidencia))
	qs = qs.Filter("PK_ID_INCIDENCIA", 1)
	var l []models.Incidencia
	qs.All(&l)
}

func TestIncidenciaGetAll(t *testing.T) {
	cases := []struct {
		name   string
		qs     querySeter
		status int
	}{
		{
			name:   "db error",
			qs:     &mockQuerySeter{data: []models.Incidencia{{PK_ID_INCIDENCIA: 1}}, allErr: errors.New("fail")},
			status: http.StatusInternalServerError,
		},
		{
			name:   "success",
			qs:     &mockQuerySeter{data: []models.Incidencia{{PK_ID_INCIDENCIA: 1}}},
			status: http.StatusOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := newOrmer
			defer func() { newOrmer = old }()
			newOrmer = func() ormer { return &mockOrmer{qs: tc.qs} }
			c, w := newIncidenciaCtx(http.MethodGet, "/incidencias", "")
			c.GetAll()
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestIncidenciaGetByDocumentAndDateValidations(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"invalid document", "/incidencias/search"},
		{"invalid month", "/incidencias/search?documento=1&mes=13&anio=2024"},
		{"invalid year", "/incidencias/search?documento=1&mes=1&anio=1800"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newIncidenciaCtx(http.MethodGet, tc.url, "")
			c.GetByDocumentAndDate()
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestIncidenciaGetByDocumentAndDateResults(t *testing.T) {
	cases := []struct {
		name   string
		qs     querySeter
		status int
	}{
		{
			name:   "not found",
			qs:     &mockQuerySeter{allErr: orm.ErrNoRows},
			status: http.StatusOK,
		},
		{
			name:   "db error",
			qs:     &mockQuerySeter{allErr: errors.New("fail")},
			status: http.StatusInternalServerError,
		},
		{
			name:   "success",
			qs:     &mockQuerySeter{data: []models.Incidencia{{PK_ID_INCIDENCIA: 1}}},
			status: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := newOrmer
			defer func() { newOrmer = old }()
			newOrmer = func() ormer { return &mockOrmer{qs: tc.qs} }
			c, w := newIncidenciaCtx(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", "")
			c.GetByDocumentAndDate()
			if w.Code != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestIncidenciaPostValidations(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"invalid json", "notjson"},
		{"missing fecha", `{"MONTO":100,"RESTA":false,"MOTIVO":"m"}`},
		{"invalid fecha", `{"FECHA":"x","MONTO":100,"RESTA":false,"MOTIVO":"m"}`},
		{"missing monto", `{"FECHA":"2024-01-01","RESTA":false,"MOTIVO":"m"}`},
		{"missing resta", `{"FECHA":"2024-01-01","MONTO":1,"MOTIVO":"m"}`},
		{"missing motivo", `{"FECHA":"2024-01-01","MONTO":1,"RESTA":false}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newIncidenciaCtx(http.MethodPost, "/incidencias", tc.body)
			c.Post()
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestIncidenciaPostInsert(t *testing.T) {
	cases := []struct {
		name      string
		insertErr error
		status    int
	}{
		{"insert error", errors.New("fail"), http.StatusInternalServerError},
		{"success", nil, http.StatusCreated},
	}

	body := `{"FECHA":"2024-01-01","MONTO":1,"RESTA":true,"MOTIVO":"m","PK_DOCUMENTO_TRABAJADOR":1}`

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := newOrmer
			defer func() { newOrmer = old }()
			newOrmer = func() ormer { return &mockOrmer{insertErr: tc.insertErr} }
			c, w := newIncidenciaCtx(http.MethodPost, "/incidencias", body)
			c.Post()
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestIncidenciaPutInvalidCases(t *testing.T) {
	cases := []struct {
		name string
		url  string
		body string
	}{
		{"invalid id", "/incidencias?id=abc", "{}"},
		{"invalid json", "/incidencias?id=1", "notjson"},
		{"invalid fecha", "/incidencias?id=1", `{"FECHA":"x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newIncidenciaCtx(http.MethodPut, tc.url, tc.body)
			c.Put()
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestIncidenciaPutScenarios(t *testing.T) {
	cases := []struct {
		name      string
		readErr   error
		updateErr error
		status    int
	}{
		{"not found", orm.ErrNoRows, nil, http.StatusOK},
		{"update error", nil, errors.New("fail"), http.StatusInternalServerError},
		{"success", nil, nil, http.StatusOK},
	}

	body := `{"FECHA":"2024-01-01","MONTO":1,"RESTA":true,"MOTIVO":"m","PK_DOCUMENTO_TRABAJADOR":1}`

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := newOrmer
			defer func() { newOrmer = old }()
			newOrmer = func() ormer { return &mockOrmer{readErr: tc.readErr, updateErr: tc.updateErr} }
			c, w := newIncidenciaCtx(http.MethodPut, "/incidencias?id=1", body)
			c.Put()
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestIncidenciaDeleteScenarios(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		deleteErr error
		status    int
	}{
		{"invalid id", "/incidencias?id=abc", nil, http.StatusBadRequest},
		{"not found", "/incidencias?id=1", orm.ErrNoRows, http.StatusOK},
		{"delete error", "/incidencias?id=1", errors.New("fail"), http.StatusInternalServerError},
		{"success", "/incidencias?id=1", nil, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := newOrmer
			defer func() { newOrmer = old }()
			newOrmer = func() ormer { return &mockOrmer{deleteErr: tc.deleteErr} }
			c, w := newIncidenciaCtx(http.MethodDelete, tc.url, "")
			c.Delete()
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}
