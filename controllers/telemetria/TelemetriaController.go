package telemetria

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	loginc "restaurante/controllers/login"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type TelemetriaController struct {
	web.Controller
}

// Usamos los Claims del paquete login para mantener consistencia
type Claims = loginc.Claims

// TimeFilter representa los filtros de tiempo disponibles
type TimeFilter string

const (
	FilterToday       TimeFilter = "hoy"
	FilterLastWeek    TimeFilter = "ultima_semana"
	FilterLastMonth   TimeFilter = "ultimo_mes"
	FilterLast3Months TimeFilter = "ultimos_3_meses"
	FilterLast6Months TimeFilter = "ultimos_6_meses"
	FilterLastYear    TimeFilter = "ultimo_año"
	FilterHistoric    TimeFilter = "historico"
)

// getTimeRange retorna las fechas de inicio y fin basadas en el filtro
func getTimeRange(filter TimeFilter) (startDate, endDate string) {
	now := time.Now()

	switch filter {
	case FilterToday:
		startDate = now.Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterLastWeek:
		startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterLastMonth:
		startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterLast3Months:
		startDate = now.AddDate(0, -3, 0).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterLast6Months:
		startDate = now.AddDate(0, -6, 0).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterLastYear:
		startDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	case FilterHistoric:
		// Para histórico, usamos una fecha muy antigua
		startDate = "1900-01-01"
		endDate = now.Format("2006-01-02")
	default:
		// Por defecto, último mes
		startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	}

	return startDate, endDate
}

// buildDateFilter construye la condición SQL para filtrar por fechas
func buildDateFilter(startDate, endDate string) string {
	if startDate == endDate {
		return fmt.Sprintf("pe.fecha = '%s'", startDate)
	}
	return fmt.Sprintf("pe.fecha >= '%s' AND pe.fecha <= '%s'", startDate, endDate)
}

// DashboardData representa los datos del dashboard general
type DashboardData struct {
	TotalPedidos        int64   `json:"totalPedidos"`
	TotalIngresos       int64   `json:"totalIngresos"`
	TotalUsuarios       int64   `json:"totalUsuarios"`
	PromedioVentaPedido float64 `json:"promedioVentaPedido"`
	PedidosHoy          int64   `json:"pedidosHoy"`
	IngresosHoy         int64   `json:"ingresosHoy"`
}

// SalesData representa los datos de análisis de ventas
type SalesData struct {
	VentasPorMetodoPago   []VentaPorMetodo   `json:"ventasPorMetodoPago"`
	TendenciaVentas       []VentaPorFecha    `json:"tendenciaVentas"`
	EstadisticasGenerales EstadisticasVentas `json:"estadisticasGenerales"`
}

type VentaPorMetodo struct {
	MetodoPago string `json:"metodoPago"`
	Total      int64  `json:"total"`
	Cantidad   int64  `json:"cantidad"`
}

type VentaPorFecha struct {
	Fecha    string `json:"fecha"`
	Total    int64  `json:"total"`
	Cantidad int64  `json:"cantidad"`
}

type EstadisticasVentas struct {
	VentaPromedioDiaria  float64 `json:"ventaPromedioDiaria"`
	PedidoPromedioDiario float64 `json:"pedidoPromedioDiario"`
	TicketPromedio       float64 `json:"ticketPromedio"`
}

// ProductsData representa los datos de análisis de productos
type ProductsData struct {
	ProductosMasVendidos   []ProductoVendido     `json:"productosMasVendidos"`
	ProductosMenosVendidos []ProductoVendido     `json:"productosMenosVendidos"`
	EstadisticasProductos  EstadisticasProductos `json:"estadisticasProductos"`
}

type ProductoVendido struct {
	ProductoID      int64  `json:"productoId"`
	NombreProducto  string `json:"nombreProducto"`
	CantidadVendida int64  `json:"cantidadVendida"`
	IngresoTotal    int64  `json:"ingresoTotal"`
	Precio          int64  `json:"precio"`
	Imagen          string `json:"imagen"`
}

type EstadisticasProductos struct {
	TotalProductosActivos  int64  `json:"totalProductosActivos"`
	ProductoConMasVentas   string `json:"productoConMasVentas"`
	ProductoConMenosVentas string `json:"productoConMenosVentas"`
}

// UsersData representa los datos de análisis de usuarios
type UsersData struct {
	UsuariosFrecuentes   []UsuarioFrecuente   `json:"usuariosFrecuentes"`
	UsuariosInactivos    []UsuarioInactivo    `json:"usuariosInactivos"`
	EstadisticasUsuarios EstadisticasUsuarios `json:"estadisticasUsuarios"`
}

type UsuarioFrecuente struct {
	DocumentoCliente int64  `json:"documentoCliente"`
	NombreCompleto   string `json:"nombreCompleto"`
	TotalPedidos     int64  `json:"totalPedidos"`
	TotalGastado     int64  `json:"totalGastado"`
	UltimoPedido     string `json:"ultimoPedido"`
}

type UsuarioInactivo struct {
	DocumentoCliente int64  `json:"documentoCliente"`
	NombreCompleto   string `json:"nombreCompleto"`
	TotalPedidos     int64  `json:"totalPedidos"`
	UltimoPedido     string `json:"ultimoPedido"`
}

type EstadisticasUsuarios struct {
	TotalClientes           int64   `json:"totalClientes"`
	ClientesActivos         int64   `json:"clientesActivos"`
	ClientesInactivos       int64   `json:"clientesInactivos"`
	PromedioGastoPorCliente float64 `json:"promedioGastoPorCliente"`
}

// TimeAnalysisData representa los datos de análisis temporal
type TimeAnalysisData struct {
	VentasPorHora      []VentaPorHora      `json:"ventasPorHora"`
	VentasPorDiaSemana []VentaPorDiaSemana `json:"ventasPorDiaSemana"`
	VentasPorMes       []VentaPorMes       `json:"ventasPorMes"`
}

type VentaPorHora struct {
	Hora     int   `json:"hora"`
	Total    int64 `json:"total"`
	Cantidad int64 `json:"cantidad"`
}

type VentaPorDiaSemana struct {
	DiaSemana string `json:"diaSemana"`
	Total     int64  `json:"total"`
	Cantidad  int64  `json:"cantidad"`
}

type VentaPorMes struct {
	Mes      string `json:"mes"`
	Total    int64  `json:"total"`
	Cantidad int64  `json:"cantidad"`
}

// === NUEVAS ESTRUCTURAS PARA MÉTRICAS AVANZADAS ===

// RentabilidadData representa el análisis de rentabilidad
type RentabilidadData struct {
	ProductosRentables       []ProductoRentabilidad   `json:"productosRentables"`
	ProductosMenosRentables  []ProductoRentabilidad   `json:"productosMenosRentables"`
	EstadisticasRentabilidad EstadisticasRentabilidad `json:"estadisticasRentabilidad"`
}

type ProductoRentabilidad struct {
	ProductoID      int64   `json:"productoId"`
	NombreProducto  string  `json:"nombreProducto"`
	PrecioVenta     int64   `json:"precioVenta"`
	CantidadVendida int64   `json:"cantidadVendida"`
	IngresoTotal    int64   `json:"ingresoTotal"`
	MargenGanancia  float64 `json:"margenGanancia"` // Porcentaje de ganancia
	GananciaTotal   int64   `json:"gananciaTotal"`  // Ganancia total en pesos
}

