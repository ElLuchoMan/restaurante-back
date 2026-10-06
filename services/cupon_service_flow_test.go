package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/internal/dberr"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
)

// redimirEnv programa las tablas que consulta RedimirCupon.
type redimirEnv struct {
	pedido      func(dest interface{}) error
	cupon       func(dest interface{}) error
	redenciones func(call int) (int64, error)
	detalle     func(dest interface{}) error
	insert      func(interface{}) (int64, error)
	redCalls    int
	events      []string
}

func cuponVigente() *models.Cupon {
	now := time.Now()
	return &models.Cupon{
		PkIdCupon:      10,
		Activo:         true,
		FechaInicio:    now.Add(-48 * time.Hour),
		FechaFin:       now.Add(48 * time.Hour),
		Scope:          models.CuponScopeGlobal,
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 10,
	}
}

func pedidoDe(cliente int64, estado string) *models.Pedido {
	return &models.Pedido{PK_ID_PEDIDO: 9, ESTADO_PEDIDO: estado, PK_DOCUMENTO_CLIENTE: &models.Cliente{PK_DOCUMENTO_CLIENTE: cliente}}
}

func newRedimirEnv() *redimirEnv {
	return &redimirEnv{
		pedido: func(dest interface{}) error {
			*dest.(*models.Pedido) = *pedidoDe(7, models.EstadoPedidoIniciado)
			return nil
		},
		cupon:       func(dest interface{}) error { *dest.(*models.Cupon) = *cuponVigente(); return nil },
		redenciones: func(int) (int64, error) { return 0, nil },
		detalle: func(dest interface{}) error {
			*dest.(*[]*models.DetallePedido) = []*models.DetallePedido{
				{PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 2}, Cantidad: 2, Precio: 5000},
				{PKIDProducto: nil, Cantidad: 1, Precio: 1},
			}
			return nil
		},
		insert: func(interface{}) (int64, error) { return 1, nil },
	}
}

// qs arma un query seter que registra FOR UPDATE y los conteos en e.events.
func (e *redimirEnv) qs(tabla string, base *mockQuerySeter) *mockQuerySeter {
	base.forUpdateHook = func() cuponQuerySeter { e.events = append(e.events, tabla+":lock"); return base }
	return base
}

func (e *redimirEnv) service() *CuponService {
	return NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"pedido": func() cuponQuerySeter {
				return e.qs("pedido", &mockQuerySeter{oneHook: func(d interface{}, _ ...string) error { return e.pedido(d) }})
			},
			"cupon": func() cuponQuerySeter {
				return e.qs("cupon", &mockQuerySeter{oneHook: func(d interface{}, _ ...string) error { return e.cupon(d) }})
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{countHook: func() (int64, error) { e.redCalls++; return e.redenciones(e.redCalls) }}
			},
			"detalle_pedido": func() cuponQuerySeter {
				return &mockQuerySeter{allHook: func(d interface{}, _ ...string) (int64, error) { return 0, e.detalle(d) }}
			},
		},
		insertFn: func(m interface{}) (int64, error) { e.events = append(e.events, "insert"); return e.insert(m) },
	})
}

func redimir(e *redimirEnv, pedido *int64) (*models.CuponRedencion, error) {
	return e.service().RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 7, PedidoId: pedido})
}

