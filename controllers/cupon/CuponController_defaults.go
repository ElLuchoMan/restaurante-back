//go:build !unit

package cupon

import "github.com/beego/beego/v2/client/orm"

func defaultOrmProvider() orm.Ormer {
	return orm.NewOrm()
}