type EstadisticasRentabilidad struct {
	MargenPromedioGeneral float64 `json:"margenPromedioGeneral"`
	ProductoMasRentable   string  `json:"productoMasRentable"`
	ProductoMenosRentable string  `json:"productoMenosRentable"`
	TotalGanancias        int64   `json:"totalGanancias"`
	TotalIngresos         int64   `json:"totalIngresos"`
}

// SegmentacionData representa el análisis de segmentación de clientes
type SegmentacionData struct {
	ClientesVIP              []ClienteSegmento        `json:"clientesVIP"`
	ClientesRegulares        []ClienteSegmento        `json:"clientesRegulares"`
	ClientesOcasionales      []ClienteSegmento        `json:"clientesOcasionales"`
	ClientesNuevos           []ClienteSegmento        `json:"clientesNuevos"`
	EstadisticasSegmentacion EstadisticasSegmentacion `json:"estadisticasSegmentacion"`
}

type ClienteSegmento struct {
	DocumentoCliente int64   `json:"documentoCliente"`
	NombreCompleto   string  `json:"nombreCompleto"`
	TotalPedidos     int64   `json:"totalPedidos"`
	TotalGastado     int64   `json:"totalGastado"`
	PromedioGasto    float64 `json:"promedioGasto"`
	UltimoPedido     string  `json:"ultimoPedido"`
	DiasSinPedir     int     `json:"diasSinPedir"`
	Segmento         string  `json:"segmento"`
	ValorVida        int64   `json:"valorVida"` // CLV estimado
}

type EstadisticasSegmentacion struct {
	TotalClientesVIP         int64   `json:"totalClientesVIP"`
	TotalClientesRegulares   int64   `json:"totalClientesRegulares"`
	TotalClientesOcasionales int64   `json:"totalClientesOcasionales"`
	TotalClientesNuevos      int64   `json:"totalClientesNuevos"`
	PromedioGastoVIP         float64 `json:"promedioGastoVIP"`
	PromedioGastoRegular     float64 `json:"promedioGastoRegular"`
	PorcentajeVIP            float64 `json:"porcentajeVIP"`
}

// EficienciaData representa el análisis de eficiencia de entregas
type EficienciaData struct {
	TiemposEntrega          []TiempoEntrega         `json:"tiemposEntrega"`
	RendimientoTrabajadores []RendimientoTrabajador `json:"rendimientoTrabajadores"`
	AnalisisPorHora         []EficienciaPorHora     `json:"analisisPorHora"`
	EstadisticasEficiencia  EstadisticasEficiencia  `json:"estadisticasEficiencia"`
}

type TiempoEntrega struct {
	PedidoID           int64  `json:"pedidoId"`
	Cliente            string `json:"cliente"`
	FechaPedido        string `json:"fechaPedido"`
	HoraPedido         string `json:"horaPedido"`
	TiempoPreparacion  int    `json:"tiempoPreparacion"` // en minutos
	EstadoPedido       string `json:"estadoPedido"`
	TrabajadorAsignado string `json:"trabajadorAsignado"`
}

type RendimientoTrabajador struct {
	DocumentoTrabajador    int64   `json:"documentoTrabajador"`
	NombreTrabajador       string  `json:"nombreTrabajador"`
	PedidosAtendidos       int64   `json:"pedidosAtendidos"`
	TiempoPromedioAtencion float64 `json:"tiempoPromedioAtencion"`
	EficienciaScore        float64 `json:"eficienciaScore"` // 1-10
	HorasTrabajadas        float64 `json:"horasTrabajadas"`
}

type EficienciaPorHora struct {
	Hora               string  `json:"hora"`
	PedidosRecibidos   int64   `json:"pedidosRecibidos"`
	TiempoPromedioPrep float64 `json:"tiempoPromedioPrep"`
	CapacidadUtilizada float64 `json:"capacidadUtilizada"` // Porcentaje
	NivelEficiencia    string  `json:"nivelEficiencia"`    // Alto, Medio, Bajo
}

type EstadisticasEficiencia struct {
	TiempoPromedioGeneral  float64 `json:"tiempoPromedioGeneral"`
	HoraMasEficiente       string  `json:"horaMasEficiente"`
	HoraMenosEficiente     string  `json:"horaMenosEficiente"`
	TrabajadorMasEficiente string  `json:"trabajadorMasEficiente"`
	CapacidadPromedioUso   float64 `json:"capacidadPromedioUso"`
	PedidosPendientes      int64   `json:"pedidosPendientes"`
}

// === NUEVOS ENDPOINTS ADICIONALES ===

// ReservasAnalisisData representa el análisis de reservas por días y horas
type ReservasAnalisisData struct {
	ReservasPorDia       []ReservaPorDia       `json:"reservasPorDia"`
	ReservasPorHora      []ReservaPorHora      `json:"reservasPorHora"`
	ReservasPorDiaSemana []ReservaPorDiaSemana `json:"reservasPorDiaSemana"`
	EstadisticasReservas EstadisticasReservas  `json:"estadisticasReservas"`
}

type ReservaPorDia struct {
	Fecha                string  `json:"fecha"`
	TotalReservas        int64   `json:"totalReservas"`
	ReservasCompletadas  int64   `json:"reservasCompletadas"`
	TotalPersonas        int64   `json:"totalPersonas"`
	PorcentajeCompletado float64 `json:"porcentajeCompletado"`
}

type ReservaPorHora struct {
	Hora                 string  `json:"hora"`
	TotalReservas        int64   `json:"totalReservas"`
	ReservasCompletadas  int64   `json:"reservasCompletadas"`
	TotalPersonas        int64   `json:"totalPersonas"`
	PorcentajeCompletado float64 `json:"porcentajeCompletado"`
}

type ReservaPorDiaSemana struct {
	DiaSemana            string  `json:"diaSemana"`
	TotalReservas        int64   `json:"totalReservas"`
	ReservasCompletadas  int64   `json:"reservasCompletadas"`
	TotalPersonas        int64   `json:"totalPersonas"`
	PorcentajeCompletado float64 `json:"porcentajeCompletado"`
}

type EstadisticasReservas struct {
	TotalReservasCompletadas   int64   `json:"totalReservasCompletadas"`
	DiaMasReservas             string  `json:"diaMasReservas"`
	HoraMasReservas            string  `json:"horaMasReservas"`
	PromedioPersonasPorReserva float64 `json:"promedioPersonasPorReserva"`
	TasaCompletamiento         float64 `json:"tasaCompletamiento"`
}

// PedidosAnalisisData representa el análisis de pedidos por días y horas
type PedidosAnalisisData struct {
	PedidosPorDia       []PedidoPorDia       `json:"pedidosPorDia"`
	PedidosPorHora      []PedidoPorHora      `json:"pedidosPorHora"`
	PedidosPorDiaSemana []PedidoPorDiaSemana `json:"pedidosPorDiaSemana"`
	EstadisticasPedidos EstadisticasPedidos  `json:"estadisticasPedidos"`
}

type PedidoPorDia struct {
	Fecha              string  `json:"fecha"`
	TotalPedidos       int64   `json:"totalPedidos"`
	PedidosTerminados  int64   `json:"pedidosTerminados"`
	IngresoTotal       int64   `json:"ingresoTotal"`
	TasaCompletamiento float64 `json:"tasaCompletamiento"`
}

