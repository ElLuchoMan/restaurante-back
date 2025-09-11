package productopedido

import (
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

func TestProductoPedido_RequeryDefault_DirectCall(t *testing.T) {
	// Usar el Begin (envuelto por TestMain) para obtener un TxOrmer válido
	tx, err := productoPedidoBeginTx(orm.NewOrm())
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	var out models.DetallePedido
	_ = productoPedidoRequeryDetalleDefault(tx, 123, 456, &out)
}

