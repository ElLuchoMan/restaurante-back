package services

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vals = map[string]driver.Value

func rowsFn(rows ...vals) func(*fkConn, string, []driver.NamedValue) ([]map[string]driver.Value, error) {
	return func(*fkConn, string, []driver.NamedValue) ([]map[string]driver.Value, error) { return rows, nil }
}

func errRows(err error) func(*fkConn, string, []driver.NamedValue) ([]map[string]driver.Value, error) {
	return func(*fkConn, string, []driver.NamedValue) ([]map[string]driver.Value, error) { return nil, err }
}

func countFn(n int64) func(*fkConn, string, []driver.NamedValue) (int64, error) {
	return func(*fkConn, string, []driver.NamedValue) (int64, error) { return n, nil }
}

func errCount(err error) func(*fkConn, string, []driver.NamedValue) (int64, error) {
	return func(*fkConn, string, []driver.NamedValue) (int64, error) { return 0, err }
}

func fkPedido(cliente int64, estado string, pago, restaurante driver.Value) vals {
	return vals{"pk_id_pedido": int64(9), "estado_pedido": estado, "pk_documento_cliente": cliente, "pk_id_pago": pago, "pk_id_restaurante": restaurante}
}

func fkCuponRow() vals {
	now := time.Now()
	return vals{
		"pk_id_cupon": int64(10), "codigo": "VERANO", "scope": "GLOBAL", "tipo_descuento": "PORCENTAJE",
		"valor_descuento": int64(10), "fecha_inicio": now.Add(-48 * time.Hour), "fecha_fin": now.Add(48 * time.Hour), "activo": true,
	}
}

func fkOfertaRow() vals {
	now := time.Now()
	return vals{
		"pk_id_oferta": int64(5), "titulo": "Martes", "tipo_descuento": "MONTO", "valor_descuento": int64(3000),
		"fecha_inicio": now.Add(-48 * time.Hour), "fecha_fin": now.Add(48 * time.Hour), "dias_semana": "", "activo": true,
		"pk_id_restaurante": int64(1),
	}
}

// fkBase programa un pedido abierto del cliente 7 con pago 4 (PENDIENTE, 10000)
// y un detalle de 2 x 5000 del producto 2 (subtotal 10000).
func fkBase() {
	fk.reset()
	fk.rows["pedido"] = rowsFn(fkPedido(7, models.EstadoPedidoIniciado, int64(4), int64(1)))
	fk.rows["pago"] = rowsFn(vals{"pk_id_pago": int64(4), "monto": int64(10000), "estado_pago": models.EstadoPagoPendiente})
	fk.rows["detalle_pedido"] = rowsFn(vals{"pk_id_detalle": int64(1), "pk_id_pedido": int64(9), "pk_id_producto": int64(2), "precio": int64(5000), "cantidad": int64(2)})
	fk.rows["cupon"] = rowsFn(fkCuponRow())
	fk.rows["oferta"] = rowsFn(fkOfertaRow())
	fk.rows["oferta_producto"] = rowsFn(vals{"pk_id_oferta": int64(5), "pk_id_producto": int64(2)})
}

func aplicar(req *models.AplicarDescuentoRequest) (*models.DescuentoAplicadoResponse, error) {
	return NewDescuentoService(orm.NewOrm()).AplicarDescuento(context.Background(), 9, 7, req)
}

func cuponReq() *models.AplicarDescuentoRequest {
	id := int64(10)
	return &models.AplicarDescuentoRequest{PkIdCupon: &id}
}

func ofertaReq() *models.AplicarDescuentoRequest {
	id := int64(5)
	return &models.AplicarDescuentoRequest{PkIdOferta: &id}
}

func indexOf(stmts []string, fragments ...string) int {
	for i, s := range stmts {
		ok := true
		for _, f := range fragments {
			ok = ok && strings.Contains(s, f)
		}
		if ok {
			return i
		}
	}
	return -1
}

