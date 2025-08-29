package test

import (
	"log"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"restaurante/models"
)

// SeedTestData inserta registros básicos para las pruebas.
func SeedTestData() {
	o := orm.NewOrm()

	if _, err := o.Insert(&models.MetodoPago{PK_ID_METODO_PAGO: 1, TIPO: "Efectivo"}); err != nil {
		log.Println("seed METODO_PAGO:", err)
	}

	if _, err := o.Insert(&models.ProductoPedidoDetalle{
		PKIDPedido:     1,
		PKIDProducto:   1,
		CANTIDAD:       1,
		PRECIOUNITARIO: 1000,
		SUBTOTAL:       1000,
	}); err != nil {
		log.Println("seed PRODUCTO_PEDIDO_DETALLE:", err)
	}

	if _, err := o.Insert(&models.HorarioTrabajador{
		PK_DOCUMENTO_TRABAJADOR: 1,
		DIA:                     "Lunes",
		HORA_INICIO:             time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		HORA_FIN:                time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC),
	}); err != nil {
		log.Println("seed HORARIO_TRABAJADOR:", err)
	}
}
