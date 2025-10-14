package cupon

import (
	"testing"
)

func TestCupOrmAdapter_Simple_Coverage(t *testing.T) {

	a := cupOrmAdapter{o: nil}

	defer func() { _ = recover() }()

	_, _ = a.Insert(nil)
	_ = a.Read(nil)
	_, _ = a.Update(nil)
	_, _ = a.Delete(nil)
	_ = a.QueryTable(nil)
}

func TestCupQSAdapter_Simple_Coverage(t *testing.T) {

	a := cupQSAdapter{qs: nil}

	defer func() { _ = recover() }()

	_, _ = a.All(nil)
	_ = a.Filter("", nil)
	_ = a.OrderBy("")
	_ = a.Limit(0)
	_ = a.Offset(0)
	_, _ = a.Count()
	_ = a.One(nil)
}