type PedidoPorHora struct {
	Hora               string  `json:"hora"`
	TotalPedidos       int64   `json:"totalPedidos"`
	PedidosTerminados  int64   `json:"pedidosTerminados"`
	IngresoTotal       int64   `json:"ingresoTotal"`
	TasaCompletamiento float64 `json:"tasaCompletamiento"`
}

type PedidoPorDiaSemana struct {
	DiaSemana          string  `json:"diaSemana"`
	TotalPedidos       int64   `json:"totalPedidos"`
	PedidosTerminados  int64   `json:"pedidosTerminados"`
	IngresoTotal       int64   `json:"ingresoTotal"`
	TasaCompletamiento float64 `json:"tasaCompletamiento"`
}

type EstadisticasPedidos struct {
	TotalPedidosTerminados    int64   `json:"totalPedidosTerminados"`
	DiaMasPedidos             string  `json:"diaMasPedidos"`
	HoraMasPedidos            string  `json:"horaMasPedidos"`
	IngresoPromedioHora       float64 `json:"ingresoPromedioHora"`
	TasaCompletamientoGeneral float64 `json:"tasaCompletamientoGeneral"`
}

// validateAdminRole valida que el usuario tenga rol de administrador
func (c *TelemetriaController) validateAdminRole() (*Claims, bool) {
	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token no proporcionado",
		}
		_ = c.ServeJSON()
		return nil, false
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	tokenString := authHeader[len("Bearer "):]

	// Usar la función del LoginController para parsear el token
	claims, err := loginc.ParseTokenClaims(tokenString)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return nil, false
	}

	// Validar que sea administrador
	if claims.Rol != string(models.RolAdministrador) {
		c.Ctx.Output.SetStatus(http.StatusForbidden)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusForbidden,
			Message: "Acceso denegado: se requiere rol de administrador",
		}
		_ = c.ServeJSON()
		return nil, false
	}

	return claims, true
}

// @Title GetDashboard
// @Summary Dashboard General de Telemetría
// @Description Obtiene métricas generales del dashboard usando datos existentes: total de pedidos, ingresos, usuarios, promedios y métricas del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.DashboardData} "Dashboard obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/dashboard [get]
func (c *TelemetriaController) GetDashboard() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var dashboardData DashboardData

	// Total de pedidos en el período
	var totalPedidos int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM pedido pe WHERE %s", dateFilter)).QueryRow(&totalPedidos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener total de pedidos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	dashboardData.TotalPedidos = totalPedidos

	// Total de ingresos en el período
	var totalIngresos int64
	err = o.Raw(fmt.Sprintf(`
		SELECT COALESCE(SUM(p.monto), 0)
		FROM pago p
		INNER JOIN pedido pe ON p.pk_id_pago = pe.pk_id_pago
		WHERE %s AND p.estado_pago = 'PAGADO'
	`, dateFilter)).QueryRow(&totalIngresos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener total de ingresos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	dashboardData.TotalIngresos = totalIngresos

	// Total de usuarios (clientes) - siempre histórico
	totalUsuarios, err := o.QueryTable("cliente").Count()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener total de usuarios",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	dashboardData.TotalUsuarios = totalUsuarios

	// Promedio de venta por pedido en el período
	if totalPedidos > 0 {
		dashboardData.PromedioVentaPedido = float64(totalIngresos) / float64(totalPedidos)
	}

	// Pedidos de hoy (siempre del día actual)
	today := time.Now().Format("2006-01-02")
	pedidosHoy, err := o.QueryTable("pedido").Filter("fecha", today).Count()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pedidos de hoy",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	dashboardData.PedidosHoy = pedidosHoy

	// Ingresos de hoy (siempre del día actual)
	var ingresosHoy int64
	err = o.Raw(`
		SELECT COALESCE(SUM(p.monto), 0)
		FROM pago p
		INNER JOIN pedido pe ON p.pk_id_pago = pe.pk_id_pago
		WHERE pe.fecha = ? AND p.estado_pago = 'PAGADO'
	`, today).QueryRow(&ingresosHoy)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ingresos de hoy",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	dashboardData.IngresosHoy = ingresosHoy

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Dashboard (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    dashboardData,
	}
	_ = c.ServeJSON()
}

