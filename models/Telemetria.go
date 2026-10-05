package models

import "encoding/json"

type DashboardData struct {
	TotalPedidos        int64   `json:"totalPedidos"`
	TotalIngresos       int64   `json:"totalIngresos"`
	TotalUsuarios       int64   `json:"totalUsuarios"`
	PromedioVentaPedido float64 `json:"promedioVentaPedido"`
	PedidosHoy          int64   `json:"pedidosHoy"`
	IngresosHoy         int64   `json:"ingresosHoy"`
}

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
	// Imagen en base64; solo se rellena en GET /productos-populares, en el resto de endpoints llega "".
	Imagen string `json:"imagen"`
}

// ProductoDisponible es cada elemento de GET /productos-disponibles.
type ProductoDisponible struct {
	ProductoId     int64  `json:"productoId"`
	NombreProducto string `json:"nombreProducto"`
	Precio         int64  `json:"precio"`
	Estado         string `json:"estado" enums:"DISPONIBLE,NO_DISPONIBLE"`
	// TotalVendido suma las unidades de pedidos en estado TERMINADO.
	TotalVendido int64 `json:"totalVendido"`
}

type EstadisticasProductos struct {
	TotalProductosActivos  int64  `json:"totalProductosActivos"`
	ProductoConMasVentas   string `json:"productoConMasVentas"`
	ProductoConMenosVentas string `json:"productoConMenosVentas"`
}

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
	MargenGanancia  float64 `json:"margenGanancia"`
	GananciaTotal   int64   `json:"gananciaTotal"`
}

type EstadisticasRentabilidad struct {
	MargenPromedioGeneral float64 `json:"margenPromedioGeneral"`
	ProductoMasRentable   string  `json:"productoMasRentable"`
	ProductoMenosRentable string  `json:"productoMenosRentable"`
	TotalGanancias        int64   `json:"totalGanancias"`
	TotalIngresos         int64   `json:"totalIngresos"`
}

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
	ValorVida        int64   `json:"valorVida"`
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
	TiempoPreparacion  int    `json:"tiempoPreparacion"`
	EstadoPedido       string `json:"estadoPedido"`
	TrabajadorAsignado string `json:"trabajadorAsignado"`
}

type RendimientoTrabajador struct {
	DocumentoTrabajador    int64   `json:"documentoTrabajador"`
	NombreTrabajador       string  `json:"nombreTrabajador"`
	PedidosAtendidos       int64   `json:"pedidosAtendidos"`
	TiempoPromedioAtencion float64 `json:"tiempoPromedioAtencion"`
	EficienciaScore        float64 `json:"eficienciaScore"`
	HorasTrabajadas        float64 `json:"horasTrabajadas"`
}

type EficienciaPorHora struct {
	Hora               string  `json:"hora"`
	PedidosRecibidos   int64   `json:"pedidosRecibidos"`
	TiempoPromedioPrep float64 `json:"tiempoPromedioPrep"`
	CapacidadUtilizada float64 `json:"capacidadUtilizada"`
	NivelEficiencia    string  `json:"nivelEficiencia"`
}

type EstadisticasEficiencia struct {
	TiempoPromedioGeneral  float64 `json:"tiempoPromedioGeneral"`
	HoraMasEficiente       string  `json:"horaMasEficiente"`
	HoraMenosEficiente     string  `json:"horaMenosEficiente"`
	TrabajadorMasEficiente string  `json:"trabajadorMasEficiente"`
	CapacidadPromedioUso   float64 `json:"capacidadPromedioUso"`
	PedidosPendientes      int64   `json:"pedidosPendientes"`
}

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

