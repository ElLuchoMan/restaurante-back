//go:build unit

package descuento

import "github.com/beego/beego/v2/client/orm"

func defaultOrmReadProvider() func(interface{}, ...string) error {
	return func(interface{}, ...string) error { return nil }
}

func defaultOrmProvider() orm.Ormer {
	return nil
}
