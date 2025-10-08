//go:build !unit

package oferta

import "github.com/beego/beego/v2/client/orm"

func defaultOrmProvider() orm.Ormer {
	return orm.NewOrm()
}
