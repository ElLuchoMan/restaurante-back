package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
)

// redimirEnv programa las tablas que consulta RedimirCupon.
type redimirEnv struct {
	cupon       func(dest interface{}) error
	cliente     func() (int64, error)
	pedido      func() (int64, error)
	redenciones func(call int) (int64, error)
	detalle     func(dest interface{}) error
	insert      func(interface{}) (int64, error)
	redCalls    int
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

func newRedimirEnv() *redimirEnv {
	return &redimirEnv{
		cupon:       func(dest interface{}) error { *dest.(*models.Cupon) = *cuponVigente(); return nil },
		cliente:     func() (int64, error) { return 1, nil },
		pedido:      func() (int64, error) { return 1, nil },
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

func (e *redimirEnv) service() *CuponService {
	return NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{oneHook: func(d interface{}, _ ...string) error { return e.cupon(d) }}
			},
			"cliente": func() cuponQuerySeter { return &mockQuerySeter{countHook: e.cliente} },
			"pedido":  func() cuponQuerySeter { return &mockQuerySeter{countHook: e.pedido} },
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{countHook: func() (int64, error) { e.redCalls++; return e.redenciones(e.redCalls) }}
			},
			"detalle_pedido": func() cuponQuerySeter {
				return &mockQuerySeter{allHook: func(d interface{}, _ ...string) (int64, error) { return 0, e.detalle(d) }}
			},
		},
		insertFn: e.insert,
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

	t.Run("exito", func(t *testing.T) {
		var inserted *models.CuponRedencion
		e := newRedimirEnv()
		e.insert = func(m interface{}) (int64, error) { inserted = m.(*models.CuponRedencion); return 1, nil }
		r, err := redimir(e, &pedido)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), r.MontoDescuento)
		assert.Equal(t, int64(9), r.PkIdPedido.PK_ID_PEDIDO)
		assert.Equal(t, int64(7), inserted.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE)
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

	t.Run("cliente", func(t *testing.T) {
		e := newRedimirEnv()
		e.cliente = func() (int64, error) { return 0, nil }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrClienteNoEncontrado)
		e.cliente = func() (int64, error) { return 0, errors.New("boom") }
		_, err = redimir(e, &pedido)
		assert.ErrorContains(t, err, "cliente")
	})

	t.Run("pedido", func(t *testing.T) {
		e := newRedimirEnv()
		e.pedido = func() (int64, error) { return 0, nil }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrPedidoNoEncontrado)
		e.pedido = func() (int64, error) { return 0, errors.New("boom") }
		_, err = redimir(e, &pedido)
		assert.ErrorContains(t, err, "pedido")
	})

	t.Run("ya redimido", func(t *testing.T) {
		e := newRedimirEnv()
		e.redenciones = func(int) (int64, error) { return 1, nil }
		_, err := redimir(e, &pedido)
		assert.ErrorIs(t, err, ErrCuponConflicto)
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

	t.Run("validacion", func(t *testing.T) {
		e := newRedimirEnv()
		// Segundo conteo (usos totales) falla.
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
	})
}

func TestNewCuponServiceFromOrm(t *testing.T) {
	var table string
	var inserted bool
	s := NewCuponServiceFromOrm(&mockDescuentoOrmer{
		queryTableFn: func(name string) orm.QuerySeter { table = name; return &mockDescuentoQuerySeter{} },
		insertFn:     func(interface{}) (int64, error) { inserted = true; return 1, nil },
	})
	_ = s.ormer.QueryTable("cupon")
	_, _ = s.ormer.Insert(&models.Cupon{})
	assert.Equal(t, "cupon", table)
	assert.True(t, inserted)
}

func TestBeegoCuponQuerySeter_All(t *testing.T) {
	assert.NotNil(t, beegoCuponQuerySeter{})
	n, err := beegoCuponQuerySeter{}.All(nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)

	called := false
	q := beegoCuponQuerySeter{allFunc: func(interface{}, ...string) (int64, error) { called = true; return 3, nil }}
	n, err = q.All(nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n)
	assert.True(t, called)

	wrapped := wrapOrmQuerySeter(&dummyQuerySeter{})
	_, err = wrapped.All(&[]*models.Cupon{})
	assert.NoError(t, err)
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
