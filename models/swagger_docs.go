package models

// Estructuras SOLO para documentación Swagger: reflejan exactamente lo que
// producen los MarshalJSON de los modelos (nombres json, fechas como string,
// relaciones embebidas como objeto, campos nulos). No se usan en tiempo de
// ejecución.

// ClienteRefDoc es el objeto cliente embebido (nunca incluye la contraseña).
// En listados solo `documentoCliente` es significativo: el resto de campos vienen
// vacíos porque la relación no se carga.
type ClienteRefDoc struct {
	DocumentoCliente int64   `json:"documentoCliente" example:"1001"`
	Nombre           string  `json:"nombre"`
	Apellido         string  `json:"apellido"`
	Correo           string  `json:"correo"`
	Direccion        string  `json:"direccion"`
	Telefono         string  `json:"telefono"`
	Observaciones    *string `json:"observaciones" extensions:"x-nullable"`
}

// RestauranteRefDoc es el objeto restaurante embebido. En listados solo
// `restauranteId` es significativo (nombreRestaurante vacío y horaApertura sin
// valor real porque la relación no se carga).
type RestauranteRefDoc struct {
	RestauranteId     int64  `json:"restauranteId" example:"1"`
	NombreRestaurante string `json:"nombreRestaurante"`
	HoraApertura      string `json:"horaApertura" example:"08:00:00"`
	CambioHorarioId   any    `json:"cambioHorarioId,omitempty" swaggertype:"object"`
}

// PedidoRefDoc es el objeto pedido embebido (sin contraseñas). En listados solo
// `pedidoId` es significativo salvo que el endpoint indique lo contrario.
type PedidoRefDoc struct {
	PedidoId         int64              `json:"pedidoId" example:"10"`
	FechaPedido      string             `json:"fechaPedido" example:"31-01-2025"`
	HoraPedido       string             `json:"horaPedido" example:"18:30:00"`
	Delivery         bool               `json:"delivery"`
	EstadoPedido     string             `json:"estadoPedido"`
	DomicilioId      any                `json:"domicilioId,omitempty" swaggertype:"object"`
	PagoId           any                `json:"pagoId" swaggertype:"object" extensions:"x-nullable"`
	RestauranteId    *RestauranteRefDoc `json:"restauranteId" extensions:"x-nullable"`
	DocumentoCliente *ClienteRefDoc     `json:"documentoCliente" extensions:"x-nullable"`
	UpdatedAt        string             `json:"updatedAt" example:"31-01-2025 18:30:00"`
	UpdatedBy        *string            `json:"updatedBy,omitempty"`
}

// ProductoDoc es la forma JSON de Producto (ver Producto.MarshalJSON).
type ProductoDoc struct {
	ProductoId     int64   `json:"productoId" example:"1"`
	Nombre         string  `json:"nombre" example:"Bandeja Paisa"`
	Calorias       *int64  `json:"calorias" extensions:"x-nullable" example:"850"`
	Descripcion    *string `json:"descripcion,omitempty" example:"Plato típico"`
	Precio         int64   `json:"precio" example:"25000"`
	EstadoProducto string  `json:"estadoProducto" enums:"DISPONIBLE,NO_DISPONIBLE" example:"DISPONIBLE"`
	Imagen         string  `json:"imagen,omitempty" example:"iVBORw0KGgo..."`
	Cantidad       int     `json:"cantidad" example:"10"`
	SubcategoriaId *int64  `json:"subcategoriaId" extensions:"x-nullable" example:"1"`
}

// PrecioProductoHistDoc es la forma JSON de PrecioProductoHist.
type PrecioProductoHistDoc struct {
	PrecioHistId  int64  `json:"precioHistId" example:"1"`
	ProductoId    *int64 `json:"productoId" example:"1"`
	Precio        int64  `json:"precio" example:"25000"`
	FechaVigencia string `json:"fechaVigencia" example:"31-01-2025"`
}

// PrecioHistItemDoc es cada elemento de GET /precio_producto_hist y
// GET /precio_producto_hist/search.
type PrecioHistItemDoc struct {
	PrecioHistId   int64  `json:"precioHistId" example:"1"`
	ProductoId     int64  `json:"productoId" example:"1"`
	Nombre         string `json:"nombre" example:"Bandeja Paisa"`
	EstadoProducto string `json:"estadoProducto" enums:"DISPONIBLE,NO_DISPONIBLE" example:"DISPONIBLE"`
	Precio         int64  `json:"precio" example:"25000"`
	FechaVigencia  string `json:"fechaVigencia" example:"31-01-2025"`
}

