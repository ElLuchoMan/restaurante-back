//go:build unit

package push

import "github.com/beego/beego/v2/client/orm"

func defaultOrmProvider() orm.Ormer {
	return nil
}