// @Title GetSales
// @Summary Análisis de Ventas
// @Description Obtiene análisis detallado de ventas usando datos existentes: ventas por método de pago, tendencias y estadísticas generales del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.SalesData} "Análisis de ventas obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/sales [get]
func (c *TelemetriaController) GetSales() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var salesData SalesData

	// Ventas por método de pago en el período
	var ventasPorMetodo []VentaPorMetodo
	_, err := o.Raw(fmt.Sprintf(`
		SELECT
			mp.tipo as metodo_pago,
			COALESCE(SUM(p.monto), 0) as total,
			COUNT(p.pk_id_pago) as cantidad
		FROM metodo_pago mp
		LEFT JOIN pago p ON mp.pk_id_metodo_pago = p.pk_id_metodo_pago AND p.estado_pago = 'PAGADO'
		LEFT JOIN pedido pe ON p.pk_id_pago = pe.pk_id_pago
		WHERE pe.pk_id_pedido IS NULL OR (%s)
		GROUP BY mp.tipo
		ORDER BY total DESC
	`, dateFilter)).QueryRows(&ventasPorMetodo)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ventas por método de pago",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	salesData.VentasPorMetodoPago = ventasPorMetodo

	// Tendencia de ventas por fecha en el período
	var tendenciaVentas []VentaPorFecha
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pe.fecha::text as fecha,
			COALESCE(SUM(p.monto), 0) as total,
			COUNT(pe.pk_id_pedido) as cantidad
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY pe.fecha
		ORDER BY pe.fecha DESC
	`, dateFilter)).QueryRows(&tendenciaVentas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener tendencia de ventas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	salesData.TendenciaVentas = tendenciaVentas

	// Estadísticas generales del período
	var estadisticas EstadisticasVentas
	err = o.Raw(fmt.Sprintf(`
		SELECT
			COALESCE(AVG(daily_sales.total), 0) as venta_promedio_diaria,
			COALESCE(AVG(daily_sales.cantidad), 0) as pedido_promedio_diario,
			CASE
				WHEN SUM(daily_sales.cantidad) > 0
				THEN SUM(daily_sales.total)::float / SUM(daily_sales.cantidad)
				ELSE 0
			END as ticket_promedio
		FROM (
			SELECT
				pe.fecha,
				COALESCE(SUM(p.monto), 0) as total,
				COUNT(pe.pk_id_pedido) as cantidad
			FROM pedido pe
			LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
			WHERE %s
			GROUP BY pe.fecha
		) daily_sales
	`, dateFilter)).QueryRow(&estadisticas.VentaPromedioDiaria, &estadisticas.PedidoPromedioDiario, &estadisticas.TicketPromedio)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de ventas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	salesData.EstadisticasGenerales = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de ventas (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    salesData,
	}
	_ = c.ServeJSON()
}

// @Title GetProducts
// @Summary Análisis de Productos
// @Description Obtiene análisis de productos más y menos vendidos basado en datos de pedidos del período seleccionado, incluyendo estadísticas generales de productos.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de productos a mostrar" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.ProductsData} "Análisis de productos obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/products [get]
func (c *TelemetriaController) GetProducts() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var productsData ProductsData

	// Productos más vendidos en el período
	var productosMasVendidos []ProductoVendido
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pr.pk_id_producto as producto_id,
			pr.nombre as nombre_producto,
			COALESCE(SUM(dp.cantidad), 0) as cantidad_vendida,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as ingreso_total,
			pr.precio as precio
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pr.estado_producto = 'DISPONIBLE'
		AND (pe.pk_id_pedido IS NULL OR (%s))
		GROUP BY pr.pk_id_producto, pr.nombre, pr.precio
		ORDER BY cantidad_vendida DESC, ingreso_total DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&productosMasVendidos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos más vendidos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	productsData.ProductosMasVendidos = productosMasVendidos

	// Productos menos vendidos en el período
	var productosMenosVendidos []ProductoVendido
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pr.pk_id_producto as producto_id,
			pr.nombre as nombre_producto,
			COALESCE(SUM(dp.cantidad), 0) as cantidad_vendida,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as ingreso_total,
			pr.precio as precio
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pr.estado_producto = 'DISPONIBLE'
		AND (pe.pk_id_pedido IS NULL OR (%s))
		GROUP BY pr.pk_id_producto, pr.nombre, pr.precio
		ORDER BY cantidad_vendida ASC, ingreso_total ASC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&productosMenosVendidos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos menos vendidos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	productsData.ProductosMenosVendidos = productosMenosVendidos

	// Estadísticas de productos
	var estadisticas EstadisticasProductos
	totalProductosActivos, err := o.QueryTable("producto").Filter("estado_producto", "DISPONIBLE").Count()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener total de productos activos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	estadisticas.TotalProductosActivos = totalProductosActivos

	// Producto con más ventas
	if len(productosMasVendidos) > 0 {
		estadisticas.ProductoConMasVentas = productosMasVendidos[0].NombreProducto
	}

	// Producto con menos ventas
	if len(productosMenosVendidos) > 0 {
		estadisticas.ProductoConMenosVentas = productosMenosVendidos[0].NombreProducto
	}

	productsData.EstadisticasProductos = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de productos (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    productsData,
	}
	_ = c.ServeJSON()
}

// @Title GetUsers
// @Summary Análisis de Usuarios
// @Description Obtiene análisis de usuarios frecuentes e inactivos basado en historial de pedidos del período seleccionado, incluyendo estadísticas generales de clientes.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de usuarios a mostrar" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.UsersData} "Análisis de usuarios obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/users [get]
func (c *TelemetriaController) GetUsers() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var usersData UsersData

	// Usuarios frecuentes (con más pedidos en el período)
	var usuariosFrecuentes []UsuarioFrecuente
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			COUNT(pe.pk_id_pedido) as total_pedidos,
			COALESCE(SUM(p.monto), 0) as total_gastado,
			MAX(pe.fecha)::text as ultimo_pedido
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pe.pk_id_pedido IS NULL OR (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		HAVING COUNT(pe.pk_id_pedido) > 0
		ORDER BY total_pedidos DESC, total_gastado DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&usuariosFrecuentes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener usuarios frecuentes",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	usersData.UsuariosFrecuentes = usuariosFrecuentes

	// Usuarios inactivos (con pocos pedidos o sin pedidos en el período)
	var usuariosInactivos []UsuarioInactivo
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			COUNT(pe.pk_id_pedido) as total_pedidos,
			COALESCE(MAX(pe.fecha)::text, 'Nunca') as ultimo_pedido
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		WHERE pe.pk_id_pedido IS NULL OR NOT (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		ORDER BY total_pedidos ASC, MAX(pe.fecha) ASC NULLS FIRST
		LIMIT ?
	`, dateFilter), limit).QueryRows(&usuariosInactivos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener usuarios inactivos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	usersData.UsuariosInactivos = usuariosInactivos

	// Estadísticas de usuarios
	var estadisticas EstadisticasUsuarios

	// Total de clientes
	totalClientes, err := o.QueryTable("cliente").Count()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener total de clientes",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	estadisticas.TotalClientes = totalClientes

	// Clientes activos (con pedidos en los últimos 30 días)
	err = o.Raw(`
		SELECT COUNT(DISTINCT pe.pk_documento_cliente)
		FROM pedido pe
		WHERE pe.fecha >= CURRENT_DATE - INTERVAL '30 days'
		AND pe.pk_documento_cliente IS NOT NULL
	`).QueryRow(&estadisticas.ClientesActivos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes activos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	estadisticas.ClientesInactivos = totalClientes - estadisticas.ClientesActivos

	// Promedio de gasto por cliente
	err = o.Raw(`
		SELECT COALESCE(AVG(cliente_gasto.total_gastado), 0)
		FROM (
			SELECT
				c.pk_documento_cliente,
				COALESCE(SUM(p.monto), 0) as total_gastado
			FROM cliente c
			LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
			LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
			GROUP BY c.pk_documento_cliente
		) cliente_gasto
	`).QueryRow(&estadisticas.PromedioGastoPorCliente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener promedio de gasto por cliente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	usersData.EstadisticasUsuarios = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de usuarios (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    usersData,
	}
	_ = c.ServeJSON()
}

// @Title GetTimeAnalysis
// @Summary Análisis Temporal
// @Description Obtiene análisis temporal de pedidos y ventas del período seleccionado: por hora del día, día de la semana y por mes.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.TimeAnalysisData} "Análisis temporal obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/time-analysis [get]
func (c *TelemetriaController) GetTimeAnalysis() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var timeData TimeAnalysisData

	// Ventas por hora del día en el período
	var ventasPorHora []VentaPorHora
	_, err := o.Raw(fmt.Sprintf(`
		SELECT
			EXTRACT(HOUR FROM pe.hora) as hora,
			COALESCE(SUM(p.monto), 0) as total,
			COUNT(pe.pk_id_pedido) as cantidad
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY EXTRACT(HOUR FROM pe.hora)
		ORDER BY hora
	`, dateFilter)).QueryRows(&ventasPorHora)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ventas por hora",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	timeData.VentasPorHora = ventasPorHora

	// Ventas por día de la semana en el período
	var ventasPorDiaSemana []VentaPorDiaSemana
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			CASE EXTRACT(DOW FROM pe.fecha)
				WHEN 0 THEN 'Domingo'
				WHEN 1 THEN 'Lunes'
				WHEN 2 THEN 'Martes'
				WHEN 3 THEN 'Miércoles'
				WHEN 4 THEN 'Jueves'
				WHEN 5 THEN 'Viernes'
				WHEN 6 THEN 'Sábado'
			END as dia_semana,
			COALESCE(SUM(p.monto), 0) as total,
			COUNT(pe.pk_id_pedido) as cantidad
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY EXTRACT(DOW FROM pe.fecha)
		ORDER BY EXTRACT(DOW FROM pe.fecha)
	`, dateFilter)).QueryRows(&ventasPorDiaSemana)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ventas por día de la semana",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	timeData.VentasPorDiaSemana = ventasPorDiaSemana

	// Ventas por mes en el período
	var ventasPorMes []VentaPorMes
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			TO_CHAR(pe.fecha, 'YYYY-MM') as mes,
			COALESCE(SUM(p.monto), 0) as total,
			COUNT(pe.pk_id_pedido) as cantidad
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY TO_CHAR(pe.fecha, 'YYYY-MM')
		ORDER BY mes DESC
	`, dateFilter)).QueryRows(&ventasPorMes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ventas por mes",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	timeData.VentasPorMes = ventasPorMes

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis temporal (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    timeData,
	}
	_ = c.ServeJSON()
}