func TestAplicarDescuento_Cupon_Exito(t *testing.T) {
	fkBase()
	var upd []driver.NamedValue
	fk.onInsert = func(c *fkConn, table string, args []driver.NamedValue) {
		if table == "pago" {
			upd = args
		}
	}
	extra := `{"nota":"mesa 4","tipo":"ignorado"}`
	req := cuponReq()
	req.Detalle = []byte(extra)
	resp, err := aplicar(req)
	require.NoError(t, err)

	assert.Equal(t, int64(10000), resp.Subtotal)
	assert.Equal(t, int64(1000), resp.MontoDescuento, "10 %% de 10000 calculado en el servidor")
	assert.Equal(t, int64(9000), resp.Total)
	require.NotNil(t, resp.PagoId)
	assert.Equal(t, int64(4), *resp.PagoId)
	assert.Equal(t, int64(1000), resp.Descuento.MontoDescuento)
	assert.JSONEq(t, `{"nota":"mesa 4","tipo":"cupon","codigo":"VERANO","scope":"GLOBAL"}`, string(resp.Descuento.DetalleObj))
	_ = upd

	st := fk.statements()
	pedido := indexOf(st, `FROM "pedido"`, "FOR UPDATE")
	pago := indexOf(st, `FROM "pago"`, "FOR UPDATE")
	cupon := indexOf(st, `FROM "cupon"`, "FOR UPDATE")
	redencion := indexOf(st, `INSERT INTO "cupon_redencion"`)
	descuento := indexOf(st, `INSERT INTO "pedido_descuento_aplicado"`)
	update := indexOf(st, `UPDATE "pago"`)
	for name, i := range map[string]int{"pedido": pedido, "pago": pago, "cupon": cupon, "redencion": redencion, "descuento": descuento, "update": update} {
		require.GreaterOrEqual(t, i, 0, name)
	}
	assert.Less(t, pedido, pago)
	assert.Less(t, pago, cupon)
	assert.Less(t, cupon, redencion)
	assert.Less(t, redencion, descuento)
	assert.Less(t, descuento, update)
}

func TestAplicarDescuento_Cupon_SinPagoYClampDelPago(t *testing.T) {
	// Pedido sin pago: el total es subtotal - descuento y no se actualiza ningún pago.
	fkBase()
	fk.rows["pedido"] = rowsFn(fkPedido(7, models.EstadoPedidoIniciado, nil, nil))
	resp, err := aplicar(cuponReq())
	require.NoError(t, err)
	assert.Equal(t, int64(9000), resp.Total)
	assert.Nil(t, resp.PagoId)
	assert.Equal(t, -1, indexOf(fk.statements(), `UPDATE "pago"`))

	// Pago con monto menor que el descuento: el monto nunca queda negativo.
	fkBase()
	fk.rows["pago"] = rowsFn(vals{"pk_id_pago": int64(4), "monto": int64(500), "estado_pago": models.EstadoPagoPendiente})
	resp, err = aplicar(cuponReq())
	require.NoError(t, err)
	assert.Equal(t, int64(0), resp.Total)
}

func TestAplicarDescuento_Oferta(t *testing.T) {
	t.Run("exito", func(t *testing.T) {
		fkBase()
		resp, err := aplicar(ofertaReq())
		require.NoError(t, err)
		assert.Equal(t, int64(3000), resp.MontoDescuento)
		assert.Equal(t, int64(7000), resp.Total)
		assert.JSONEq(t, `{"tipo":"oferta","titulo":"Martes"}`, string(resp.Descuento.DetalleObj))
		assert.Equal(t, int64(5), resp.Descuento.PkIdOferta.PkIdOferta)
		assert.Equal(t, -1, indexOf(fk.statements(), `INSERT INTO "cupon_redencion"`), "una oferta no redime cupones")
	})
	t.Run("no encontrada y error", func(t *testing.T) {
		fkBase()
		fk.rows["oferta"] = rowsFn()
		_, err := aplicar(ofertaReq())
		assert.ErrorIs(t, err, ErrOfertaNoEncontrada)
		fk.rows["oferta"] = errRows(errors.New("boom"))
		_, err = aplicar(ofertaReq())
		assert.ErrorContains(t, err, "buscar oferta")
	})
	t.Run("no vigente", func(t *testing.T) {
		fkBase()
		o := fkOfertaRow()
		o["activo"] = false
		fk.rows["oferta"] = rowsFn(o)
		_, err := aplicar(ofertaReq())
		assert.ErrorIs(t, err, ErrOfertaNoAplicable)
		assert.ErrorContains(t, err, "inactiva")
	})
	t.Run("otro restaurante", func(t *testing.T) {
		fkBase()
		fk.rows["pedido"] = rowsFn(fkPedido(7, models.EstadoPedidoIniciado, int64(4), int64(2)))
		_, err := aplicar(ofertaReq())
		assert.ErrorIs(t, err, ErrOfertaNoAplicable)
		assert.ErrorContains(t, err, "restaurante")
	})
	t.Run("sin productos en el pedido", func(t *testing.T) {
		fkBase()
		fk.rows["oferta_producto"] = rowsFn(vals{"pk_id_oferta": int64(5), "pk_id_producto": int64(99)})
		_, err := aplicar(ofertaReq())
		assert.ErrorIs(t, err, ErrOfertaNoAplicable)
		assert.ErrorContains(t, err, "ningún producto")
	})
	t.Run("error al leer productos de la oferta", func(t *testing.T) {
		fkBase()
		fk.rows["oferta_producto"] = errRows(errors.New("boom"))
		_, err := aplicar(ofertaReq())
		assert.ErrorContains(t, err, "productos de la oferta")
	})
}

