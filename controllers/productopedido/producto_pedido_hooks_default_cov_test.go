package productopedido

import (
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

func TestProductoPedido_RequeryDefault_Cover(t *testing.T) {
	tx, err := productoPedidoBeginTx(orm.NewOrm())
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	var out models.DetallePedido
	_ = productoPedidoRequeryDetalle(tx, 1, 1, &out)
}