// emptySliceIfNil devuelve s o, si es nil, un slice vacío para que se
// serialice como [] y nunca como null.
func emptySliceIfNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// MarshalJSON garantiza que las listas de SalesData se serialicen como [] y nunca como null.
func (d SalesData) MarshalJSON() ([]byte, error) {
	type plain SalesData
	p := plain(d)
	p.VentasPorMetodoPago = emptySliceIfNil(p.VentasPorMetodoPago)
	p.TendenciaVentas = emptySliceIfNil(p.TendenciaVentas)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de ProductsData se serialicen como [] y nunca como null.
func (d ProductsData) MarshalJSON() ([]byte, error) {
	type plain ProductsData
	p := plain(d)
	p.ProductosMasVendidos = emptySliceIfNil(p.ProductosMasVendidos)
	p.ProductosMenosVendidos = emptySliceIfNil(p.ProductosMenosVendidos)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de UsersData se serialicen como [] y nunca como null.
func (d UsersData) MarshalJSON() ([]byte, error) {
	type plain UsersData
	p := plain(d)
	p.UsuariosFrecuentes = emptySliceIfNil(p.UsuariosFrecuentes)
	p.UsuariosInactivos = emptySliceIfNil(p.UsuariosInactivos)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de TimeAnalysisData se serialicen como [] y nunca como null.
func (d TimeAnalysisData) MarshalJSON() ([]byte, error) {
	type plain TimeAnalysisData
	p := plain(d)
	p.VentasPorHora = emptySliceIfNil(p.VentasPorHora)
	p.VentasPorDiaSemana = emptySliceIfNil(p.VentasPorDiaSemana)
	p.VentasPorMes = emptySliceIfNil(p.VentasPorMes)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de RentabilidadData se serialicen como [] y nunca como null.
func (d RentabilidadData) MarshalJSON() ([]byte, error) {
	type plain RentabilidadData
	p := plain(d)
	p.ProductosRentables = emptySliceIfNil(p.ProductosRentables)
	p.ProductosMenosRentables = emptySliceIfNil(p.ProductosMenosRentables)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de SegmentacionData se serialicen como [] y nunca como null.
func (d SegmentacionData) MarshalJSON() ([]byte, error) {
	type plain SegmentacionData
	p := plain(d)
	p.ClientesVIP = emptySliceIfNil(p.ClientesVIP)
	p.ClientesRegulares = emptySliceIfNil(p.ClientesRegulares)
	p.ClientesOcasionales = emptySliceIfNil(p.ClientesOcasionales)
	p.ClientesNuevos = emptySliceIfNil(p.ClientesNuevos)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de EficienciaData se serialicen como [] y nunca como null.
func (d EficienciaData) MarshalJSON() ([]byte, error) {
	type plain EficienciaData
	p := plain(d)
	p.TiemposEntrega = emptySliceIfNil(p.TiemposEntrega)
	p.RendimientoTrabajadores = emptySliceIfNil(p.RendimientoTrabajadores)
	p.AnalisisPorHora = emptySliceIfNil(p.AnalisisPorHora)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de ReservasAnalisisData se serialicen como [] y nunca como null.
func (d ReservasAnalisisData) MarshalJSON() ([]byte, error) {
	type plain ReservasAnalisisData
	p := plain(d)
	p.ReservasPorDia = emptySliceIfNil(p.ReservasPorDia)
	p.ReservasPorHora = emptySliceIfNil(p.ReservasPorHora)
	p.ReservasPorDiaSemana = emptySliceIfNil(p.ReservasPorDiaSemana)
	return json.Marshal(p)
}

// MarshalJSON garantiza que las listas de PedidosAnalisisData se serialicen como [] y nunca como null.
func (d PedidosAnalisisData) MarshalJSON() ([]byte, error) {
	type plain PedidosAnalisisData
	p := plain(d)
	p.PedidosPorDia = emptySliceIfNil(p.PedidosPorDia)
	p.PedidosPorHora = emptySliceIfNil(p.PedidosPorHora)
	p.PedidosPorDiaSemana = emptySliceIfNil(p.PedidosPorDiaSemana)
	return json.Marshal(p)
}
