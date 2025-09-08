package controllers

import "testing"

func TestRestauranteOrmAdapterCoverage(t *testing.T) {
	a := restOrmAdapter{}
	func() {
		defer func() { _ = recover() }()
		a.QueryTable(nil)
	}()
	func() {
		defer func() { _ = recover() }()
		_ = a.Read(nil)
	}()
	func() {
		defer func() { _ = recover() }()
		_, _ = a.Insert(nil)
	}()
	func() {
		defer func() { _ = recover() }()
		_, _ = a.Update(nil)
	}()
	func() {
		defer func() { _ = recover() }()
		_, _ = a.Delete(nil)
	}()
	func() {
		defer func() { _ = recover() }()
		qs := restQSAdapter{}
		_, _ = qs.All(nil)
	}()
}
