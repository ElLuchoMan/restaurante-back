package models

import (
	"encoding/json"
	"strconv"
)

// CheckoutDomicilio es el domicilio opcional de POST /pedidos/checkout.
// `direccion` y `telefono` son obligatorios; `fechaDomicilio` (YYYY-MM-DD) por
// defecto es hoy en Bogotá. El estado nace PENDIENTE y no se asigna domiciliario.
type CheckoutDomicilio struct {
	Direccion      string  `json:"direccion" example:"Calle 123 #45-67"`
	Telefono       string  `json:"telefono" example:"3001234567"`
	FechaDomicilio string  `json:"fechaDomicilio,omitempty" example:"2025-01-31"`
	Observaciones  *string `json:"observaciones,omitempty" example:"Dejar en portería"`
}

// CheckoutPago es el pago de POST /pedidos/checkout. El monto lo calcula el
// servidor: `monto` solo lo respeta el personal (ajuste manual > 0) y a un
// Cliente se le ignora.
type CheckoutPago struct {
	MetodoPagoId int64  `json:"metodoPagoId" example:"1"`                                                  // debe existir
	FechaPago    string `json:"fechaPago,omitempty" example:"2025-01-31"`                                  // YYYY-MM-DD; por defecto hoy en Bogotá
	HoraPago     string `json:"horaPago,omitempty" example:"14:30:00"`                                     // HH:MM[:SS]; por defecto ahora en Bogotá
	EstadoPago   string `json:"estadoPago,omitempty" enums:"PAGADO,PENDIENTE,NO_PAGO" example:"PENDIENTE"` // por defecto PENDIENTE (un Cliente solo puede PENDIENTE)
	Monto        int64  `json:"monto,omitempty" example:"0"`                                               // Cliente: se ignora; personal: ajuste manual > 0
}

// CheckoutRequest es el cuerpo de POST /pedidos/checkout.
type CheckoutRequest struct {
	RestauranteId    int64                     `json:"restauranteId,omitempty" example:"1"`             // opcional; si se envía debe existir
	DocumentoCliente *int64                    `json:"documentoCliente,omitempty" example:"1234567890"` // solo personal; un Cliente sale del token
	Domicilio        *CheckoutDomicilio        `json:"domicilio,omitempty"`                             // opcional: si viene, el pedido es delivery
	Productos        []ProductoPedidoItemInput `json:"productos"`                                       // al menos uno, con cantidad > 0
	Pago             CheckoutPago              `json:"pago"`
}

// CheckoutResult es el `data` de la respuesta 201 de POST /pedidos/checkout: el
// pedido completo (con su pago y domicilio) más el `monto` que fijó el servidor.
type CheckoutResult struct {
	Pedido Pedido
	Monto  int64
}

// MarshalJSON serializa el pedido tal cual (ver Pedido.MarshalJSON) y le agrega
// la clave `monto`.
func (r CheckoutResult) MarshalJSON() ([]byte, error) {
	b, _ := json.Marshal(r.Pedido) // Pedido siempre serializa como objeto y no falla
	b = append(b[:len(b)-1], `,"monto":`...)
	b = strconv.AppendInt(b, r.Monto, 10)
	return append(b, '}'), nil
}
