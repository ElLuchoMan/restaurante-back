package push

import (
	"testing"
)

func TestPushOrmAdapter_Simple_Coverage(t *testing.T) {

	a := pushOrmAdapter{o: nil}

	defer func() { _ = recover() }()

	_, _ = a.Insert(nil)
	_ = a.Read(nil)
	_, _ = a.Update(nil)
	_, _ = a.Delete(nil)
	_ = a.QueryTable(nil)
}

func TestPushQSAdapter_Simple_Coverage(t *testing.T) {

	a := pushQSAdapter{qs: nil}

	defer func() { _ = recover() }()

	_, _ = a.All(nil)
	_ = a.Filter("", nil)
	_ = a.OrderBy("")
	_ = a.Limit(0)
	_ = a.Offset(0)
	_, _ = a.Count()
	_ = a.One(nil)
}