func TestRedimirCupon_Flujo(t *testing.T) {
	pedido := int64(9)

	_, err := (&CuponService{}).RedimirCupon(context.Background(), "C", &models.RedimirCuponRequest{PedidoId: &pedido})
	assert.Error(t, err)
	_, err = redimir(newRedimirEnv(), nil)
	assert.ErrorIs(t, err, ErrPedidoRequerido)

	t.Run("exito y orden de bloqueo", func(t *testing.T) {
		var inserted *models.CuponRedencion
		e := newRedimirEnv()
		e.insert = func(m interface{}) (int64, error) { inserted = m.(*models.CuponRedencion); return 1, nil }
		r, err := redimir(e, &pedido)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), r.MontoDescuento)
		assert.Equal(t, int64(9), r.PkIdPedido.PK_ID_PEDIDO)
		assert.Equal(t, int64(7), inserted.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
		// Primero se bloquea el pedido, luego el cupón, y recién después se inserta.
		assert.Equal(t, []string{"pedido:lock", "cupon:lock", "insert"}, e.events)
	})

	t.Run("pedido", func(t *testing.T) {
		e := newRedimirEnv()
		e.pedido = func(interface{}) error { return orm.ErrNoRows }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrPedidoNoEncontrado)
		e.pedido = func(interface{}) error { return errors.New("boom") }
		_, err = redimir(e, &pedido)
		assert.ErrorContains(t, err, "pedido")
		assert.NotErrorIs(t, err, ErrPedidoNoEncontrado)
	})

	t.Run("pedido ajeno o cerrado", func(t *testing.T) {
		e := newRedimirEnv()
		e.pedido = func(d interface{}) error {
			*d.(*models.Pedido) = *pedidoDe(99, models.EstadoPedidoIniciado)
			return nil
		}
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrPedidoAjeno)
		e.pedido = func(d interface{}) error { *d.(*models.Pedido) = models.Pedido{PK_ID_PEDIDO: 9}; return nil }
		_, err = redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrPedidoAjeno)
		for _, estado := range []string{models.EstadoPedidoTerminado, models.EstadoPedidoCancelado} {
			e.pedido = func(d interface{}) error { *d.(*models.Pedido) = *pedidoDe(7, estado); return nil }
			_, err = redimir(e, &pedido)
			assert.ErrorIs(t, err, ErrPedidoCerrado)
		}
	})

	t.Run("cupon", func(t *testing.T) {
		e := newRedimirEnv()
		e.cupon = func(interface{}) error { return orm.ErrNoRows }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponNoEncontrado)
		e.cupon = func(interface{}) error { return errors.New("boom") }
		_, err = redimir(e, &pedido)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrCuponNoEncontrado)
	})

	t.Run("ya redimido en el pedido", func(t *testing.T) {
		e := newRedimirEnv()
		e.redenciones = func(int) (int64, error) { return 1, nil }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponConflicto)
		e.redCalls = 0
		e.redenciones = func(int) (int64, error) { return 0, errors.New("boom") }
		_, err = redimir(e, &pedido)
		assert.ErrorContains(t, err, "redenciones")
	})

	t.Run("detalle", func(t *testing.T) {
		e := newRedimirEnv()
		e.detalle = func(interface{}) error { return errors.New("boom") }
		_, err := redimir(e, &pedido)
		assert.ErrorContains(t, err, "detalle")
	})

	t.Run("validacion bajo el bloqueo", func(t *testing.T) {
		e := newRedimirEnv()
		// El conteo de usos totales (2.ª consulta a cupon_redencion) falla.
		e.cupon = func(d interface{}) error {
			c := cuponVigente()
			mx := 5
			c.MaxUsos = &mx
			*d.(*models.Cupon) = *c
			return nil
		}
		e.redenciones = func(call int) (int64, error) {
			if call == 2 {
				return 0, errors.New("boom")
			}
			return 0, nil
		}
		_, err := redimir(e, &pedido)
		assert.ErrorContains(t, err, "validar")

		// Usos agotados: conflicto.
		e.redCalls = 0
		e.redenciones = func(call int) (int64, error) {
			if call == 2 {
				return 5, nil
			}
			return 0, nil
		}
		_, err = redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponConflicto)

		// Cupón inactivo: no aplicable.
		e = newRedimirEnv()
		e.cupon = func(d interface{}) error { c := cuponVigente(); c.Activo = false; *d.(*models.Cupon) = *c; return nil }
		_, err = redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponNoAplicable)
	})

	t.Run("insert", func(t *testing.T) {
		e := newRedimirEnv()
		e.insert = func(interface{}) (int64, error) { return 0, errors.New("boom") }
		_, err := redimir(e, &pedido)
		assert.ErrorContains(t, err, "registrar")
		// Índice único (cupón, pedido): carrera perdida => conflicto.
		e.insert = func(interface{}) (int64, error) {
			return 0, errors.New("duplicate key value violates unique constraint")
		}
		_, err = redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponConflicto)
		assert.True(t, dberr.IsUnique(errors.New("duplicate key")))
	})
}

