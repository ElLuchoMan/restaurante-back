package telemetria

import (
	"net/http"
	"strings"
	"testing"
)

// Sin filas, todas las listas de la respuesta deben salir como [] y nunca null.
func TestListasVaciasSeSerializanComoArreglo(t *testing.T) {
	scalar := map[string]bool{
		"count": true, "sum": true, "avg": true,
		"venta_promedio_diaria": true, "margen_promedio_general": true,
		"total_clientes_vip": true, "tiempo_promedio_general": true,
		"total_reservas_completadas": true, "total_pedidos_terminados": true,
	}
	vaciar := func(specs []responseSpec) []responseSpec {
		out := copyResponses(specs)
		for i := range out {
			if !scalar[out[i].columns[0]] {
				out[i].values = nil
			}
		}
		return out
	}

	casos := []struct {
		nombre  string
		url     string
		handler func(*TelemetriaController)
		resp    func() []responseSpec
		auth    bool
		listas  []string
	}{
		{"sales", "/telemetria/sales", (*TelemetriaController).GetSales, salesResponses, true, []string{`"ventasPorMetodoPago":[]`, `"tendenciaVentas":[]`}},
		{"products", "/telemetria/products", (*TelemetriaController).GetProducts, productsResponses, true, []string{`"productosMasVendidos":[]`, `"productosMenosVendidos":[]`}},
		{"users", "/telemetria/users", (*TelemetriaController).GetUsers, usersResponses, true, []string{`"usuariosFrecuentes":[]`, `"usuariosInactivos":[]`}},
		{"time", "/telemetria/time-analysis", (*TelemetriaController).GetTimeAnalysis, timeAnalysisResponses, true, []string{`"ventasPorHora":[]`, `"ventasPorDiaSemana":[]`, `"ventasPorMes":[]`}},
		{"rentabilidad", "/telemetria/rentabilidad", (*TelemetriaController).GetRentabilidad, rentabilidadResponses, true, []string{`"productosRentables":[]`, `"productosMenosRentables":[]`}},
		{"segmentacion", "/telemetria/segmentacion", (*TelemetriaController).GetSegmentacion, segmentacionResponses, true, []string{`"clientesVIP":[]`, `"clientesRegulares":[]`, `"clientesOcasionales":[]`, `"clientesNuevos":[]`}},
		{"eficiencia", "/telemetria/eficiencia", (*TelemetriaController).GetEficiencia, eficienciaResponses, true, []string{`"tiemposEntrega":[]`, `"rendimientoTrabajadores":[]`, `"analisisPorHora":[]`}},
		{"reservas", "/telemetria/reservas-analisis", (*TelemetriaController).GetReservasAnalisis, reservasResponses, true, []string{`"reservasPorDia":[]`, `"reservasPorHora":[]`, `"reservasPorDiaSemana":[]`}},
		{"pedidos", "/telemetria/pedidos-analisis", (*TelemetriaController).GetPedidosAnalisis, pedidosResponses, true, []string{`"pedidosPorDia":[]`, `"pedidosPorHora":[]`, `"pedidosPorDiaSemana":[]`}},
		{"populares", "/productos-populares", (*TelemetriaController).GetProductosPopulares, productosPopularesResponses, false, []string{`"productosPopulares":[]`}},
		{"disponibles", "/productos-disponibles", (*TelemetriaController).GetProductosDisponibles, productosDisponiblesResponses, false, []string{`"data":[]`}},
	}

	token := generateAdminToken(t, "Admin")
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			defer setMockQueries(vaciar(c.resp()))()
			tk := ""
			if c.auth {
				tk = token
			}
			w, _ := executeRequest(t, c.handler, tk, c.url)
			if w.Code != http.StatusOK {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			body := strings.Join(strings.Fields(w.Body.String()), "")
			if strings.Contains(body, "null") {
				t.Errorf("la respuesta contiene null: %s", body)
			}
			for _, l := range c.listas {
				if !strings.Contains(body, l) {
					t.Errorf("falta %s en %s", l, body)
				}
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	casos := []struct {
		query string
		def   int
		want  int
	}{
		{"", 10, 10},
		{"5", 10, 5},
		{"100", 10, 100},
		{"101", 10, 100},
		{"100000", 4, 100},
		{"0", 4, 4},
		{"-3", 4, 4},
		{"abc", 4, 4},
	}
	for _, c := range casos {
		params := map[string]string{}
		if c.query != "" {
			params["limit"] = c.query
		}
		ctrl, _ := newTestController("", params)
		if got := parseLimit(&ctrl.Controller, c.def); got != c.want {
			t.Errorf("limit=%q def=%d: got %d want %d", c.query, c.def, got, c.want)
		}
	}
}
