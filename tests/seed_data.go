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

	inicio, _ := time.Parse("15:04:05", "08:00:00")
	fin, _ := time.Parse("15:04:05", "16:00:00")
	if _, err := o.Insert(&models.HorarioTrabajador{
		PK_DOCUMENTO_TRABAJADOR: 1,
		DIA:                     "Lunes",
		HORA_INICIO:             inicio,
		HORA_FIN:                fin,
	}); err != nil {
		log.Println("seed HORARIO_TRABAJADOR:", err)
	}
}