// === NUEVOS ENDPOINTS PARA MÉTRICAS AVANZADAS ===

// @Title GetRentabilidad
// @Summary Análisis de Rentabilidad por Producto
// @Description Obtiene análisis detallado de rentabilidad de productos: margen de ganancia, productos más y menos rentables del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de productos a mostrar" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.RentabilidadData} "Análisis de rentabilidad obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/rentabilidad [get]
func (c *TelemetriaController) GetRentabilidad() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var rentabilidadData RentabilidadData

	// Productos más rentables (asumiendo 70% de margen como ejemplo)
	var productosRentables []ProductoRentabilidad
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pr.pk_id_producto as producto_id,
			pr.nombre as nombre_producto,
			pr.precio as precio_venta,
			COALESCE(SUM(dp.cantidad), 0) as cantidad_vendida,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as ingreso_total,
			-- Asumimos 70%% de margen para cálculo (en producción usar tabla de costos)
			70.0 as margen_ganancia,
			COALESCE(SUM(dp.precio * dp.cantidad * 0.70), 0) as ganancia_total
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pr.estado_producto = 'DISPONIBLE'
		AND (pe.pk_id_pedido IS NULL OR (%s))
		GROUP BY pr.pk_id_producto, pr.nombre, pr.precio
		HAVING COALESCE(SUM(dp.cantidad), 0) > 0
		ORDER BY ganancia_total DESC, margen_ganancia DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&productosRentables)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos rentables",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	rentabilidadData.ProductosRentables = productosRentables

	// Productos menos rentables
	var productosMenosRentables []ProductoRentabilidad
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pr.pk_id_producto as producto_id,
			pr.nombre as nombre_producto,
			pr.precio as precio_venta,
			COALESCE(SUM(dp.cantidad), 0) as cantidad_vendida,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as ingreso_total,
			-- Asumimos diferentes márgenes para variedad
			CASE
				WHEN pr.precio < 5000 THEN 60.0
				WHEN pr.precio < 15000 THEN 65.0
				ELSE 70.0
			END as margen_ganancia,
			COALESCE(SUM(dp.precio * dp.cantidad *
				CASE
					WHEN pr.precio < 5000 THEN 0.60
					WHEN pr.precio < 15000 THEN 0.65
					ELSE 0.70
				END), 0) as ganancia_total
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pr.estado_producto = 'DISPONIBLE'
		AND (pe.pk_id_pedido IS NULL OR (%s))
		GROUP BY pr.pk_id_producto, pr.nombre, pr.precio
		HAVING COALESCE(SUM(dp.cantidad), 0) > 0
		ORDER BY ganancia_total ASC, margen_ganancia ASC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&productosMenosRentables)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos menos rentables",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	rentabilidadData.ProductosMenosRentables = productosMenosRentables

	// Estadísticas de rentabilidad
	var estadisticas EstadisticasRentabilidad
	err = o.Raw(fmt.Sprintf(`
		SELECT
			AVG(CASE
				WHEN pr.precio < 5000 THEN 60.0
				WHEN pr.precio < 15000 THEN 65.0
				ELSE 70.0
			END) as margen_promedio_general,
			COALESCE(SUM(dp.precio * dp.cantidad *
				CASE
					WHEN pr.precio < 5000 THEN 0.60
					WHEN pr.precio < 15000 THEN 0.65
					ELSE 0.70
				END), 0) as total_ganancias,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as total_ingresos
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pr.estado_producto = 'DISPONIBLE'
		AND (pe.pk_id_pedido IS NULL OR (%s))
	`, dateFilter)).QueryRow(&estadisticas.MargenPromedioGeneral, &estadisticas.TotalGanancias, &estadisticas.TotalIngresos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de rentabilidad",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener producto más y menos rentable
	if len(productosRentables) > 0 {
		estadisticas.ProductoMasRentable = productosRentables[0].NombreProducto
	}
	if len(productosMenosRentables) > 0 {
		estadisticas.ProductoMenosRentable = productosMenosRentables[0].NombreProducto
	}

	rentabilidadData.EstadisticasRentabilidad = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de rentabilidad (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    rentabilidadData,
	}
	_ = c.ServeJSON()
}

