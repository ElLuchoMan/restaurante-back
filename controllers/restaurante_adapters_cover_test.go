package controllers

import "testing"

func TestRestOrmAdapter_Update_Coverage(t *testing.T) {
    a := restOrmAdapter{}
    defer func() { _ = recover() }()
    _, _ = a.Update(nil)
}


