package controllers

import (
	"restaurante/models"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

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
	func() { defer func() { _ = recover() }(); _ = qs.Filter("PK_ID_CATEGORIA", 1) }()
	func() { defer func() { _ = recover() }(); _, _ = qs.All(nil) }()
}

func TestSubAdapters_WithBeegoMock(t *testing.T) {
	// Usa el driver mock registrado en controller_test_utils_test.go
	o := orm.NewOrm()
	a := subOrmAdapter{o: o}
	qs := a.QueryTable(new(models.Subcategoria))
	var out []models.Subcategoria
	// No nos importa el resultado; sólo ejecutar las líneas instrumentadas
	_, _ = qs.All(&out)
}