// @Title GetSegmentacion
// @Summary Análisis de Segmentación de Clientes
// @Description Obtiene segmentación de clientes en VIP, regulares, ocasionales y nuevos basado en frecuencia y valor de compras del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de clientes por segmento" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.SegmentacionData} "Análisis de segmentación obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/segmentacion [get]
func (c *TelemetriaController) GetSegmentacion() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var segmentacionData SegmentacionData

	// Clientes VIP (más de 5 pedidos y más de $50,000 gastados)
	var clientesVIP []ClienteSegmento
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			COUNT(pe.pk_id_pedido) as total_pedidos,
			COALESCE(SUM(p.monto), 0) as total_gastado,
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 0
				THEN COALESCE(SUM(p.monto), 0)::float / COUNT(pe.pk_id_pedido)
				ELSE 0
			END as promedio_gasto,
			COALESCE(MAX(pe.fecha)::text, 'Nunca') as ultimo_pedido,
			COALESCE(CURRENT_DATE - MAX(pe.fecha), 0) as dias_sin_pedir,
			'VIP' as segmento,
			COALESCE(SUM(p.monto) * 2, 0) as valor_vida -- CLV estimado (2x gasto actual)
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pe.pk_id_pedido IS NULL OR (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		HAVING COUNT(pe.pk_id_pedido) > 5 AND COALESCE(SUM(p.monto), 0) > 50000
		ORDER BY total_gastado DESC, total_pedidos DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&clientesVIP)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes VIP",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	segmentacionData.ClientesVIP = clientesVIP

	// Clientes Regulares (2-5 pedidos)
	var clientesRegulares []ClienteSegmento
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			COUNT(pe.pk_id_pedido) as total_pedidos,
			COALESCE(SUM(p.monto), 0) as total_gastado,
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 0
				THEN COALESCE(SUM(p.monto), 0)::float / COUNT(pe.pk_id_pedido)
				ELSE 0
			END as promedio_gasto,
			COALESCE(MAX(pe.fecha)::text, 'Nunca') as ultimo_pedido,
			COALESCE(CURRENT_DATE - MAX(pe.fecha), 0) as dias_sin_pedir,
			'Regular' as segmento,
			COALESCE(SUM(p.monto) * 1.5, 0) as valor_vida
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pe.pk_id_pedido IS NULL OR (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		HAVING COUNT(pe.pk_id_pedido) BETWEEN 2 AND 5
		ORDER BY total_gastado DESC, total_pedidos DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&clientesRegulares)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes regulares",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	segmentacionData.ClientesRegulares = clientesRegulares

	// Clientes Ocasionales (1 pedido)
	var clientesOcasionales []ClienteSegmento
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			COUNT(pe.pk_id_pedido) as total_pedidos,
			COALESCE(SUM(p.monto), 0) as total_gastado,
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 0
				THEN COALESCE(SUM(p.monto), 0)::float / COUNT(pe.pk_id_pedido)
				ELSE 0
			END as promedio_gasto,
			COALESCE(MAX(pe.fecha)::text, 'Nunca') as ultimo_pedido,
			COALESCE(CURRENT_DATE - MAX(pe.fecha), 0) as dias_sin_pedir,
			'Ocasional' as segmento,
			COALESCE(SUM(p.monto) * 1.2, 0) as valor_vida
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pe.pk_id_pedido IS NULL OR (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		HAVING COUNT(pe.pk_id_pedido) = 1
		ORDER BY total_gastado DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&clientesOcasionales)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes ocasionales",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	segmentacionData.ClientesOcasionales = clientesOcasionales

	// Clientes Nuevos (sin pedidos en el período)
	var clientesNuevos []ClienteSegmento
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			c.pk_documento_cliente as documento_cliente,
			CONCAT(c.nombre, ' ', c.apellido) as nombre_completo,
			0 as total_pedidos,
			0 as total_gastado,
			0 as promedio_gasto,
			'Nunca' as ultimo_pedido,
			999 as dias_sin_pedir,
			'Nuevo' as segmento,
			0 as valor_vida
		FROM cliente c
		LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
		WHERE pe.pk_id_pedido IS NULL OR NOT (%s)
		GROUP BY c.pk_documento_cliente, c.nombre, c.apellido
		HAVING COUNT(pe.pk_id_pedido) = 0
		ORDER BY c.nombre
		LIMIT ?
	`, dateFilter), limit).QueryRows(&clientesNuevos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes nuevos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	segmentacionData.ClientesNuevos = clientesNuevos

	// Estadísticas de segmentación
	var estadisticas EstadisticasSegmentacion
	err = o.Raw(fmt.Sprintf(`
		SELECT
			COUNT(CASE WHEN pedidos > 5 AND gasto > 50000 THEN 1 END) as total_clientes_vip,
			COUNT(CASE WHEN pedidos BETWEEN 2 AND 5 THEN 1 END) as total_clientes_regulares,
			COUNT(CASE WHEN pedidos = 1 THEN 1 END) as total_clientes_ocasionales,
			COUNT(CASE WHEN pedidos = 0 THEN 1 END) as total_clientes_nuevos,
			AVG(CASE WHEN pedidos > 5 AND gasto > 50000 THEN gasto END) as promedio_gasto_vip,
			AVG(CASE WHEN pedidos BETWEEN 2 AND 5 THEN gasto END) as promedio_gasto_regular
		FROM (
			SELECT
				c.pk_documento_cliente,
				COUNT(pe.pk_id_pedido) as pedidos,
				COALESCE(SUM(p.monto), 0) as gasto
			FROM cliente c
			LEFT JOIN pedido pe ON c.pk_documento_cliente = pe.pk_documento_cliente
			LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
			WHERE pe.pk_id_pedido IS NULL OR (%s)
			GROUP BY c.pk_documento_cliente
		) cliente_stats
	`, dateFilter)).QueryRow(
		&estadisticas.TotalClientesVIP,
		&estadisticas.TotalClientesRegulares,
		&estadisticas.TotalClientesOcasionales,
		&estadisticas.TotalClientesNuevos,
		&estadisticas.PromedioGastoVIP,
		&estadisticas.PromedioGastoRegular,
	)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de segmentación",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Calcular porcentaje VIP
	totalClientes := estadisticas.TotalClientesVIP + estadisticas.TotalClientesRegulares +
		estadisticas.TotalClientesOcasionales + estadisticas.TotalClientesNuevos
	if totalClientes > 0 {
		estadisticas.PorcentajeVIP = float64(estadisticas.TotalClientesVIP) / float64(totalClientes) * 100
	}

	segmentacionData.EstadisticasSegmentacion = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de segmentación (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    segmentacionData,
	}
	_ = c.ServeJSON()
}

// @Title GetEficiencia
// @Summary Análisis de Eficiencia de Entregas
// @Description Obtiene análisis de eficiencia operacional: tiempos de entrega, rendimiento de trabajadores y análisis por horas del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de registros por sección" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.EficienciaData} "Análisis de eficiencia obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/eficiencia [get]
func (c *TelemetriaController) GetEficiencia() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var eficienciaData EficienciaData

	// Tiempos de entrega (simulamos tiempo de preparación basado en la diferencia de horas)
	var tiemposEntrega []TiempoEntrega
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pe.pk_id_pedido as pedido_id,
			CONCAT(c.nombre, ' ', c.apellido) as cliente,
			pe.fecha::text as fecha_pedido,
			pe.hora::text as hora_pedido,
			-- Simulamos tiempo de preparación basado en la hora (30-90 minutos)
			CASE
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 12 AND 14 THEN 60 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 19 AND 21 THEN 45 + (EXTRACT(MINUTE FROM pe.hora) %% 45)
				ELSE 30 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
			END as tiempo_preparacion,
			pe.estado_pedido::text as estado_pedido,
			COALESCE(CONCAT(t.nombre, ' ', t.apellido), 'No asignado') as trabajador_asignado
		FROM pedido pe
		LEFT JOIN cliente c ON pe.pk_documento_cliente = c.pk_documento_cliente
		LEFT JOIN domicilio d ON pe.pk_id_domicilio = d.pk_id_domicilio
		LEFT JOIN trabajador t ON d.pk_documento_trabajador = t.pk_documento_trabajador
		WHERE %s
		ORDER BY pe.fecha DESC, pe.hora DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&tiemposEntrega)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener tiempos de entrega",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	eficienciaData.TiemposEntrega = tiemposEntrega

	// Rendimiento de trabajadores (basado en domicilios asignados)
	var rendimientoTrabajadores []RendimientoTrabajador
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			t.pk_documento_trabajador as documento_trabajador,
			CONCAT(t.nombre, ' ', t.apellido) as nombre_trabajador,
			COUNT(pe.pk_id_pedido) as pedidos_atendidos,
			-- Tiempo promedio simulado
			AVG(CASE
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 12 AND 14 THEN 60 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 19 AND 21 THEN 45 + (EXTRACT(MINUTE FROM pe.hora) %% 45)
				ELSE 30 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
			END) as tiempo_promedio_atencion,
			-- Score de eficiencia (1-10) basado en pedidos y tiempo
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 20 THEN 9.0 + (RANDOM() * 1)
				WHEN COUNT(pe.pk_id_pedido) > 10 THEN 7.0 + (RANDOM() * 2)
				WHEN COUNT(pe.pk_id_pedido) > 5 THEN 6.0 + (RANDOM() * 2)
				ELSE 4.0 + (RANDOM() * 3)
			END as eficiencia_score,
			-- Horas trabajadas simuladas (8 horas por día trabajado)
			COUNT(DISTINCT pe.fecha) * 8.0 as horas_trabajadas
		FROM trabajador t
		LEFT JOIN domicilio d ON t.pk_documento_trabajador = d.pk_documento_trabajador
		LEFT JOIN pedido pe ON d.pk_id_domicilio = pe.pk_id_domicilio
		WHERE t.fecha_retiro IS NULL
		AND (pe.pk_id_pedido IS NULL OR (%s))
		GROUP BY t.pk_documento_trabajador, t.nombre, t.apellido
		HAVING COUNT(pe.pk_id_pedido) > 0
		ORDER BY eficiencia_score DESC, pedidos_atendidos DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&rendimientoTrabajadores)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener rendimiento de trabajadores",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	eficienciaData.RendimientoTrabajadores = rendimientoTrabajadores

	// Análisis por hora
	var analisisPorHora []EficienciaPorHora
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			EXTRACT(HOUR FROM pe.hora)::text || ':00' as hora,
			COUNT(pe.pk_id_pedido) as pedidos_recibidos,
			AVG(CASE
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 12 AND 14 THEN 60 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 19 AND 21 THEN 45 + (EXTRACT(MINUTE FROM pe.hora) %% 45)
				ELSE 30 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
			END) as tiempo_promedio_prep,
			-- Capacidad utilizada (simulada basada en pedidos por hora)
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 15 THEN 95.0 + (RANDOM() * 5)
				WHEN COUNT(pe.pk_id_pedido) > 10 THEN 75.0 + (RANDOM() * 20)
				WHEN COUNT(pe.pk_id_pedido) > 5 THEN 50.0 + (RANDOM() * 25)
				ELSE 20.0 + (RANDOM() * 30)
			END as capacidad_utilizada,
			-- Nivel de eficiencia
			CASE
				WHEN COUNT(pe.pk_id_pedido) > 15 THEN 'Alto'
				WHEN COUNT(pe.pk_id_pedido) > 8 THEN 'Medio'
				ELSE 'Bajo'
			END as nivel_eficiencia
		FROM pedido pe
		WHERE %s
		GROUP BY EXTRACT(HOUR FROM pe.hora)
		ORDER BY EXTRACT(HOUR FROM pe.hora)
	`, dateFilter)).QueryRows(&analisisPorHora)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener análisis por hora",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	eficienciaData.AnalisisPorHora = analisisPorHora

	// Estadísticas generales de eficiencia
	var estadisticas EstadisticasEficiencia
	err = o.Raw(fmt.Sprintf(`
		SELECT
			AVG(CASE
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 12 AND 14 THEN 60 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
				WHEN EXTRACT(HOUR FROM pe.hora) BETWEEN 19 AND 21 THEN 45 + (EXTRACT(MINUTE FROM pe.hora) %% 45)
				ELSE 30 + (EXTRACT(MINUTE FROM pe.hora) %% 30)
			END) as tiempo_promedio_general,
			COUNT(CASE WHEN pe.estado_pedido NOT IN ('TERMINADO', 'CANCELADO') THEN 1 END) as pedidos_pendientes
		FROM pedido pe
		WHERE %s
	`, dateFilter)).QueryRow(&estadisticas.TiempoPromedioGeneral, &estadisticas.PedidosPendientes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de eficiencia",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener hora más y menos eficiente
	if len(analisisPorHora) > 0 {
		maxEficiencia := 0.0
		minEficiencia := 100.0
		for _, hora := range analisisPorHora {
			if hora.CapacidadUtilizada > maxEficiencia {
				maxEficiencia = hora.CapacidadUtilizada
				estadisticas.HoraMasEficiente = hora.Hora
			}
			if hora.CapacidadUtilizada < minEficiencia {
				minEficiencia = hora.CapacidadUtilizada
				estadisticas.HoraMenosEficiente = hora.Hora
			}
		}

		// Calcular capacidad promedio de uso
		totalCapacidad := 0.0
		for _, hora := range analisisPorHora {
			totalCapacidad += hora.CapacidadUtilizada
		}
		if len(analisisPorHora) > 0 {
			estadisticas.CapacidadPromedioUso = totalCapacidad / float64(len(analisisPorHora))
		}
	}

	// Obtener trabajador más eficiente
	if len(rendimientoTrabajadores) > 0 {
		estadisticas.TrabajadorMasEficiente = rendimientoTrabajadores[0].NombreTrabajador
	}

	eficienciaData.EstadisticasEficiencia = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de eficiencia (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    eficienciaData,
	}
	_ = c.ServeJSON()
}

