package models

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// strictDecode comprueba que el JSON real cabe exactamente en el tipo de
// documentación (sin claves desconocidas) y que no filtra contraseñas.
func strictDecode(t *testing.T, v any, doc any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(doc); err != nil {
		t.Fatalf("el JSON real no coincide con el tipo de documentación: %v\n%s", err, b)
	}
	if bytes.Contains(bytes.ToLower(b), []byte("password")) {
		t.Fatalf("la respuesta no debe incluir contraseñas: %s", b)
	}
	return b
}

func TestPedidoDocsCoincidenConLaSalidaReal(t *testing.T) {
	ts := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)
	tel, by := "300", "op"
	metodo := &MetodoPago{PK_ID_METODO_PAGO: 1, TIPO: "NEQUI", DETALLE: "x"}
	pago := &Pago{PK_ID_PAGO: 4, FECHA: ts, HORA: ts, MONTO: 5, ESTADO_PAGO: EstadoPagoPagado, PK_ID_METODO_PAGO: metodo, UPDATED_AT: ts, UPDATED_BY: &by}
	dom := &Domicilio{ID: 3, Direccion: "d", Telefono: "t", Estado: EstadoDomicilioPendiente, Fecha: ts, Observ: &by, CreatedAt: ts, UpdatedAt: ts, CreatedBy: &by, UpdatedBy: &by,
		Trabajador: &Trabajador{PK_DOCUMENTO_TRABAJADOR: 77, TELEFONO: &tel, PASSWORD: "secreta", FECHA_INGRESO: ts, PK_ID_RESTAURANTE: &Restaurante{PK_ID_RESTAURANTE: 1}}}
	rest := &Restaurante{PK_ID_RESTAURANTE: 1, NOMBRE_RESTAURANTE: "R", HORA_APERTURA: ts, PK_ID_CAMBIO_HORARIO: &CambiosHorario{PK_ID_CAMBIO_HORARIO: 2}}
	cliente := &Cliente{PK_DOCUMENTO_CLIENTE: 1001, PASSWORD: "secreta"}
	pedido := Pedido{PK_ID_PEDIDO: 10, FECHA: ts, HORA: ts, ESTADO_PEDIDO: EstadoPedidoIniciado, PK_ID_DOMICILIO: dom, PK_ID_PAGO: pago,
		PK_ID_RESTAURANTE: rest, PK_DOCUMENTO_CLIENTE: cliente, UPDATED_AT: ts, UPDATED_BY: &by}

	strictDecode(t, MetodoPago{PK_ID_METODO_PAGO: 1, TIPO: "NEQUI"}, &MetodoPagoDoc{})
	strictDecode(t, pago, &PagoDoc{})
	strictDecode(t, dom, &DomicilioDoc{})
	strictDecode(t, Domicilio{ID: 1}, &DomicilioDoc{})
	strictDecode(t, rest, &RestauranteDoc{})
	strictDecode(t, Restaurante{PK_ID_RESTAURANTE: 1}, &RestauranteDoc{})
	strictDecode(t, pedido, &PedidoDoc{})
	strictDecode(t, Pedido{PK_ID_PEDIDO: 1}, &PedidoDoc{})
	strictDecode(t, DetallePedido{PK_ID_DETALLE: 1, PKIDPedido: &pedido, PKIDProducto: &Producto{PK_ID_PRODUCTO: 2}, Precio: 3, Cantidad: 4}, &DetallePedidoDoc{})
	strictDecode(t, PedidoDetails{PedidoID: 1}, &PedidoDetails{})
}
