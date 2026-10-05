package models

// ActualizarOfertaRequest es el cuerpo de PUT /ofertas (merge): los campos
// ausentes se conservan. `horaInicio` y `horaFin` aceptan null explícito para
// quitar el horario (deben limpiarse juntos); el resto no admite null.
// Fechas en YYYY-MM-DD; horas HH:MM o HH:MM:SS.
type ActualizarOfertaRequest struct {
	Titulo          *string        `json:"titulo,omitempty" example:"Martes de gaseosas"`
	TipoDescuento   *TipoDescuento `json:"tipoDescuento,omitempty" enums:"PORCENTAJE,MONTO" example:"PORCENTAJE"`
	ValorDescuento  *int64         `json:"valorDescuento,omitempty" example:"30"`
	FechaInicio     *string        `json:"fechaInicio,omitempty" example:"2025-01-01"`
	FechaFin        *string        `json:"fechaFin,omitempty" example:"2025-12-31"`
	DiasSemana      []string       `json:"diasSemana,omitempty" swaggertype:"array,string" enums:"Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo"`
	HoraInicio      *string        `json:"horaInicio,omitempty" extensions:"x-nullable" example:"08:00"`
	HoraFin         *string        `json:"horaFin,omitempty" extensions:"x-nullable" example:"18:00"`
	PkIdRestaurante *int64         `json:"restauranteId,omitempty" example:"1"`
	Activo          *bool          `json:"activo,omitempty" example:"true"`
}

// ActualizarCuponRequest es el cuerpo de PUT /cupones (merge): los campos
// ausentes se conservan. Admiten null explícito (se limpian): maxUsos,
// limitePorCliente, montoMinimo, productoId, categoriaId y documentoCliente; el
// resto no admite null. Fechas en YYYY-MM-DD.
type ActualizarCuponRequest struct {
	Codigo             *string        `json:"codigo,omitempty" example:"VERANO10"`
	Scope              *CuponScope    `json:"scope,omitempty" enums:"GLOBAL,PRODUCTO,CATEGORIA,CLIENTE" example:"GLOBAL"`
	TipoDescuento      *TipoDescuento `json:"tipoDescuento,omitempty" enums:"PORCENTAJE,MONTO" example:"PORCENTAJE"`
	ValorDescuento     *int64         `json:"valorDescuento,omitempty" example:"10"`
	MaxUsos            *int           `json:"maxUsos,omitempty" extensions:"x-nullable" example:"100"`
	LimitePorCliente   *int           `json:"limitePorCliente,omitempty" extensions:"x-nullable" example:"1"`
	MontoMinimo        *int64         `json:"montoMinimo,omitempty" extensions:"x-nullable" example:"20000"`
	FechaInicio        *string        `json:"fechaInicio,omitempty" example:"2025-01-01"`
	FechaFin           *string        `json:"fechaFin,omitempty" example:"2025-12-31"`
	PkIdProducto       *int64         `json:"productoId,omitempty" extensions:"x-nullable" example:"1"`
	PkIdCategoria      *int64         `json:"categoriaId,omitempty" extensions:"x-nullable" example:"1"`
	PkDocumentoCliente *int64         `json:"documentoCliente,omitempty" extensions:"x-nullable" example:"1001"`
	Activo             *bool          `json:"activo,omitempty" example:"true"`
}

// OfertaProductoAsociacionDoc es el `data` de POST /ofertas/productos.
type OfertaProductoAsociacionDoc struct {
	OfertaId   int64 `json:"ofertaId" example:"1"`
	ProductoId int64 `json:"productoId" example:"2"`
}
