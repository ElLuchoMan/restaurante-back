package controllers

import (
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

// Implementa un QuerySeter mínimo que satisface la interfaz concreta
type minimalQS struct{ orm.QuerySeter }

func (m minimalQS) OrderBy(...string) orm.QuerySeter { return m }
func (m minimalQS) Exist() bool                      { return true }

func TestNtQSAdapter_OrderBy_Exist_NoPanic(t *testing.T) {
	// inyectamos un QuerySeter compatible
	a := ntQSAdapter{qs: minimalQS{}}
	_ = a.OrderBy("-fecha")
	if !a.Exist() {
		t.Fatalf("expected exist true")
	}
}
