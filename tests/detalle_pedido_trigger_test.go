package test

import (
	"github.com/beego/beego/v2/client/orm"
	"restaurante/models"
	"testing"
)

func TestDetallePedidoTrigger(t *testing.T) {
	o := orm.NewOrm()
	var d models.DetallePedido
	err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", 1).Filter("PKIDProducto", 1).One(&d)
	if err != nil {
		t.Skipf("detalle_pedido not found or DB unavailable: %v", err)
	}
	if d.Precio == 0 {
		t.Errorf("expected precio to be set by trigger, got 0")
	}
}
