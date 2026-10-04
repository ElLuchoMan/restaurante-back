package telemetria

import (
	stdcontext "context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	loginc "restaurante/controllers/login"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beegoctx "github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

type mockRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (r *mockRows) Columns() []string { return r.columns }

func (r *mockRows) Close() error { return nil }

func (r *mockRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	row := r.values[r.idx]
	if len(dest) != len(row) {
		return errors.New("unexpected scan destination length")
	}
	copy(dest, row)
	r.idx++
	return nil
}

type queryResponder struct {
	mu        sync.Mutex
	responses []responseSpec
}

type responseSpec struct {
	columns []string
	values  [][]driver.Value
	err     error
}

func singleRow(columns []string, values ...driver.Value) responseSpec {
	row := make([]driver.Value, len(values))
	copy(row, values)
	return responseSpec{columns: columns, values: [][]driver.Value{row}}
}

func multiRow(columns []string, rows [][]driver.Value) responseSpec {
	copied := make([][]driver.Value, len(rows))
	for i, row := range rows {
		copied[i] = make([]driver.Value, len(row))
		copy(copied[i], row)
	}
	return responseSpec{columns: columns, values: copied}
}

func copyResponses(src []responseSpec) []responseSpec {
	dst := make([]responseSpec, len(src))
	copy(dst, src)
	return dst
}

func dashboardResponses() []responseSpec {
	return []responseSpec{
		singleRow([]string{"count"}, int64(12)),
		singleRow([]string{"sum"}, int64(480000)),
		singleRow([]string{"count"}, int64(8)),
		singleRow([]string{"count"}, int64(3)),
		singleRow([]string{"sum"}, int64(54000)),
	}
}

func salesResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"metodo_pago", "total", "cantidad"},
			[][]driver.Value{{"Tarjeta", int64(300000), int64(5)}},
		),
		multiRow(
			[]string{"fecha", "total", "cantidad"},
			[][]driver.Value{{"2025-01-01", int64(120000), int64(2)}},
		),
		singleRow([]string{"venta_promedio_diaria", "pedido_promedio_diario", "ticket_promedio"}, float64(150000), float64(3), float64(50000)),
	}
}

func productsResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"producto_id", "nombre_producto", "cantidad_vendida", "ingreso_total", "precio"},
			[][]driver.Value{{int64(1), "Bandeja Paisa", int64(25), int64(250000), int64(10000)}},
		),
		multiRow(
			[]string{"producto_id", "nombre_producto", "cantidad_vendida", "ingreso_total", "precio"},
			[][]driver.Value{{int64(2), "Ensalada Cesar", int64(2), int64(20000), int64(10000)}},
		),
		singleRow([]string{"count"}, int64(42)),
	}
}

func usersResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "total_gastado", "ultimo_pedido"},
			[][]driver.Value{{int64(1010), "Ana Pérez", int64(7), int64(210000), "2025-01-01"}},
		),
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "ultimo_pedido"},
			[][]driver.Value{{int64(2020), "Luis Gómez", int64(0), "Nunca"}},
		),
		singleRow([]string{"count"}, int64(15)),
		singleRow([]string{"count"}, int64(6)),
		singleRow([]string{"avg"}, float64(32000)),
	}
}

func timeAnalysisResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"hora", "total", "cantidad"},
			[][]driver.Value{{int64(10), int64(150000), int64(4)}},
		),
		multiRow(
			[]string{"dia_semana", "total", "cantidad"},
			[][]driver.Value{{"Lunes", int64(210000), int64(6)}},
		),
		multiRow(
			[]string{"mes", "total", "cantidad"},
			[][]driver.Value{{"2025-01", int64(900000), int64(18)}},
		),
	}
}

func rentabilidadResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"producto_id", "nombre_producto", "precio_venta", "cantidad_vendida", "ingreso_total", "margen_ganancia", "ganancia_total"},
			[][]driver.Value{{int64(11), "Posta Cartagenera", int64(40000), int64(30), int64(1200000), float64(70), int64(840000)}},
		),
		multiRow(
			[]string{"producto_id", "nombre_producto", "precio_venta", "cantidad_vendida", "ingreso_total", "margen_ganancia", "ganancia_total"},
			[][]driver.Value{{int64(12), "Jugo Natural", int64(8000), int64(5), int64(40000), float64(60), int64(24000)}},
		),
		singleRow([]string{"margen_promedio_general", "total_ganancias", "total_ingresos"}, float64(65), int64(864000), int64(1240000)),
	}
}

func segmentacionResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "total_gastado", "promedio_gasto", "ultimo_pedido", "dias_sin_pedir", "segmento", "valor_vida"},
			[][]driver.Value{{int64(1), "Ana VIP", int64(8), int64(800000), float64(100000), "2025-01-05", int64(2), "VIP", int64(1600000)}},
		),
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "total_gastado", "promedio_gasto", "ultimo_pedido", "dias_sin_pedir", "segmento", "valor_vida"},
			[][]driver.Value{{int64(2), "Carlos Regular", int64(3), int64(180000), float64(60000), "2025-01-03", int64(5), "Regular", int64(270000)}},
		),
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "total_gastado", "promedio_gasto", "ultimo_pedido", "dias_sin_pedir", "segmento", "valor_vida"},
			[][]driver.Value{{int64(3), "Laura Ocasional", int64(1), int64(45000), float64(45000), "2025-01-02", int64(10), "Ocasional", int64(54000)}},
		),
		multiRow(
			[]string{"documento_cliente", "nombre_completo", "total_pedidos", "total_gastado", "promedio_gasto", "ultimo_pedido", "dias_sin_pedir", "segmento", "valor_vida"},
			[][]driver.Value{{int64(4), "Nuevo Cliente", int64(0), int64(0), float64(0), "Nunca", int64(999), "Nuevo", int64(0)}},
		),
		singleRow(
			[]string{"total_clientes_vip", "total_clientes_regulares", "total_clientes_ocasionales", "total_clientes_nuevos", "promedio_gasto_vip", "promedio_gasto_regular"},
			int64(1), int64(1), int64(1), int64(1), float64(800000), float64(180000),
		),
	}
}

func eficienciaResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"pedido_id", "cliente", "fecha_pedido", "hora_pedido", "tiempo_preparacion", "estado_pedido", "trabajador_asignado"},
			[][]driver.Value{{int64(301), "Cliente Rápido", "2025-01-01", "12:00:00", int64(45), "TERMINADO", "Juan Repartidor"}},
		),
		multiRow(
			[]string{"documento_trabajador", "nombre_trabajador", "pedidos_atendidos", "tiempo_promedio_atencion", "eficiencia_score", "horas_trabajadas"},
			[][]driver.Value{{int64(501), "Juan Repartidor", int64(12), float64(52), float64(8.5), float64(40)}},
		),
		multiRow(
			[]string{"hora", "pedidos_recibidos", "tiempo_promedio_prep", "capacidad_utilizada", "nivel_eficiencia"},
			[][]driver.Value{{"10:00", int64(5), float64(40), float64(80), "Medio"}, {"20:00", int64(8), float64(55), float64(95), "Alto"}},
		),
		singleRow([]string{"tiempo_promedio_general", "pedidos_pendientes"}, float64(48), int64(2)),
	}
}

func reservasResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"fecha", "total_reservas", "reservas_completadas", "total_personas", "porcentaje_completado"},
			[][]driver.Value{{"2025-01-01", int64(10), int64(8), int64(32), float64(80)}},
		),
		multiRow(
			[]string{"hora", "total_reservas", "reservas_completadas", "total_personas", "porcentaje_completado"},
			[][]driver.Value{{"18:00", int64(5), int64(4), int64(16), float64(80)}},
		),
		multiRow(
			[]string{"dia_semana", "total_reservas", "reservas_completadas", "total_personas", "porcentaje_completado"},
			[][]driver.Value{{"Viernes", int64(12), int64(10), int64(40), float64(83)}},
		),
		singleRow([]string{"total_reservas_completadas", "promedio_personas_por_reserva", "tasa_completamiento"}, int64(22), float64(4), float64(81)),
	}
}

func pedidosResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"fecha", "total_pedidos", "pedidos_terminados", "ingreso_total", "tasa_completamiento"},
			[][]driver.Value{{"2025-01-02", int64(18), int64(16), int64(540000), float64(88)}},
		),
		multiRow(
			[]string{"hora", "total_pedidos", "pedidos_terminados", "ingreso_total", "tasa_completamiento"},
			[][]driver.Value{{"13:00", int64(9), int64(8), int64(270000), float64(88)}},
		),
		multiRow(
			[]string{"dia_semana", "total_pedidos", "pedidos_terminados", "ingreso_total", "tasa_completamiento"},
			[][]driver.Value{{"Sábado", int64(20), int64(18), int64(600000), float64(90)}},
		),
		singleRow([]string{"total_pedidos_terminados", "tasa_completamiento_general", "ingreso_promedio_hora"}, int64(40), float64(85), float64(450000)),
	}
}

func estadosPedidosResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"estado", "count"},
			[][]driver.Value{{"TERMINADO", int64(10)}, {"EN_PROCESO", int64(3)}, {"CANCELADO", int64(1)}},
		),
	}
}

func productosPopularesResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"producto_id", "nombre_producto", "cantidad_vendida", "ingreso_total", "precio", "imagen"},
			[][]driver.Value{{int64(91), "Ajiaco", int64(40), int64(400000), int64(10000), ""}},
		),
	}
}

func productosDisponiblesResponses() []responseSpec {
	return []responseSpec{
		multiRow(
			[]string{"producto_id", "nombre_producto", "precio", "estado", "total_vendido"},
			[][]driver.Value{{int64(111), "Caldo de Costilla", int64(12000), "DISPONIBLE", int64(200)}},
		),
	}
}

func newQueryResponder(responses []responseSpec) *queryResponder {
	return &queryResponder{responses: responses}
}

func (r *queryResponder) handle(query string) (driver.Rows, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalized := strings.Join(strings.Fields(query), " ")
	if len(r.responses) == 0 {
		return nil, fmt.Errorf("unexpected query: %s", normalized)
	}
	spec := r.responses[0]
	r.responses = r.responses[1:]
	if spec.err != nil {
		return nil, spec.err
	}
	return &mockRows{columns: spec.columns, values: spec.values}, nil
}

type telemetriaMockDriver struct{}

func (telemetriaMockDriver) Open(name string) (driver.Conn, error) {
	return &telemetriaMockConn{}, nil
}

type telemetriaMockConn struct{}

func (c *telemetriaMockConn) Prepare(query string) (driver.Stmt, error) {
	return &telemetriaMockStmt{query: query}, nil
}

func (c *telemetriaMockConn) Close() error { return nil }

func (c *telemetriaMockConn) Begin() (driver.Tx, error) { return telemetriaMockTx{}, nil }

func (c *telemetriaMockConn) ExecContext(ctx stdcontext.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return telemetriaMockResult{}, nil
}

func (c *telemetriaMockConn) QueryContext(ctx stdcontext.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return consumeQuery(query)
}

func (c *telemetriaMockConn) Ping(ctx stdcontext.Context) error { return nil }

type telemetriaMockStmt struct {
	query string
}

func (s *telemetriaMockStmt) Close() error { return nil }

func (s *telemetriaMockStmt) NumInput() int { return -1 }

func (s *telemetriaMockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return telemetriaMockResult{}, nil
}

func (s *telemetriaMockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return consumeQuery(s.query)
}

func (s *telemetriaMockStmt) ExecContext(ctx stdcontext.Context, args []driver.NamedValue) (driver.Result, error) {
	return telemetriaMockResult{}, nil
}

func (s *telemetriaMockStmt) QueryContext(ctx stdcontext.Context, args []driver.NamedValue) (driver.Rows, error) {
	return consumeQuery(s.query)
}

type telemetriaMockTx struct{}

func (telemetriaMockTx) Commit() error   { return nil }
func (telemetriaMockTx) Rollback() error { return nil }

type telemetriaMockResult struct{}

func (telemetriaMockResult) LastInsertId() (int64, error) { return 0, nil }
func (telemetriaMockResult) RowsAffected() (int64, error) { return 0, nil }

var (
	responderMu      sync.Mutex
	currentResponder *queryResponder
)

