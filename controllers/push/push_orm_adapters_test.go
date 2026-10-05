//go:build !unit

package push

import (
	"database/sql/driver"
	"errors"
	"testing"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFakeDB(t *testing.T) {
	t.Helper()
	mockQueryHook, mockExecHook = nil, nil
	t.Cleanup(func() { mockQueryHook, mockExecHook = nil, nil })
}

func TestDefaultProvidersReturnRealOrm(t *testing.T) {
	assert.NotNil(t, orm.NewOrm())
	assert.NotNil(t, newServiceOrm())
	assert.NotNil(t, newPushService(newServiceOrm()))

	_, ok := newPushService(newServiceOrm()).(*services.PushService)
	assert.True(t, ok)

	// init() de PushController_test.go reemplaza pushOrmNew; el adaptador
	// real se valida construyéndolo con el proveedor por defecto.
	a := pushOrmAdapter{o: orm.NewOrm()}
	assert.NotNil(t, a.QueryTable("push_dispositivo"))
}

func TestPushOrmAdapterQueryTableChain(t *testing.T) {
	resetFakeDB(t)
	a := pushOrmAdapter{o: orm.NewOrm()}

	qs := a.QueryTable("push_dispositivo").
		Filter("enabled", true).
		OrderBy("-created_at").
		Limit(5).
		Offset(2)
	_, isAdapter := qs.(pushQSAdapter)
	require.True(t, isAdapter)

	total, err := qs.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	var one models.PushDispositivo
	require.NoError(t, qs.One(&one))

	var all []*models.PushDispositivo
	n, err := qs.All(&all)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	assert.Len(t, all, 1)
}

func TestPushOrmAdapterCountAndOneErrors(t *testing.T) {
	resetFakeDB(t)
	boom := errors.New("db down")
	mockQueryHook = func(string) (driver.Rows, error) { return nil, boom }
	a := pushOrmAdapter{o: orm.NewOrm()}
	qs := a.QueryTable("push_dispositivo")

	_, err := qs.Count()
	assert.ErrorIs(t, err, boom)
	var one models.PushDispositivo
	assert.ErrorIs(t, qs.One(&one), boom)
	_, err = qs.All(&[]*models.PushDispositivo{})
	assert.ErrorIs(t, err, boom)
}

func TestPushOrmAdapterCRUD(t *testing.T) {
	resetFakeDB(t)
	a := pushOrmAdapter{o: orm.NewOrm()}

	d := &models.PushDispositivo{PkIdPushDispositivo: 5}
	require.NoError(t, a.Read(d))
	assert.Equal(t, int64(0), d.PkIdPushDispositivo) // fila con columnas NULL

	d = &models.PushDispositivo{PkIdPushDispositivo: 5, Plataforma: models.PlataformaWeb, Enabled: true}
	rows, err := a.Update(d)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	// El borrado en cascada consulta filas relacionadas: ninguna.
	mockQueryHook = func(string) (driver.Rows, error) {
		return &fakeRows{columns: []string{"c"}}, nil
	}
	rows, err = a.Delete(d)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)
	mockQueryHook = nil

	id, err := a.Insert(&models.PushDispositivo{Plataforma: models.PlataformaWeb, Enabled: true})
	require.NoError(t, err)
	assert.Equal(t, int64(3), id)
}

// Se captura antes de que init() de PushController_test.go sustituya el hook.
var originalPushOrmNew = pushOrmNew

func TestPushOrmNewDefaultBuildsAdapter(t *testing.T) {
	resetFakeDB(t)
	o, ok := originalPushOrmNew().(pushOrmAdapter)
	require.True(t, ok)
	require.NotNil(t, o.o)

	total, err := o.QueryTable("push_dispositivo").Count()
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
}
