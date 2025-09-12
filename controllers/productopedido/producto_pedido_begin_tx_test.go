package productopedido

import (
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

func TestProductoPedidoBeginTx_Default(t *testing.T) {
	orig := productoPedidoBeginTx
	productoPedidoBeginTx = func(o orm.Ormer) (orm.TxOrmer, error) {
		return o.Begin()
	}
	defer func() { productoPedidoBeginTx = orig }()
	tx, err := productoPedidoBeginTx(orm.NewOrm())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tx.Rollback()
}
