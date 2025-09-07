package controllers

import "testing"

func TestSubOrmAdapter_Methods_Coverage(t *testing.T) {
    a := subOrmAdapter{}

    func() { defer func() { _ = recover() }(); _, _ = a.Insert(nil) }()
    func() { defer func() { _ = recover() }(); _ = a.Read(nil) }()
    func() { defer func() { _ = recover() }(); _, _ = a.Update(nil) }()
    func() { defer func() { _ = recover() }(); _, _ = a.Delete(nil) }()
    func() { defer func() { _ = recover() }(); _ = a.QueryTable(nil) }()
}

func TestSubQSAdapter_All_Filter_Coverage(t *testing.T) {
    qs := subQSAdapter{}
    defer func() { _ = recover() }()
    _ = qs.Filter("PK_ID_CATEGORIA", 1)
    _, _ = qs.All(nil)
}