// @Title GetReservasAnalisis
// @Summary Análisis de Reservas por Días y Horas
// @Description Obtiene análisis detallado de reservas completadas por días y horas: identificar patrones de reservas más exitosas del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de registros por sección" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.ReservasAnalisisData} "Análisis de reservas obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/reservas-analisis [get]
func (c *TelemetriaController) GetReservasAnalisis() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var reservasData ReservasAnalisisData

	// Reservas por día
	var reservasPorDia []ReservaPorDia
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			r.fecha::text as fecha,
			COUNT(*) as total_reservas,
			COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END) as reservas_completadas,
			SUM(r.personas) as total_personas,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as porcentaje_completado
		FROM reserva r
		WHERE %s
		GROUP BY r.fecha
		ORDER BY reservas_completadas DESC, total_reservas DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&reservasPorDia)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas por día",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	reservasData.ReservasPorDia = reservasPorDia

	// Reservas por hora
	var reservasPorHora []ReservaPorHora
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			EXTRACT(HOUR FROM r.hora)::text || ':00' as hora,
			COUNT(*) as total_reservas,
			COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END) as reservas_completadas,
			SUM(r.personas) as total_personas,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as porcentaje_completado
		FROM reserva r
		WHERE %s
		GROUP BY EXTRACT(HOUR FROM r.hora)
		ORDER BY reservas_completadas DESC, total_reservas DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&reservasPorHora)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas por hora",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	reservasData.ReservasPorHora = reservasPorHora

	// Reservas por día de la semana
	var reservasPorDiaSemana []ReservaPorDiaSemana
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			CASE EXTRACT(DOW FROM r.fecha)
				WHEN 0 THEN 'Domingo'
				WHEN 1 THEN 'Lunes'
				WHEN 2 THEN 'Martes'
				WHEN 3 THEN 'Miércoles'
				WHEN 4 THEN 'Jueves'
				WHEN 5 THEN 'Viernes'
				WHEN 6 THEN 'Sábado'
			END as dia_semana,
			COUNT(*) as total_reservas,
			COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END) as reservas_completadas,
			SUM(r.personas) as total_personas,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as porcentaje_completado
		FROM reserva r
		WHERE %s
		GROUP BY EXTRACT(DOW FROM r.fecha)
		ORDER BY reservas_completadas DESC, total_reservas DESC
	`, dateFilter)).QueryRows(&reservasPorDiaSemana)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas por día de la semana",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	reservasData.ReservasPorDiaSemana = reservasPorDiaSemana

	// Estadísticas de reservas
	var estadisticas EstadisticasReservas
	err = o.Raw(fmt.Sprintf(`
		SELECT
			COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END) as total_reservas_completadas,
			AVG(r.personas) as promedio_personas_por_reserva,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN r.estado_reserva = 'CUMPLIDA' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as tasa_completamiento
		FROM reserva r
		WHERE %s
	`, dateFilter)).QueryRow(&estadisticas.TotalReservasCompletadas, &estadisticas.PromedioPersonasPorReserva, &estadisticas.TasaCompletamiento)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de reservas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener día y hora con más reservas
	if len(reservasPorDia) > 0 {
		estadisticas.DiaMasReservas = reservasPorDia[0].Fecha
	}
	if len(reservasPorHora) > 0 {
		estadisticas.HoraMasReservas = reservasPorHora[0].Hora
	}

	reservasData.EstadisticasReservas = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de reservas (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    reservasData,
	}
	_ = c.ServeJSON()
}

// @Title GetPedidosAnalisis
// @Summary Análisis de Pedidos por Días y Horas
// @Description Obtiene análisis detallado de pedidos realizados por días y horas: identificar patrones de pedidos más exitosos y horarios de mayor demanda del período seleccionado.
// @Tags telemetria
// @Accept json
// @Produce json
// @Param limit query int false "Límite de registros por sección" minimum(1) maximum(100) default(10)
// @Param periodo query string false "Período de tiempo" Enums(hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico) default(ultimo_mes)
// @Success 200 {object} models.ApiResponse{data=telemetria.PedidosAnalisisData} "Análisis de pedidos obtenido exitosamente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 403 {object} models.ApiResponse "Acceso denegado - se requiere rol de administrador"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /telemetria/pedidos-analisis [get]
func (c *TelemetriaController) GetPedidosAnalisis() {
	claims, valid := c.validateAdminRole()
	if !valid {
		return
	}

	// Obtener límite de la query string
	limitStr := c.GetString("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtener filtro de tiempo
	periodoStr := c.GetString("periodo", "ultimo_mes")
	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	var pedidosData PedidosAnalisisData

	// Pedidos por día
	var pedidosPorDia []PedidoPorDia
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			pe.fecha::text as fecha,
			COUNT(*) as total_pedidos,
			COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END) as pedidos_terminados,
			COALESCE(SUM(p.monto), 0) as ingreso_total,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as tasa_completamiento
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY pe.fecha
		ORDER BY pedidos_terminados DESC, total_pedidos DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&pedidosPorDia)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pedidos por día",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	pedidosData.PedidosPorDia = pedidosPorDia

	// Pedidos por hora
	var pedidosPorHora []PedidoPorHora
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			EXTRACT(HOUR FROM pe.hora)::text || ':00' as hora,
			COUNT(*) as total_pedidos,
			COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END) as pedidos_terminados,
			COALESCE(SUM(p.monto), 0) as ingreso_total,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as tasa_completamiento
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY EXTRACT(HOUR FROM pe.hora)
		ORDER BY pedidos_terminados DESC, total_pedidos DESC
		LIMIT ?
	`, dateFilter), limit).QueryRows(&pedidosPorHora)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pedidos por hora",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	pedidosData.PedidosPorHora = pedidosPorHora

	// Pedidos por día de la semana
	var pedidosPorDiaSemana []PedidoPorDiaSemana
	_, err = o.Raw(fmt.Sprintf(`
		SELECT
			CASE EXTRACT(DOW FROM pe.fecha)
				WHEN 0 THEN 'Domingo'
				WHEN 1 THEN 'Lunes'
				WHEN 2 THEN 'Martes'
				WHEN 3 THEN 'Miércoles'
				WHEN 4 THEN 'Jueves'
				WHEN 5 THEN 'Viernes'
				WHEN 6 THEN 'Sábado'
			END as dia_semana,
			COUNT(*) as total_pedidos,
			COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END) as pedidos_terminados,
			COALESCE(SUM(p.monto), 0) as ingreso_total,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as tasa_completamiento
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
		GROUP BY EXTRACT(DOW FROM pe.fecha)
		ORDER BY pedidos_terminados DESC, total_pedidos DESC
	`, dateFilter)).QueryRows(&pedidosPorDiaSemana)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pedidos por día de la semana",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	pedidosData.PedidosPorDiaSemana = pedidosPorDiaSemana

	// Estadísticas de pedidos
	var estadisticas EstadisticasPedidos
	err = o.Raw(fmt.Sprintf(`
		SELECT
			COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END) as total_pedidos_terminados,
			CASE
				WHEN COUNT(*) > 0
				THEN (COUNT(CASE WHEN pe.estado_pedido = 'TERMINADO' THEN 1 END)::float / COUNT(*)) * 100
				ELSE 0
			END as tasa_completamiento_general,
			COALESCE(AVG(p.monto), 0) as ingreso_promedio_hora
		FROM pedido pe
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE %s
	`, dateFilter)).QueryRow(&estadisticas.TotalPedidosTerminados, &estadisticas.TasaCompletamientoGeneral, &estadisticas.IngresoPromedioHora)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener estadísticas de pedidos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener día y hora con más pedidos
	if len(pedidosPorDia) > 0 {
		estadisticas.DiaMasPedidos = pedidosPorDia[0].Fecha
	}
	if len(pedidosPorHora) > 0 {
		estadisticas.HoraMasPedidos = pedidosPorHora[0].Hora
	}

	pedidosData.EstadisticasPedidos = estadisticas

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Análisis de pedidos (%s) obtenido exitosamente por %s", string(periodo), claims.Nombre),
		Data:    pedidosData,
	}
	_ = c.ServeJSON()
}

