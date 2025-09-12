package test

import (
	"fmt"
	"os"
	"restaurante/models"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

func getOrmer() (orm.Ormer, error) {
	var o orm.Ormer
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		o = orm.NewOrm()
	}()
	return o, err
}

func TestDetallePedidoTrigger(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	o, err := getOrmer()
	if err != nil {
		t.Fatalf("orm not available: %v", err)
	}
	var d models.DetallePedido
	if err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", 1).Filter("PKIDProducto", 1).One(&d); err != nil {
		t.Fatalf("detalle_pedido not found or DB unavailable: %v", err)
	}
	if d.Precio == 0 {
		t.Errorf("expected precio to be set by trigger, got 0")
	}
}

func TestDetallePedidoTriggerOnInsert(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	o, err := getOrmer()
	if err != nil {
		t.Fatalf("orm not available: %v", err)
	}
	var prod models.Producto
	if err := o.QueryTable(new(models.Producto)).Filter("PK_ID_PRODUCTO", 1).One(&prod); err != nil {
		t.Fatalf("producto not found or DB unavailable: %v", err)
	}
	pid := int64(9999)
	prodID := prod.PK_ID_PRODUCTO
	det := models.DetallePedido{
		PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pid},
		PKIDProducto: &models.Producto{PK_ID_PRODUCTO: prodID},
		Cantidad:     1,
	}
	if _, err := o.Insert(&det); err != nil {
		t.Fatalf("insert detalle_pedido failed: %v", err)
	}
	defer o.Delete(&det)
	if err := o.Read(&det); err != nil {
		t.Fatalf("read detalle_pedido failed: %v", err)
	}
	if det.Precio != prod.PRECIO {
		t.Errorf("expected precio %d, got %d", prod.PRECIO, det.Precio)
	}
}

func TestSetPrecioDetallePedidoTriggerExists(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	o, err := getOrmer()
	if err != nil {
		t.Fatalf("orm not available: %v", err)
	}
	var count int
	if err := o.Raw("SELECT COUNT(*) FROM pg_trigger WHERE tgname = ?", "set_precio_detalle_pedido").QueryRow(&count); err != nil {
		t.Fatalf("cannot query trigger existence: %v", err)
	}
	if count == 0 {
		t.Fatalf("trigger set_precio_detalle_pedido not found")
	}
}
