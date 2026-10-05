package cupon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// stubQS implementa solo lo que usa el servicio de cupones; el resto de la
// interfaz orm.QuerySeter queda embebida (nil) y no debe invocarse.
type stubQS struct {
	orm.QuerySeter
	table string
	st    *stubOrm
}

func (q *stubQS) Filter(string, ...interface{}) orm.QuerySeter { return q }

func (q *stubQS) One(container interface{}, _ ...string) error {
	if q.st.oneErr != nil {
		return q.st.oneErr
	}
	if c, ok := container.(*models.Cupon); ok && q.st.cupon != nil {
		*c = *q.st.cupon
	}
	return nil
}

func (q *stubQS) Count() (int64, error) { return q.st.count, nil }

type stubOrm struct {
	orm.Ormer
	cupon     *models.Cupon
	oneErr    error
	count     int64
	insertErr error
	inserted  interface{}
}

func (s *stubOrm) QueryTable(i interface{}) orm.QuerySeter {
	name, _ := i.(string)
	return &stubQS{table: name, st: s}
}

func (s *stubOrm) Insert(v interface{}) (int64, error) {
	s.inserted = v
	return 1, s.insertErr
}

func useRealService(t *testing.T, st *stubOrm) {
	t.Helper()
	origSvc, origProv, origZone := newCuponService, ormProvider, database.BogotaZone
	database.BogotaZone = time.UTC
	newCuponService = realNewCuponService
	ormProvider = func() orm.Ormer { return st }
	t.Cleanup(func() { newCuponService, ormProvider, database.BogotaZone = origSvc, origProv, origZone })
}

func newCtrl(method, target string, body interface{}) (*CuponController, *httptest.ResponseRecorder) {
	b, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, bytes.NewReader(b))
	ctx := beecontext.NewContext()
	ctx.Reset(rec, req)
	ctx.Input.RequestBody = b
	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, rec
}

func vigente(activo bool) *models.Cupon {
	return &models.Cupon{
		PkIdCupon: 10, Codigo: "OK", Scope: models.CuponScopeGlobal,
		TipoDescuento: models.TipoDescuentoPorcentaje, ValorDescuento: 10,
		FechaInicio: time.Now().AddDate(-1, 0, 0), FechaFin: time.Now().AddDate(1, 0, 0),
		Activo: activo,
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var r models.ApiResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	return r
}

func TestPost_AsignaRelacionesSegunScope(t *testing.T) {
	id := int64(42)
	cases := []struct {
		name  string
		scope string
		set   func(m map[string]interface{})
		check func(t *testing.T, c *models.Cupon)
	}{
		{"producto", "PRODUCTO", func(m map[string]interface{}) { m["productoId"] = id }, func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkIdProducto)
			assert.Equal(t, id, c.PkIdProducto.PK_ID_PRODUCTO)
		}},
		{"categoria", "CATEGORIA", func(m map[string]interface{}) { m["categoriaId"] = id }, func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkIdCategoria)
			assert.Equal(t, id, c.PkIdCategoria.PK_ID_CATEGORIA)
		}},
		{"cliente", "CLIENTE", func(m map[string]interface{}) { m["documentoCliente"] = id }, func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkDocumentoCliente)
			assert.Equal(t, id, c.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller, recorder, ctx := setupTest()
			body := map[string]interface{}{
				"codigo": "REL" + tc.name, "scope": tc.scope, "tipoDescuento": "PORCENTAJE",
				"valorDescuento": 10, "fechaInicio": "2025-01-01", "fechaFin": "2025-12-31",
			}
			tc.set(body)
			raw, _ := json.Marshal(body)
			ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(raw))
			ctx.Input.RequestBody = raw

			var got *models.Cupon
			mockOrmer.On("Insert", mock.AnythingOfType("*models.Cupon")).Run(func(args mock.Arguments) {
				got = args.Get(0).(*models.Cupon)
			}).Return(int64(1), nil)

			controller.Post()

			assert.Equal(t, http.StatusCreated, recorder.Code)
			require.NotNil(t, got)
			tc.check(t, got)
		})
	}
}

func TestPost_ReglaDeNegocioInvalida(t *testing.T) {
	controller, recorder, ctx := setupTest()
	body, _ := json.Marshal(map[string]interface{}{
		"codigo": "BADDATES", "scope": "GLOBAL", "tipoDescuento": "PORCENTAJE",
		"valorDescuento": 10, "fechaInicio": "2025-12-31", "fechaFin": "2025-01-01",
	})
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Input.RequestBody = body

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	resp := decode(t, recorder)
	assert.Equal(t, "Error de validación", resp.Message)
	assert.Contains(t, resp.Cause, "fecha de fin")
	mockOrmer.AssertNotCalled(t, "Insert", mock.Anything)
}

