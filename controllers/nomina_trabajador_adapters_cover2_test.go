package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

func TestNtQSAdapter_OrderBy_Exist_WithRealOrm(t *testing.T) {
	// Mock Exist() underlying query to return at least one row
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery = origQ })

	qs := orm.NewOrm().QueryTable(new(models.Nomina))
	a := ntQSAdapter{qs: qs}
	_ = a.OrderBy("-fecha")
	_ = a.Exist()
}

func TestNtOrmAdapter_Insert_WithRealOrm(t *testing.T) {
	origE := MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockExec = origE })

	o := orm.NewOrm()
	a := ntOrmAdapter{o: o}
	_, _ = a.Insert(&models.Nomina{})
}
