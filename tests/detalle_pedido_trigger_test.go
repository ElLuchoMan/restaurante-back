package test

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"restaurante/models"
	"testing"
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
	o, err := getOrmer()
	if err != nil {
		t.Skipf("orm not available: %v", err)
	}
	var d models.DetallePedido
	if err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", 1).Filter("PKIDProducto", 1).One(&d); err != nil {
		t.Skipf("detalle_pedido not found or DB unavailable: %v", err)
	}
	if d.Precio == 0 {
		t.Errorf("expected precio to be set by trigger, got 0")
	}
}

func TestDetallePedidoTriggerOnInsert(t *testing.T) {
	o, err := getOrmer()
	if err != nil {
		t.Skipf("orm not available: %v", err)
	}
	var prod models.Producto
	if err := o.QueryTable(new(models.Producto)).Filter("PK_ID_PRODUCTO", 1).One(&prod); err != nil {
		t.Skipf("producto not found or DB unavailable: %v", err)
	}
	det := models.DetallePedido{PKIDPedido: 9999, PKIDProducto: prod.PK_ID_PRODUCTO, Cantidad: 1}
	if _, err := o.Insert(&det); err != nil {
		t.Skipf("insert detalle_pedido failed: %v", err)
	}
	defer o.Delete(&det)
	if err := o.Read(&det); err != nil {
		t.Skipf("read detalle_pedido failed: %v", err)
	}
	if det.Precio != float64(prod.PRECIO) {
		t.Errorf("expected precio %.2f, got %.2f", float64(prod.PRECIO), det.Precio)
	}
}