func TestPut_AsignaRelacionesSegunScope(t *testing.T) {
	id := int64(42)
	cases := []struct {
		name  string
		scope string
		key   string
		check func(t *testing.T, c *models.Cupon)
	}{
		{"producto", "PRODUCTO", "productoId", func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkIdProducto)
			assert.Equal(t, id, c.PkIdProducto.PK_ID_PRODUCTO)
		}},
		{"categoria", "CATEGORIA", "categoriaId", func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkIdCategoria)
			assert.Equal(t, id, c.PkIdCategoria.PK_ID_CATEGORIA)
		}},
		{"cliente", "CLIENTE", "documentoCliente", func(t *testing.T, c *models.Cupon) {
			require.NotNil(t, c.PkDocumentoCliente)
			assert.Equal(t, id, c.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller, recorder, _ := setupTest()
			body, _ := json.Marshal(map[string]interface{}{
				"codigo": "UPD", "scope": tc.scope, "tipoDescuento": "PORCENTAJE",
				"valorDescuento": 10, "fechaInicio": "2025-01-01", "fechaFin": "2025-12-31", tc.key: id,
			})
			req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
			ctx := beecontext.NewContext()
			ctx.Reset(recorder, req)
			ctx.Input.RequestBody = body
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})

			mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(nil)
			var got *models.Cupon
			mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
				got = args.Get(0).(*models.Cupon)
			}).Return(int64(1), nil)

			controller.Put()

			assert.Equal(t, http.StatusOK, recorder.Code)
			require.NotNil(t, got)
			tc.check(t, got)
		})
	}
}

func TestValidarCupon_ServicioReal(t *testing.T) {
	t.Run("aplicable", func(t *testing.T) {
		useRealService(t, &stubOrm{cupon: vigente(true)})
		c, rec := newCtrl("POST", "/cupones/validar", models.ValidarCuponRequest{
			Codigo: "OK", ClienteId: 1,
			Items: []models.ValidarCuponItemRequest{{ProductoId: 1, Cantidad: 2, Precio: 1000}},
		})
		c.ValidarCupon()
		assert.Equal(t, http.StatusOK, rec.Code)
		resp := decode(t, rec)
		assert.Equal(t, "Cupón validado exitosamente", resp.Message)
		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, true, data["aplicable"])
		assert.Equal(t, float64(200), data["montoDescuento"])
	})
	t.Run("inactivo", func(t *testing.T) {
		useRealService(t, &stubOrm{cupon: vigente(false)})
		c, rec := newCtrl("POST", "/cupones/validar", models.ValidarCuponRequest{Codigo: "OK", ClienteId: 1})
		c.ValidarCupon()
		assert.Equal(t, http.StatusOK, rec.Code)
		data := decode(t, rec).Data.(map[string]interface{})
		assert.Equal(t, false, data["aplicable"])
		assert.Equal(t, "Cupón inactivo", data["motivo"])
	})
	t.Run("error del servicio", func(t *testing.T) {
		useRealService(t, &stubOrm{oneErr: fmt.Errorf("boom")})
		c, rec := newCtrl("POST", "/cupones/validar", models.ValidarCuponRequest{Codigo: "OK", ClienteId: 1})
		c.ValidarCupon()
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, decode(t, rec).Cause, "boom")
	})
}

func TestRedimirCupon_ServicioReal(t *testing.T) {
	t.Run("exitoso", func(t *testing.T) {
		st := &stubOrm{cupon: vigente(true)}
		useRealService(t, st)
		pedido := int64(5)
		c, rec := newCtrl("POST", "/cupones/OK/redimir", models.RedimirCuponRequest{ClienteId: 7, PedidoId: &pedido})
		c.Ctx.Input.SetParam(":codigo", "OK")
		c.RedimirCupon()
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "Cupón redimido exitosamente", decode(t, rec).Message)
		red, ok := st.inserted.(*models.CuponRedencion)
		require.True(t, ok)
		assert.Equal(t, int64(7), red.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
		assert.Equal(t, pedido, red.PkIdPedido.PK_ID_PEDIDO)
	})
	t.Run("no encontrado", func(t *testing.T) {
		useRealService(t, &stubOrm{oneErr: orm.ErrNoRows})
		c, rec := newCtrl("POST", "/cupones/NO/redimir", models.RedimirCuponRequest{ClienteId: 7})
		c.Ctx.Input.SetParam(":codigo", "NO")
		c.RedimirCupon()
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		resp := decode(t, rec)
		assert.Equal(t, "Error al redimir cupón", resp.Message)
		assert.True(t, strings.HasPrefix(resp.Cause, "cupón no encontrado"))
	})
	t.Run("no aplicable", func(t *testing.T) {
		useRealService(t, &stubOrm{cupon: vigente(false)})
		c, rec := newCtrl("POST", "/cupones/OK/redimir", models.RedimirCuponRequest{ClienteId: 7})
		c.Ctx.Input.SetParam(":codigo", "OK")
		c.RedimirCupon()
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, decode(t, rec).Cause, "cupón no aplicable: Cupón inactivo")
	})
}

func TestNewCuponService_ConOrmerYNil(t *testing.T) {
	assert.NotNil(t, realNewCuponService(nil))
	st := &stubOrm{cupon: vigente(true)}
	useRealService(t, st)
	svc := realNewCuponService(st)
	require.NotNil(t, svc)
	resp, err := svc.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "OK", ClienteId: 1})
	require.NoError(t, err)
	assert.True(t, resp.Aplicable)
}

func TestCupOrmNewDefault_UsaOrmReal(t *testing.T) {
	// Restaura el constructor real (la suite lo sustituye por mocks en init()).
	assert.NotNil(t, realCupOrmNew())
}