func consumeQuery(query string) (driver.Rows, error) {
	responderMu.Lock()
	defer responderMu.Unlock()
	if currentResponder == nil {
		return nil, errors.New("no mock responder configured")
	}
	return currentResponder.handle(query)
}

func setMockQueries(responses []responseSpec) func() {
	responder := newQueryResponder(responses)
	responderMu.Lock()
	currentResponder = responder
	responderMu.Unlock()

	return func() {
		responderMu.Lock()
		currentResponder = nil
		responderMu.Unlock()
	}
}

func init() {
	sql.Register("telemetria-mock", telemetriaMockDriver{})
	orm.RegisterDriver("telemetria-mock", orm.DRPostgres)
	if err := orm.RegisterDataBase("default", "telemetria-mock", ""); err != nil {
		panic(err)
	}

	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "telemetria-secret")
	}
}

func generateToken(t *testing.T, nombre, rol string) string {
	t.Helper()
	secret := loginc.GetJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, loginc.Claims{
		Documento: 1,
		Rol:       rol,
		Nombre:    nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenStr
}

func generateAdminToken(t *testing.T, nombre string) string {
	return generateToken(t, nombre, string(models.RolAdministrador))
}

type apiResponse models.ApiResponse

func executeRequest(t *testing.T, handler func(*TelemetriaController), token, url string) (*httptest.ResponseRecorder, apiResponse) {
	t.Helper()

	req := httptest.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()

	ctx := beegoctx.NewContext()
	ctx.Reset(w, req)

	controller := &TelemetriaController{}
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	handler(controller)

	var resp apiResponse
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), (*models.ApiResponse)(&resp)); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
	}

	return w, resp
}

func newTestController(authHeader string, params map[string]string) (*TelemetriaController, *httptest.ResponseRecorder) {
	req := httptest.NewRequest("GET", "/telemetria/test", nil)
	query := req.URL.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	req.URL.RawQuery = query.Encode()
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	recorder := httptest.NewRecorder()
	ctx := beegoctx.NewContext()
	ctx.Reset(recorder, req)

	controller := &TelemetriaController{}
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})
	return controller, recorder
}

func TestParseFilterParamsVariants(t *testing.T) {
	t.Run("MonthYear", func(t *testing.T) {
		ctrl, _ := newTestController("", map[string]string{
			"mes":         "5",
			"año":         "2024",
			"hora_inicio": "08:30",
			"hora_fin":    "18:15",
		})

		start, end, startTime, endTime := parseFilterParams(&ctrl.Controller)
		if startTime != "08:30:00" || endTime != "18:15:00" {
			t.Fatalf("expected custom times, got %s - %s", startTime, endTime)
		}
		if start == "" || end == "" {
			t.Fatalf("expected non-empty dates for month filter")
		}
	})

	t.Run("DateRange", func(t *testing.T) {
		ctrl, _ := newTestController("", map[string]string{
			"fecha_inicio": "2025-01-01",
			"fecha_fin":    "2025-01-31",
			"hora_inicio":  "09:00:00",
			"hora_fin":     "21:30:00",
		})

		start, end, startTime, endTime := parseFilterParams(&ctrl.Controller)
		if start != "2025-01-01" || end != "2025-01-31" {
			t.Fatalf("expected explicit range, got %s - %s", start, end)
		}
		if startTime != "09:00:00" || endTime != "21:30:00" {
			t.Fatalf("expected explicit times, got %s - %s", startTime, endTime)
		}
	})
}