func TestValidarCupon_ConPedido(t *testing.T) {
	pedido := int64(9)
	req := func() *models.ValidarCuponRequest {
		return &models.ValidarCuponRequest{Codigo: "CODE", ClienteId: 7, PedidoId: &pedido,
			Items: []models.ValidarCuponItemRequest{{ProductoId: 1, Cantidad: 1, Precio: 999999}}}
	}
	t.Run("usa el detalle real del pedido y no los ítems del cliente", func(t *testing.T) {
		e := newRedimirEnv()
		resp, err := e.service().ValidarCupon(context.Background(), req())
		assert.NoError(t, err)
		assert.True(t, resp.Aplicable)
		// 10 % de 2*5000 + 1*0 (el segundo detalle no tiene producto).
		assert.Equal(t, int64(1000), resp.MontoDescuento)
		assert.NotContains(t, e.events, "pedido:lock", "validar no bloquea")
	})
	t.Run("errores", func(t *testing.T) {
		e := newRedimirEnv()
		e.pedido = func(interface{}) error { return orm.ErrNoRows }
		_, err := e.service().ValidarCupon(context.Background(), req())
		assert.ErrorIs(t, err, ErrPedidoNoEncontrado)

		e = newRedimirEnv()
		e.pedido = func(d interface{}) error {
			*d.(*models.Pedido) = *pedidoDe(99, models.EstadoPedidoIniciado)
			return nil
		}
		_, err = e.service().ValidarCupon(context.Background(), req())
		assert.ErrorIs(t, err, ErrPedidoAjeno)

		e = newRedimirEnv()
		e.detalle = func(interface{}) error { return errors.New("boom") }
		_, err = e.service().ValidarCupon(context.Background(), req())
		assert.ErrorContains(t, err, "detalle")
	})
}

func TestInTxYAdaptadores(t *testing.T) {
	// Sin runTx se ejecuta directamente sobre el mismo servicio.
	s := NewCuponService(&mockCuponOrmer{})
	called := false
	assert.NoError(t, s.inTx(context.Background(), func(ts *CuponService) error { called = ts == s; return nil }))
	assert.True(t, called)

	// Con runTx se delega (y se propaga el error).
	boom := errors.New("boom")
	s.runTx = func(_ context.Context, fn func(*CuponService) error) error { return boom }
	assert.ErrorIs(t, s.inTx(context.Background(), func(*CuponService) error { return nil }), boom)

	// Adaptador sobre un orm.QuerySeter real (stub).
	dummy := &dummyQuerySeter{}
	var table string
	var inserted bool
	a := ormerAdapter{
		queryTable: func(name string) orm.QuerySeter { table = name; return dummy },
		insert:     func(interface{}) (int64, error) { inserted = true; return 3, nil },
	}
	qs := a.QueryTable("cupon").Filter("codigo", "X").RelatedSel("a").ForUpdate()
	assert.Equal(t, "cupon", table)
	assert.Equal(t, 1, dummy.filterCalls)
	assert.Equal(t, 1, dummy.relatedCalls)
	assert.NoError(t, qs.One(&models.Cupon{}))
	assert.True(t, dummy.oneCalled)
	n, err := qs.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
	_, err = qs.All(&[]*models.Cupon{})
	assert.NoError(t, err)
	id, _ := a.Insert(&models.Cupon{})
	assert.Equal(t, int64(3), id)
	assert.True(t, inserted)
}

// txStub es un orm.TxOrmer que solo implementa lo que usa el servicio.
type txStub struct {
	orm.TxOrmer
	table     string
	inserted  bool
	commits   int
	rollbacks int
	commitErr error
}

func (t *txStub) QueryTable(name interface{}) orm.QuerySeter {
	t.table = name.(string)
	return &dummyQuerySeter{}
}
func (t *txStub) Insert(interface{}) (int64, error) { t.inserted = true; return 1, nil }
func (t *txStub) Commit() error                     { t.commits++; return t.commitErr }
func (t *txStub) Rollback() error                   { t.rollbacks++; return nil }

func TestNewCuponServiceFromTx(t *testing.T) {
	tx := &txStub{}
	s := newCuponServiceFromTx(tx)
	_ = s.ormer.QueryTable("cupon")
	_, _ = s.ormer.Insert(&models.Cupon{})
	assert.Equal(t, "cupon", tx.table)
	assert.True(t, tx.inserted)
	assert.Nil(t, s.runTx)
}

