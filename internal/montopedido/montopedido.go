// Package montopedido calcula en el servidor cuánto debe pagar un pedido.
//
// Fórmula (la única que existe en el negocio y en el esquema; no hay costo de
// domicilio, propina ni impuestos):
//
//	subtotal = SUM(detalle_pedido.precio * detalle_pedido.cantidad)
//	total    = MAX(0, subtotal - SUM(pedido_descuento_aplicado.monto_descuento))
//
// `detalle_pedido.precio` lo fija un trigger de la base con el precio vigente del
// producto, así que ningún dato del cliente interviene en el cálculo.
package montopedido

import "github.com/beego/beego/v2/client/orm"

// Monto es el resultado de Calcular.
type Monto struct {
	// Lineas es la cantidad de productos distintos del pedido.
	Lineas int64
	// Subtotal es la suma de precio x cantidad del detalle.
	Subtotal int64
	// Descuento es la suma de los descuentos ya aplicados al pedido.
	Descuento int64
	// Total es lo que debe quedar en pago.monto (nunca negativo).
	Total int64
}

// Calcular obtiene el monto del pedido con q (ORM o transacción). Dentro de una
// transacción con el pedido bloqueado el resultado no cambia hasta el COMMIT.
func Calcular(q orm.QueryExecutor, pedidoID int64) (Monto, error) {
	var m Monto
	if err := q.Raw("SELECT COALESCE(SUM(precio * cantidad), 0), COUNT(*) FROM detalle_pedido WHERE pk_id_pedido = ?", pedidoID).QueryRow(&m.Subtotal, &m.Lineas); err != nil {
		return Monto{}, err
	}
	if err := q.Raw("SELECT COALESCE(SUM(monto_descuento), 0) FROM pedido_descuento_aplicado WHERE pk_id_pedido = ?", pedidoID).QueryRow(&m.Descuento); err != nil {
		return Monto{}, err
	}
	m.Total = max(m.Subtotal-m.Descuento, 0)
	return m, nil
}
