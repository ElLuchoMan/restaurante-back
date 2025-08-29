package test

import (
	"log"

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
		PKIDProductoPedido: 1,
		PKIDProducto:       1,
		CANTIDAD:           1,
		PRECIO:             1000,
	}); err != nil {
		log.Println("seed PRODUCTO_PEDIDO_DETALLE:", err)
	}

	if _, err := o.Insert(&models.HorarioTrabajador{
		DocumentoTrabajador: 1,
		Dia:                 "Lunes",
		HoraInicio:          "08:00:00",
		HoraFin:             "16:00:00",
	}); err != nil {
		log.Println("seed HORARIO_TRABAJADOR:", err)
	}
}
