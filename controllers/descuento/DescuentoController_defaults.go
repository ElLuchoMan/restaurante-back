//go:build !unit

package descuento

import "github.com/beego/beego/v2/client/orm"

func defaultOrmReadProvider() func(interface{}, ...string) error {
	return orm.NewOrm().Read
}

func defaultOrmProvider() orm.Ormer {
	return orm.NewOrm()
}