func TestAplicarDescuento_Errores(t *testing.T) {
	cases := []struct {
		name  string
		setup func()
		req   func() *models.AplicarDescuentoRequest
		is    error
		text  string
	}{
		{"pedido inexistente", func() { fk.rows["pedido"] = rowsFn() }, cuponReq, ErrPedidoNoEncontrado, ""},
		{"pedido de otro cliente", func() { fk.rows["pedido"] = rowsFn(fkPedido(8, models.EstadoPedidoIniciado, int64(4), int64(1))) }, cuponReq, ErrPedidoAjeno, ""},
		{"pedido cancelado", func() { fk.rows["pedido"] = rowsFn(fkPedido(7, models.EstadoPedidoCancelado, int64(4), int64(1))) }, cuponReq, ErrPedidoCerrado, ""},
		{"error al leer el pago", func() { fk.rows["pago"] = errRows(errors.New("boom")) }, cuponReq, nil, "pago del pedido"},
		{"pedido ya pagado", func() {
			fk.rows["pago"] = rowsFn(vals{"pk_id_pago": int64(4), "monto": int64(10000), "estado_pago": models.EstadoPagoPagado})
		}, cuponReq, ErrPedidoPagado, ""},
		{"error al contar descuentos", func() { fk.counts["pedido_descuento_aplicado"] = errCount(errors.New("boom")) }, cuponReq, nil, "descuentos existentes"},
		{"ya tiene descuento", func() { fk.counts["pedido_descuento_aplicado"] = countFn(1) }, cuponReq, ErrDescuentoYaAplicado, ""},
		{"error en el detalle", func() { fk.rows["detalle_pedido"] = errRows(errors.New("boom")) }, cuponReq, nil, "detalle del pedido"},
		{"cupon inexistente", func() { fk.rows["cupon"] = rowsFn() }, cuponReq, ErrCuponNoEncontrado, ""},
		{"cupon ya redimido en el pedido", func() { fk.counts["cupon_redencion"] = countFn(1) }, cuponReq, ErrCuponConflicto, ""},
		{"cupon inactivo", func() {
			c := fkCuponRow()
			c["activo"] = false
			fk.rows["cupon"] = rowsFn(c)
		}, cuponReq, ErrCuponNoAplicable, ""},
		{"insert duplicado", func() { fk.insertErr["pedido_descuento_aplicado"] = errors.New("duplicate key value") }, cuponReq, ErrDescuentoYaAplicado, ""},
		{"insert falla", func() { fk.insertErr["pedido_descuento_aplicado"] = errors.New("boom") }, cuponReq, nil, "registrar descuento"},
		{"update del pago falla", func() { fk.updateErr["pago"] = errors.New("boom") }, cuponReq, nil, "total del pago"},
		{"redencion duplicada (indice unico)", func() { fk.insertErr["cupon_redencion"] = errors.New("duplicate key value") }, cuponReq, ErrCuponConflicto, ""},
		{"commit falla", func() { fk.commitErr = errors.New("commit boom") }, cuponReq, nil, "commit boom"},
		{"begin falla", func() { fk.beginErr = errors.New("begin boom") }, cuponReq, nil, "begin boom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fkBase()
			tc.setup()
			resp, err := aplicar(tc.req())
			require.Error(t, err)
			assert.Nil(t, resp)
			if tc.is != nil {
				assert.ErrorIs(t, err, tc.is)
			}
			if tc.text != "" {
				assert.ErrorContains(t, err, tc.text)
			}
		})
	}
}

func TestAplicarDescuento_SolicitudInvalida(t *testing.T) {
	fkBase()
	id := int64(1)
	for name, req := range map[string]*models.AplicarDescuentoRequest{
		"sin fuente":        {},
		"ambas fuentes":     {PkIdCupon: &id, PkIdOferta: &id},
		"detalle no objeto": {PkIdCupon: &id, Detalle: []byte(`[1]`)},
		"detalle null":      {PkIdCupon: &id, Detalle: []byte(`null`)},
	} {
		_, err := aplicar(req)
		assert.ErrorIs(t, err, ErrDescuentoInvalido, name)
	}
	assert.Empty(t, fk.statements(), "una solicitud inválida no toca la base")
}

func TestAplicarDescuento_NadaQuedaAMedias(t *testing.T) {
	// Si falla el UPDATE del pago, los INSERT (redención y descuento) quedan en la
	// misma transacción y se revierten: solo se confirma cuando todo salió bien.
	fkBase()
	var committed, rolledBack int
	fk.onInsert = func(c *fkConn, table string, _ []driver.NamedValue) {
		c.onCommit(func() { committed++ })
	}
	fk.updateErr["pago"] = errors.New("boom")
	_, err := aplicar(cuponReq())
	require.Error(t, err)
	assert.Equal(t, 0, committed, "ningún INSERT se confirma si falla el pago")
	_ = rolledBack

	fkBase()
	committed = 0
	fk.onInsert = func(c *fkConn, table string, _ []driver.NamedValue) {
		c.onCommit(func() { committed++ })
	}
	_, err = aplicar(cuponReq())
	require.NoError(t, err)
	assert.Equal(t, 2, committed, "redención y descuento se confirman juntos")
}
