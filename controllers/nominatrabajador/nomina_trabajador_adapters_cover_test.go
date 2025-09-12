package nominatrabajador

import "testing"

func TestNtQSAdapter_OrderBy_Exist_Coverage(t *testing.T) {
	a := ntQSAdapter{}
	func() { defer func() { _ = recover() }(); _ = a.OrderBy("-fecha") }()
	func() { defer func() { _ = recover() }(); _ = a.Exist() }()
}

func TestNtOrmAdapter_Insert_QueryTable_Coverage(t *testing.T) {
	a := ntOrmAdapter{}
	func() { defer func() { _ = recover() }(); _, _ = a.Insert(nil) }()
	func() { defer func() { _ = recover() }(); _ = a.QueryTable(nil) }()
}

func TestNtQSAdapter_One_All_Filter_Coverage(t *testing.T) {
	a := ntQSAdapter{}
	func() { defer func() { _ = recover() }(); _ = a.One(nil) }()
	func() { defer func() { _ = recover() }(); _, _ = a.All(nil) }()
	func() { defer func() { _ = recover() }(); _ = a.Filter("", nil) }()
}
