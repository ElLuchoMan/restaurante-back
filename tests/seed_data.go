//go:build integration
// +build integration

package test

import (
	"log"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// SeedTestData inserta registros básicos para las pruebas.
func SeedTestData() {
	o := orm.NewOrm()

	if _, err := o.Insert(&models.MetodoPago{PK_ID_METODO_PAGO: 1, TIPO: "Efectivo"}); err != nil {
		log.Println("seed metodo_pago:", err)
	}

	if _, err := o.Insert(&models.Categoria{PK_ID_CATEGORIA: 1, NOMBRE: "General"}); err != nil {
		log.Println("seed categoria:", err)
	}

	if _, err := o.Insert(&models.Subcategoria{PK_ID_SUBCATEGORIA: 1, PK_ID_CATEGORIA: &models.Categoria{PK_ID_CATEGORIA: 1}, NOMBRE: "Subgeneral"}); err != nil {
		log.Println("seed subcategoria:", err)
	}

	prod := models.Producto{PK_ID_PRODUCTO: 1}
	if err := o.Read(&prod); err == orm.ErrNoRows {
		desc := "Usado en tests"
		prod = models.Producto{
			PK_ID_PRODUCTO:     1,
			NOMBRE:             "Producto Prueba",
			DESCRIPCION:        &desc,
			PRECIO:             1000,
			ESTADO_PRODUCTO:    models.EstadoProductoDisponible,
			CANTIDAD:           1,
			PK_ID_SUBCATEGORIA: &models.Subcategoria{PK_ID_SUBCATEGORIA: 1},
		}
		if _, err := o.Insert(&prod); err != nil {
			log.Println("seed producto:", err)
		}
	} else if err != nil {
		log.Println("seed producto:", err)
	}

	pedidoID := int64(1)
	productoID := int64(1)
	if _, err := o.Insert(&models.DetallePedido{
		PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
		PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
		Cantidad:     1,
	}); err != nil {
		log.Println("seed detalle_pedido:", err)
	}

	// Crear tiempos con año válido (1) para evitar error de rango en PostgreSQL
	inicio := time.Date(1, 1, 1, 8, 0, 0, 0, time.UTC)
	fin := time.Date(1, 1, 1, 16, 0, 0, 0, time.UTC)
	if _, err := o.Raw(
		"INSERT INTO horario_trabajador (pk_documento_trabajador, dia, hora_inicio, hora_fin) VALUES (?, ?, ?, ?)",
		1, models.DiaLunes, inicio, fin,
	).Exec(); err != nil {
		log.Println("seed horario_trabajador:", err)
	}
}
