package models

// DashboardData representa los datos del dashboard general de telemetría
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

// === MÉTRICAS AVANZADAS ===

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

// === ANÁLISIS DE RESERVAS ===

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

// === ANÁLISIS DE PEDIDOS ===

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
