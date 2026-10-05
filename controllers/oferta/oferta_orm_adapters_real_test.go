package oferta

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

func TestOfertOrmAdapter_RealOrm_QueryBuilding(t *testing.T) {
	t.Cleanup(resetFakeDB)
	var queries []string
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		queries = append(queries, q)
		return &fakeRows{columns: []string{"cnt"}, values: [][]driver.Value{{int64(5)}}}, nil
	}

	a := ofertOrmAdapter{o: orm.NewOrm()}
	qs := a.QueryTable("oferta")
	require.NotNil(t, qs)
	qs = qs.Filter("activo", true).OrderBy("-pk_id_oferta").Limit(10).Offset(20)

	n, err := qs.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
	require.Len(t, queries, 1)
	assert.Contains(t, queries[0], "COUNT(")
	assert.Contains(t, queries[0], "activo")
}

func TestOfertOrmAdapter_RealOrm_OneAndAll(t *testing.T) {
	t.Cleanup(resetFakeDB)
	a := ofertOrmAdapter{o: orm.NewOrm()}

	var one models.Oferta
	err := a.QueryTable("oferta").Filter("pk_id_oferta", 1).One(&one)
	assert.ErrorIs(t, err, orm.ErrNoRows)

	var list []*models.Oferta
	n, err := a.QueryTable("oferta").All(&list)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Empty(t, list)

	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, assert.AnError }
	assert.Error(t, a.QueryTable("oferta").One(&one))
	_, err = a.QueryTable("oferta").Count()
	assert.Error(t, err)
}

func TestOfertOrmAdapter_RealOrm_CRUD(t *testing.T) {
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
	a := ofertOrmAdapter{o: orm.NewOrm()}

	e := &models.Oferta{PkIdOferta: 3, Titulo: "x", PkIdRestaurante: &models.Restaurante{}}

	id, err := a.Insert(e)
	require.NoError(t, err)
	assert.Equal(t, int64(7), id)

	n, err := a.Update(e, "Titulo")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	n, err = a.Delete(&models.Incidencia{PK_ID_INCIDENCIA: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	joined := strings.Join(execs, "\n")
	assert.Contains(t, joined, "INSERT")
	assert.Contains(t, joined, "UPDATE")
	assert.Contains(t, joined, "DELETE")

	fakeQuery = nil
	err = a.Read(&models.Oferta{PkIdOferta: 99})
	assert.ErrorIs(t, err, orm.ErrNoRows)
}
