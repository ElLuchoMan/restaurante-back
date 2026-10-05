package models

// Estructuras SOLO para documentación Swagger del grupo pedidos / pagos /
// domicilios / restaurante. Reflejan exactamente lo que producen los
// MarshalJSON de los modelos (ver swagger_docs_pedidos_test.go, que lo
// verifica). Las claves de relación (`pagoId`, `domicilioId`, `metodoPagoId`,
// `restauranteId`, `documentoCliente`, `trabajadorAsignado`, ...) NO son
// números: el API serializa el objeto relacionado, en el que solo el id es
// fiable salvo que el endpoint indique lo contrario.

// MetodoPagoDoc es un método de pago (ver MetodoPago).
type MetodoPagoDoc struct {
	MetodoPagoId int64  `json:"metodoPagoId" example:"1"`
	Tipo         string `json:"tipo" example:"NEQUI"`
	Detalle      string `json:"detalle" example:"Cuenta 3001234567"`
}

// PagoDoc es la forma JSON de Pago (ver Pago.MarshalJSON). `fechaPago` va como
// DD-MM-YYYY, `horaPago` como HH:MM:SS y `updatedAt` como DD-MM-YYYY HH:MM:SS
// (hora de Bogotá). `metodoPagoId` es el método de pago como objeto.
type PagoDoc struct {
	PagoId       int64          `json:"pagoId" example:"4"`
	FechaPago    string         `json:"fechaPago" example:"31-01-2025"`
	HoraPago     string         `json:"horaPago" example:"14:30:00"`
	Monto        int64          `json:"monto" example:"50000"`
	EstadoPago   string         `json:"estadoPago" enums:"PAGADO,PENDIENTE,NO_PAGO" example:"PAGADO"`
	MetodoPagoId *MetodoPagoDoc `json:"metodoPagoId"`
	UpdatedAt    string         `json:"updatedAt" example:"31-01-2025 14:30:00"`
	UpdatedBy    *string        `json:"updatedBy,omitempty" example:"operador@example.com"`
}

// DomicilioDoc es la forma JSON de Domicilio (ver Domicilio.MarshalJSON).
// `fechaDomicilio` va como DD-MM-YYYY. `trabajadorAsignado` es el trabajador
// como objeto (sin contraseña) y se omite si no hay domiciliario; en él solo
// `documentoTrabajador` es significativo.
type DomicilioDoc struct {
	DomicilioId        int64               `json:"domicilioId" example:"3"`
	Direccion          string              `json:"direccion" example:"Calle 123 #45-67"`
	Telefono           string              `json:"telefono" example:"3001234567"`
	EstadoDomicilio    string              `json:"estadoDomicilio" enums:"PENDIENTE,EN_CAMINO,ENTREGADO" example:"PENDIENTE"`
	Entregado          bool                `json:"entregado" example:"false"`
	FechaDomicilio     string              `json:"fechaDomicilio" example:"31-01-2025"`
	Observaciones      *string             `json:"observaciones,omitempty" example:"Dejar en portería"`
	CreatedAt          string              `json:"createdAt" example:"31-01-2025 18:30:00"`
	UpdatedAt          string              `json:"updatedAt" example:"31-01-2025 18:30:00"`
	CreatedBy          *string             `json:"createdBy,omitempty" example:"admin@example.com"`
	UpdatedBy          *string             `json:"updatedBy,omitempty" example:"operador@example.com"`
	TrabajadorAsignado *TrabajadorResponse `json:"trabajadorAsignado,omitempty"`
}

// PedidoDoc es la forma JSON de Pedido (ver Pedido.MarshalJSON). `fechaPedido`
// va como DD-MM-YYYY, `horaPedido` como HH:MM:SS y `updatedAt` como
// DD-MM-YYYY HH:MM:SS. Las relaciones son objetos (o null si el pedido no las
// tiene; `domicilioId` se omite): en ellos solo el id es fiable.
type PedidoDoc struct {
	PedidoId         int64              `json:"pedidoId" example:"10"`
	FechaPedido      string             `json:"fechaPedido" example:"31-01-2025"`
	HoraPedido       string             `json:"horaPedido" example:"18:30:00"`
	Delivery         bool               `json:"delivery" example:"false"`
	EstadoPedido     string             `json:"estadoPedido" enums:"INICIADO,EN_PREPARACION,LISTO,TERMINADO,CANCELADO" example:"INICIADO"`
	DomicilioId      *DomicilioDoc      `json:"domicilioId,omitempty"`
	PagoId           *PagoDoc           `json:"pagoId" extensions:"x-nullable"`
	RestauranteId    *RestauranteRefDoc `json:"restauranteId" extensions:"x-nullable"`
	DocumentoCliente *ClienteRefDoc     `json:"documentoCliente" extensions:"x-nullable"`
	UpdatedAt        string             `json:"updatedAt" example:"31-01-2025 18:30:00"`
	UpdatedBy        *string            `json:"updatedBy,omitempty"`
}