func TestNewCuponServiceFromOrm(t *testing.T) {
	var table string
	var inserted, txCalled bool
	m := &mockDescuentoOrmer{
		queryTableFn: func(name string) orm.QuerySeter { table = name; return &mockDescuentoQuerySeter{} },
		insertFn:     func(interface{}) (int64, error) { inserted = true; return 1, nil },
		beginFn:      func() (orm.TxOrmer, error) { txCalled = true; return &txStub{}, nil },
	}
	s := NewCuponServiceFromOrm(m)
	_ = s.ormer.QueryTable("cupon")
	_, _ = s.ormer.Insert(&models.Cupon{})
	assert.Equal(t, "cupon", table)
	assert.True(t, inserted)

	var bound *CuponService
	assert.NoError(t, s.inTx(context.Background(), func(ts *CuponService) error { bound = ts; return nil }))
	assert.True(t, txCalled)
	assert.NotSame(t, s, bound, "dentro de la transacción el servicio se liga al TxOrmer")
}

func TestAhoraBogota_SinZona(t *testing.T) {
	prev := database.BogotaZone
	database.BogotaZone = nil
	defer func() { database.BogotaZone = prev }()
	_, off := ahoraBogota().Zone()
	assert.Equal(t, -5*60*60, off)
}

func TestValidarReglasNegocioCupon_NegativosYCategoria(t *testing.T) {
	s := &CuponService{}
	neg := -1
	base := func() *models.Cupon {
		return &models.Cupon{Scope: models.CuponScopeGlobal, TipoDescuento: models.TipoDescuentoMonto, ValorDescuento: 1}
	}
	c := base()
	c.MaxUsos = &neg
	assert.ErrorContains(t, s.ValidarReglasNegocioCupon(c), "maxUsos")
	c = base()
	c.LimitePorCliente = &neg
	assert.ErrorContains(t, s.ValidarReglasNegocioCupon(c), "limitePorCliente")
	m := int64(-1)
	c = base()
	c.MontoMinimo = &m
	assert.ErrorContains(t, s.ValidarReglasNegocioCupon(c), "montoMinimo")
}

func TestEsProductoAplicable_SubcategoriaSinCategoria(t *testing.T) {
	s := &CuponService{ormer: &mockCuponOrmer{tables: map[string]func() cuponQuerySeter{
		"producto": func() cuponQuerySeter {
			return &mockQuerySeter{oneHook: func(d interface{}, _ ...string) error {
				d.(*models.Producto).PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: 7}
				return nil
			}}
		},
		"subcategoria": func() cuponQuerySeter {
			return &mockQuerySeter{oneHook: func(d interface{}, _ ...string) error { d.(*models.Subcategoria).PK_ID_CATEGORIA = nil; return nil }}
		},
	}}}
	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: &models.Categoria{PK_ID_CATEGORIA: 5}}
	assert.False(t, s.esProductoAplicable(cupon, 10))
}

func TestDoTx(t *testing.T) {
	ctx := context.Background()
	newMock := func(tx *txStub, beginErr error) *mockDescuentoOrmer {
		return &mockDescuentoOrmer{beginFn: func() (orm.TxOrmer, error) { return tx, beginErr }}
	}

	tx := &txStub{}
	assert.NoError(t, doTx(ctx, newMock(tx, nil), func(orm.TxOrmer) error { return nil }))
	assert.Equal(t, 1, tx.commits)
	assert.Equal(t, 0, tx.rollbacks)

	boom := errors.New("boom")
	tx = &txStub{}
	assert.ErrorIs(t, doTx(ctx, newMock(tx, nil), func(orm.TxOrmer) error { return boom }), boom)
	assert.Equal(t, 0, tx.commits)
	assert.Equal(t, 1, tx.rollbacks)

	// El error del COMMIT se devuelve (orm.DoTx lo ignoraría).
	tx = &txStub{commitErr: errors.New("commit boom")}
	assert.ErrorContains(t, doTx(ctx, newMock(tx, nil), func(orm.TxOrmer) error { return nil }), "commit boom")

	assert.ErrorIs(t, doTx(ctx, newMock(nil, boom), func(orm.TxOrmer) error { return nil }), boom)

	// Un pánico revierte y se vuelve a lanzar.
	tx = &txStub{}
	assert.PanicsWithValue(t, "pánico", func() {
		_ = doTx(ctx, newMock(tx, nil), func(orm.TxOrmer) error { panic("pánico") })
	})
	assert.Equal(t, 1, tx.rollbacks)
	assert.Equal(t, 0, tx.commits)
}