// OfertaDoc es la forma JSON de Oferta (ver Oferta.MarshalJSON).
type OfertaDoc struct {
	OfertaId       int64              `json:"ofertaId" example:"1"`
	Titulo         string             `json:"titulo" example:"Martes de gaseosas"`
	TipoDescuento  string             `json:"tipoDescuento" enums:"PORCENTAJE,MONTO" example:"PORCENTAJE"`
	ValorDescuento int64              `json:"valorDescuento" example:"30"`
	FechaInicio    string             `json:"fechaInicio" example:"01-01-2025"`
	FechaFin       string             `json:"fechaFin" example:"31-12-2025"`
	DiasSemana     []string           `json:"diasSemana" enums:"Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo"`
	HoraInicio     *string            `json:"horaInicio,omitempty" example:"08:00:00"`
	HoraFin        *string            `json:"horaFin,omitempty" example:"18:00:00"`
	Activo         bool               `json:"activo"`
	RestauranteId  *RestauranteRefDoc `json:"restauranteId"`
}

// CuponDoc es la forma JSON de Cupon (ver Cupon.MarshalJSON). Las relaciones
// opcionales se omiten cuando son nulas.
type CuponDoc struct {
	CuponId          int64          `json:"cuponId" example:"1"`
	Codigo           string         `json:"codigo" example:"VERANO10"`
	Scope            string         `json:"scope" enums:"GLOBAL,PRODUCTO,CATEGORIA,CLIENTE" example:"GLOBAL"`
	TipoDescuento    string         `json:"tipoDescuento" enums:"PORCENTAJE,MONTO" example:"PORCENTAJE"`
	ValorDescuento   int64          `json:"valorDescuento" example:"10"`
	MaxUsos          *int           `json:"maxUsos,omitempty" example:"100"`
	LimitePorCliente *int           `json:"limitePorCliente,omitempty" example:"1"`
	MontoMinimo      *int64         `json:"montoMinimo,omitempty" example:"20000"`
	FechaInicio      string         `json:"fechaInicio" example:"01-01-2025"`
	FechaFin         string         `json:"fechaFin" example:"31-12-2025"`
	ProductoId       *ProductoDoc   `json:"productoId,omitempty"`
	CategoriaId      *Categoria     `json:"categoriaId,omitempty"`
	DocumentoCliente *ClienteRefDoc `json:"documentoCliente,omitempty"`
	Activo           bool           `json:"activo"`
}

// CuponRedencionDoc es la forma JSON de CuponRedencion.
type CuponRedencionDoc struct {
	CuponRedencionId int64          `json:"cuponRedencionId" example:"1"`
	CuponId          *CuponDoc      `json:"cuponId"`
	DocumentoCliente *ClienteRefDoc `json:"documentoCliente"`
	PedidoId         *PedidoRefDoc  `json:"pedidoId,omitempty"`
	MontoDescuento   int64          `json:"montoDescuento" example:"5000"`
	CreatedAt        string         `json:"createdAt" example:"31-01-2025 18:30:00"`
}

// PedidoDescuentoDoc es la forma JSON de PedidoDescuentoAplicado.
type PedidoDescuentoDoc struct {
	PedidoDescuentoId int64         `json:"pedidoDescuentoId" example:"1"`
	PedidoId          *PedidoRefDoc `json:"pedidoId"`
	CuponId           *CuponDoc     `json:"cuponId,omitempty"`
	OfertaId          *OfertaDoc    `json:"ofertaId,omitempty"`
	MontoDescuento    int64         `json:"montoDescuento" example:"5000"`
	Detalle           any           `json:"detalle,omitempty" swaggertype:"object"`
	CreatedAt         string        `json:"createdAt" example:"31-01-2025 18:30:00"`
}

// OfertaPaginadaDoc, CuponPaginadoDoc y CuponRedencionPaginadaDoc documentan la
// forma de PaginatedResponse para cada listado.
type OfertaPaginadaDoc struct {
	Data       []OfertaDoc `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

type CuponPaginadoDoc struct {
	Data       []CuponDoc `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

type CuponRedencionPaginadaDoc struct {
	Data       []CuponRedencionDoc `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalPages int                 `json:"totalPages"`
}