// DetallePedidoDoc es una línea de pedido (ver DetallePedido). `pedidoId` y
// `productoId` son objetos de la relación en los que solo el id es fiable;
// `precio` es el precio unitario fijado por la base de datos al insertar.
type DetallePedidoDoc struct {
	DetalleId  int64       `json:"detalleId" example:"1"`
	PedidoId   PedidoDoc   `json:"pedidoId"`
	ProductoId ProductoDoc `json:"productoId"`
	Precio     int64       `json:"precio" example:"25000"`
	Cantidad   int         `json:"cantidad" example:"2"`
}

// ProductoPedidoDoc es el `data` de GET/POST/PUT /producto_pedido.
type ProductoPedidoDoc struct {
	PedidoId int64              `json:"pedidoId" example:"10"`
	Detalles []DetallePedidoDoc `json:"detalles"`
}

// InventarioInsuficienteDoc es cada elemento del `data` de la respuesta 409 de
// POST/PUT /producto_pedido.
type InventarioInsuficienteDoc struct {
	ProductoId int64 `json:"productoId" example:"1"`
	Requerido  int   `json:"requerido" example:"5"`
	Disponible int   `json:"disponible" example:"2"`
}

// DomicilioClienteDoc es el cliente del pedido asociado a un domicilio.
type DomicilioClienteDoc struct {
	Documento int64  `json:"documento" example:"1234567890"`
	Nombre    string `json:"nombre" example:"Juan"`
	Apellido  string `json:"apellido" example:"Pérez"`
}

// DetalleProductoDoc es cada producto del resumen de pedido (`pedido.productos`
// de GET /domicilios/search y el JSON de `productos` de GET /pedidos/detalles).
type DetalleProductoDoc struct {
	PkIdProducto int64  `json:"pk_id_producto" example:"1"`
	Nombre       string `json:"nombre" example:"Bandeja Paisa"`
	Cantidad     int    `json:"cantidad" example:"2"`
	Precio       int64  `json:"precio" example:"25000"`
	Subtotal     int64  `json:"subtotal" example:"50000"`
}

// DomicilioPedidoDoc es el resumen del último pedido asociado a un domicilio.
type DomicilioPedidoDoc struct {
	PedidoId          int64                `json:"pedidoId" example:"10"`
	PagoId            *int64               `json:"pagoId" extensions:"x-nullable" example:"4"`
	MontoPago         float64              `json:"montoPago" example:"50000"`
	SubtotalProductos float64              `json:"subtotalProductos" example:"50000"`
	Total             float64              `json:"total" example:"50000"`
	Productos         []DetalleProductoDoc `json:"productos"`
}

// DomicilioDetalleDoc es el `data` de GET /domicilios/search. `cliente` y
// `pedido` solo vienen si hay un pedido asociado al domicilio.
type DomicilioDetalleDoc struct {
	Domicilio DomicilioDoc         `json:"domicilio"`
	Cliente   *DomicilioClienteDoc `json:"cliente,omitempty"`
	Pedido    *DomicilioPedidoDoc  `json:"pedido,omitempty"`
}

// RestauranteDoc es la forma JSON de Restaurante (ver Restaurante.MarshalJSON).
// `cambioHorarioId` es el cambio de horario como objeto y se omite si no hay.
type RestauranteDoc struct {
	RestauranteId     int64                   `json:"restauranteId" example:"1"`
	NombreRestaurante string                  `json:"nombreRestaurante" example:"El Fogón de María"`
	HoraApertura      string                  `json:"horaApertura" example:"08:00:00"`
	CambioHorarioId   *CambiosHorarioResponse `json:"cambioHorarioId,omitempty"`
}