func TestValidateAdminRoleBranches(t *testing.T) {
	t.Run("NonAdminForbidden", func(t *testing.T) {
		token := generateToken(t, "Usuario", string(models.RolMesero))
		ctrl, rec := newTestController("Bearer "+token, nil)
		if _, ok := ctrl.validateAdminRole(); ok {
			t.Fatal("expected validation to fail for non-admin")
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden status, got %d", rec.Code)
		}
	})

	t.Run("PrefixAddedSuccess", func(t *testing.T) {
		token := generateAdminToken(t, "Prefijo")
		ctrl, rec := newTestController(token, nil)
		claims, ok := ctrl.validateAdminRole()
		if !ok {
			t.Fatal("expected validation success")
		}
		if claims == nil || claims.Nombre != "Prefijo" {
			t.Fatalf("unexpected claims: %+v", claims)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}

type endpointScenario struct {
	name           string
	url            string
	handler        func(*TelemetriaController)
	responsesFn    func() []responseSpec
	requireAuth    bool
	successMessage string
	errorMessages  []string
}

func TestTelemetriaController_FullCoverageEndpoints(t *testing.T) {
	adminName := "Cobertura Total"
	token := generateAdminToken(t, adminName)

	scenarios := []endpointScenario{
		{
			name:           "GetDashboard",
			url:            "/telemetria/dashboard",
			handler:        func(c *TelemetriaController) { c.GetDashboard() },
			responsesFn:    dashboardResponses,
			requireAuth:    true,
			successMessage: "Dashboard obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener total de pedidos",
				"Error al obtener total de ingresos",
				"Error al obtener total de usuarios",
				"Error al obtener pedidos de hoy",
				"Error al obtener ingresos de hoy",
			},
		},
		{
			name:           "GetSales",
			url:            "/telemetria/sales",
			handler:        func(c *TelemetriaController) { c.GetSales() },
			responsesFn:    salesResponses,
			requireAuth:    true,
			successMessage: "Análisis de ventas obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener ventas por método de pago",
				"Error al obtener tendencia de ventas",
				"Error al obtener estadísticas de ventas",
			},
		},
		{
			name:           "GetProducts",
			url:            "/telemetria/products?limit=0",
			handler:        func(c *TelemetriaController) { c.GetProducts() },
			responsesFn:    productsResponses,
			requireAuth:    true,
			successMessage: "Análisis de productos obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener productos más vendidos",
				"Error al obtener productos menos vendidos",
				"Error al obtener total de productos activos",
			},
		},
		{
			name:           "GetUsers",
			url:            "/telemetria/users?limit=-1",
			handler:        func(c *TelemetriaController) { c.GetUsers() },
			responsesFn:    usersResponses,
			requireAuth:    true,
			successMessage: "Análisis de usuarios obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener usuarios frecuentes",
				"Error al obtener usuarios inactivos",
				"Error al obtener total de clientes",
				"Error al obtener clientes activos",
				"Error al obtener promedio de gasto por cliente",
			},
		},
		{
			name:           "GetTimeAnalysis",
			url:            "/telemetria/time-analysis",
			handler:        func(c *TelemetriaController) { c.GetTimeAnalysis() },
			responsesFn:    timeAnalysisResponses,
			requireAuth:    true,
			successMessage: "Análisis temporal obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener ventas por hora",
				"Error al obtener ventas por día de la semana",
				"Error al obtener ventas por mes",
			},
		},
		{
			name:           "GetRentabilidad",
			url:            "/telemetria/rentabilidad?limit=abc",
			handler:        func(c *TelemetriaController) { c.GetRentabilidad() },
			responsesFn:    rentabilidadResponses,
			requireAuth:    true,
			successMessage: "Análisis de rentabilidad obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener productos rentables",
				"Error al obtener productos menos rentables",
				"Error al obtener estadísticas de rentabilidad",
			},
		},
		{
			name:           "GetSegmentacion",
			url:            "/telemetria/segmentacion?limit=xyz",
			handler:        func(c *TelemetriaController) { c.GetSegmentacion() },
			responsesFn:    segmentacionResponses,
			requireAuth:    true,
			successMessage: "Análisis de segmentación obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener clientes VIP",
				"Error al obtener clientes regulares",
				"Error al obtener clientes ocasionales",
				"Error al obtener clientes nuevos",
				"Error al obtener estadísticas de segmentación",
			},
		},
		{
			name:           "GetEficiencia",
			url:            "/telemetria/eficiencia?limit=-5",
			handler:        func(c *TelemetriaController) { c.GetEficiencia() },
			responsesFn:    eficienciaResponses,
			requireAuth:    true,
			successMessage: "Análisis de eficiencia obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener tiempos de entrega",
				"Error al obtener rendimiento de trabajadores",
				"Error al obtener análisis por hora",
				"Error al obtener estadísticas de eficiencia",
			},
		},
		{
			name:           "GetReservasAnalisis",
			url:            "/telemetria/reservas-analisis?limit=0",
			handler:        func(c *TelemetriaController) { c.GetReservasAnalisis() },
			responsesFn:    reservasResponses,
			requireAuth:    true,
			successMessage: "Análisis de reservas obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener reservas por día",
				"Error al obtener reservas por hora",
				"Error al obtener reservas por día de la semana",
				"Error al obtener estadísticas de reservas",
			},
		},
		{
			name:           "GetPedidosAnalisis",
			url:            "/telemetria/pedidos-analisis?limit=texto",
			handler:        func(c *TelemetriaController) { c.GetPedidosAnalisis() },
			responsesFn:    pedidosResponses,
			requireAuth:    true,
			successMessage: "Análisis de pedidos obtenido exitosamente por",
			errorMessages: []string{
				"Error al obtener pedidos por día",
				"Error al obtener pedidos por hora",
				"Error al obtener pedidos por día de la semana",
				"Error al obtener estadísticas de pedidos",
			},
		},
		{
			name:           "GetEstadosPedidos",
			url:            "/telemetria/estados-pedidos",
			handler:        func(c *TelemetriaController) { c.GetEstadosPedidos() },
			responsesFn:    estadosPedidosResponses,
			requireAuth:    false,
			successMessage: "Estados de pedidos obtenidos exitosamente",
			errorMessages: []string{
				"Error al obtener estados de pedidos",
			},
		},
		{
			name:           "GetProductosPopulares",
			url:            "/telemetria/productos-populares?limit=0&periodo=hoy",
			handler:        func(c *TelemetriaController) { c.GetProductosPopulares() },
			responsesFn:    productosPopularesResponses,
			requireAuth:    false,
			successMessage: "Productos populares obtenidos exitosamente",
			errorMessages: []string{
				"Error al obtener productos populares",
			},
		},
		{
			name:           "GetProductosDisponibles",
			url:            "/telemetria/productos-disponibles",
			handler:        func(c *TelemetriaController) { c.GetProductosDisponibles() },
			responsesFn:    productosDisponiblesResponses,
			requireAuth:    false,
			successMessage: "Productos disponibles obtenidos exitosamente",
			errorMessages: []string{
				"Error al obtener productos disponibles",
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name+"_Success", func(t *testing.T) {
			responses := copyResponses(scenario.responsesFn())
			cleanup := setMockQueries(responses)
			defer cleanup()

			useToken := ""
			if scenario.requireAuth {
				useToken = token
			}

			w, resp := executeRequest(t, scenario.handler, useToken, scenario.url)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}
			if resp.Code != http.StatusOK {
				t.Fatalf("expected response code 200, got %d", resp.Code)
			}
			if scenario.requireAuth && !strings.Contains(resp.Message, scenario.successMessage) {
				t.Fatalf("expected message to contain %q, got %q", scenario.successMessage, resp.Message)
			}
			if scenario.requireAuth && !strings.Contains(resp.Message, adminName) {
				t.Fatalf("expected message to contain admin name %q, got %q", adminName, resp.Message)
			}
			if !scenario.requireAuth && resp.Message != scenario.successMessage {
				t.Fatalf("expected message %q, got %q", scenario.successMessage, resp.Message)
			}
		})

		for idx, msg := range scenario.errorMessages {
			t.Run(fmt.Sprintf("%s_Error_%d", scenario.name, idx+1), func(t *testing.T) {
				responses := copyResponses(scenario.responsesFn())
				responses[idx].err = errors.New("forced error")
				cleanup := setMockQueries(responses)
				defer cleanup()

				useToken := ""
				if scenario.requireAuth {
					useToken = token
				}

				w, resp := executeRequest(t, scenario.handler, useToken, scenario.url)

				if w.Code != http.StatusInternalServerError {
					t.Fatalf("expected status 500, got %d", w.Code)
				}
				if resp.Code != http.StatusInternalServerError {
					t.Fatalf("expected response code 500, got %d", resp.Code)
				}
				if resp.Message != msg {
					t.Fatalf("expected error message %q, got %q", msg, resp.Message)
				}
				if resp.Cause != "forced error" {
					t.Fatalf("expected cause 'forced error', got %q", resp.Cause)
				}
			})
		}
	}
}
