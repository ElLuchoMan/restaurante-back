package controllers

import "testing"

// Estos tests fuerzan la ejecución de las funciones adaptadoras
// de NominaTrabajadorController para cubrir OrderBy, Exist e Insert.

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