// ProductosPopularesData estructura para productos más vendidos (público)
type ProductosPopularesData struct {
	ProductosPopulares []ProductoVendido `json:"productosPopulares"`
}

// GetProductosPopulares obtiene los productos más vendidos (endpoint público)
// @Title GetProductosPopulares
// @Description Obtiene los productos más vendidos - endpoint público sin autenticación
// @Param limit query int false "Número de productos a retornar (default: 4)"
// @Param periodo query string false "Filtro temporal: hoy, ultima_semana, ultimo_mes, ultimos_3_meses, ultimos_6_meses, ultimo_año, historico"
// @Success 200 {object} models.ApiResponse{data=ProductosPopularesData}
// @Failure 500 {object} models.ApiResponse
// @router /productos-populares [get]
func (c *TelemetriaController) GetProductosPopulares() {
	// Obtener parámetros de consulta
	limitStr := c.GetString("limit", "4") // Default 4 productos
	periodoStr := c.GetString("periodo", "ultimo_mes")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 4
	}

	periodo := TimeFilter(periodoStr)
	startDate, endDate := getTimeRange(periodo)
	dateFilter := buildDateFilter(startDate, endDate)

	o := orm.NewOrm()

	// Consulta para productos más vendidos
	var productosMasVendidos []ProductoVendido
	sql := fmt.Sprintf(`
		SELECT
			pr.pk_id_producto as producto_id,
			pr.nombre as nombre_producto,
			COALESCE(SUM(dp.cantidad), 0) as cantidad_vendida,
			COALESCE(SUM(dp.precio * dp.cantidad), 0) as ingreso_total,
			pr.precio,
			COALESCE(encode(pr.imagen, 'base64'), '') as imagen
		FROM producto pr
		LEFT JOIN detalle_pedido dp ON pr.pk_id_producto = dp.pk_id_producto
		LEFT JOIN pedido pe ON dp.pk_id_pedido = pe.pk_id_pedido
		LEFT JOIN pago p ON pe.pk_id_pago = p.pk_id_pago AND p.estado_pago = 'PAGADO'
		WHERE pe.estado_pedido = 'TERMINADO' AND %s
		GROUP BY pr.pk_id_producto, pr.nombre, pr.precio, pr.imagen
		ORDER BY cantidad_vendida DESC
		LIMIT %d
	`, dateFilter, limit)

	_, err = o.Raw(sql).QueryRows(&productosMasVendidos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos populares",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	productosData := ProductosPopularesData{
		ProductosPopulares: productosMasVendidos,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Productos populares (%s) obtenidos exitosamente", string(periodo)),
		Data:    productosData,
	}
	_ = c.ServeJSON()
}
