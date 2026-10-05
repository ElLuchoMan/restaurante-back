package cupon

import (
	"database/sql/driver"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOrmProvider_ReturnsOrmer(t *testing.T) {
	assert.NotNil(t, defaultOrmProvider())
}

func TestCupOrmAdapter_RealOrm_QueryBuilding(t *testing.T) {
	t.Cleanup(resetFakeDB)
	var queries []string
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		queries = append(queries, q)
		return &fakeRows{columns: []string{"cnt"}, values: [][]driver.Value{{int64(5)}}}, nil
	}

	a := cupOrmAdapter{o: orm.NewOrm()}
	qs := a.QueryTable("cupon")
	require.NotNil(t, qs)
	qs = qs.Filter("activo", true).OrderBy("-pk_id_cupon").Limit(10).Offset(20)

	n, err := qs.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
	require.Len(t, queries, 1)
	assert.Contains(t, queries[0], "COUNT(")
	assert.Contains(t, queries[0], "activo")
}

func TestCupOrmAdapter_RealOrm_OneAndAll(t *testing.T) {
	t.Cleanup(resetFakeDB)
	a := cupOrmAdapter{o: orm.NewOrm()}

	var one models.Cupon
	err := a.QueryTable("cupon").Filter("pk_id_cupon", 1).One(&one)
	assert.ErrorIs(t, err, orm.ErrNoRows)

	var list []*models.Cupon
	n, err := a.QueryTable("cupon").All(&list)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Empty(t, list)

	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, assert.AnError }
	assert.Error(t, a.QueryTable("cupon").One(&one))
	_, err = a.QueryTable("cupon").Count()
	assert.Error(t, err)
}

func TestCupOrmAdapter_RealOrm_CRUD(t *testing.T) {
	t.Cleanup(resetFakeDB)
	var execs []string
	fakeExec = func(q string, args []driver.NamedValue) (driver.Result, error) {
		execs = append(execs, q)
		return fakeResult{}, nil
	}
	// En Postgres el INSERT usa RETURNING y se ejecuta como consulta.
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		execs = append(execs, q)
		return &fakeRows{columns: []string{"id"}, values: [][]driver.Value{{int64(7)}}}, nil
	}
	a := cupOrmAdapter{o: orm.NewOrm()}

	e := &models.Cupon{PkIdCupon: 3, Codigo: "x"}

	id, err := a.Insert(e)
	require.NoError(t, err)
	assert.Equal(t, int64(7), id)

	n, err := a.Update(e, "Codigo")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	n, err = a.Delete(e)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	joined := strings.Join(execs, "\n")
	assert.Contains(t, joined, "INSERT")
	assert.Contains(t, joined, "UPDATE")
	assert.Contains(t, joined, "DELETE")

	fakeQuery = nil
	err = a.Read(&models.Cupon{PkIdCupon: 99})
	assert.ErrorIs(t, err, orm.ErrNoRows)
}
