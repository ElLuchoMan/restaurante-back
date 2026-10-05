package services

import (
	"context"
	"encoding/json"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ormerSinDescuentos(insertFn func(interface{}) (int64, error), readFn func(interface{}, ...string) error) *mockDescuentoOrmer {
	return &mockDescuentoOrmer{
		readFn: readFn,
		queryTableFn: func(string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{}
		},
		insertFn: insertFn,
	}
}

func TestDescuentoService_AplicarDescuento_MontoNegativo(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{})
	cuponID := int64(1)
	_, err := service.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdCupon: &cuponID, MontoDescuento: -1})
	assert.ErrorIs(t, err, ErrDescuentoInvalido)
}

func TestDescuentoService_AplicarDescuento_DetalleNoObjeto(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{})
	cuponID := int64(1)
	for _, raw := range []string{`[1,2]`, `null`, `"texto"`} {
		_, err := service.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdCupon: &cuponID, Detalle: json.RawMessage(raw)})
		assert.ErrorIs(t, err, ErrDescuentoInvalido, raw)
	}
}

func TestDescuentoService_AplicarDescuento_ErroresTipados(t *testing.T) {
	cuponID, ofertaID := int64(1), int64(2)

	t.Run("pedido", func(t *testing.T) {
		s := NewDescuentoService(ormerSinDescuentos(nil, func(interface{}, ...string) error { return orm.ErrNoRows }))
		_, err := s.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdCupon: &cuponID})
		assert.ErrorIs(t, err, ErrPedidoNoEncontrado)
	})
	t.Run("cupon", func(t *testing.T) {
		n := 0
		s := NewDescuentoService(ormerSinDescuentos(nil, func(interface{}, ...string) error {
			n++
			if n == 2 {
				return orm.ErrNoRows
			}
			return nil
		}))
		_, err := s.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdCupon: &cuponID})
		assert.ErrorIs(t, err, ErrCuponNoEncontrado)
	})
	t.Run("oferta", func(t *testing.T) {
		n := 0
		s := NewDescuentoService(ormerSinDescuentos(nil, func(interface{}, ...string) error {
			n++
			if n == 2 {
				return orm.ErrNoRows
			}
			return nil
		}))
		_, err := s.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdOferta: &ofertaID})
		assert.ErrorIs(t, err, ErrOfertaNoEncontrada)
	})
	t.Run("ya aplicado", func(t *testing.T) {
		s := NewDescuentoService(&mockDescuentoOrmer{
			queryTableFn: func(string) orm.QuerySeter {
				return &mockDescuentoQuerySeter{countFn: func() (int64, error) { return 1, nil }}
			},
		})
		// readFn nil devuelve ErrNoRows: se fuerza lectura exitosa del pedido
		s.ormer.(*mockDescuentoOrmer).readFn = func(interface{}, ...string) error { return nil }
		_, err := s.AplicarDescuento(context.Background(), 1, &models.AplicarDescuentoRequest{PkIdCupon: &cuponID})
		assert.ErrorIs(t, err, ErrDescuentoYaAplicado)
	})
}

func TestDescuentoService_AplicarDescuento_FusionaDetalle(t *testing.T) {
	cuponID := int64(1)
	var insertado *models.PedidoDescuentoAplicado
	s := NewDescuentoService(ormerSinDescuentos(
		func(m interface{}) (int64, error) {
			insertado = m.(*models.PedidoDescuentoAplicado)
			return 1, nil
		},
		func(m interface{}, _ ...string) error {
			if c, ok := m.(*models.Cupon); ok {
				c.Codigo = "VERANO10"
				c.Scope = models.CuponScopeGlobal
			}
			return nil
		},
	))
	res, err := s.AplicarDescuento(context.Background(), 5, &models.AplicarDescuentoRequest{
		PkIdCupon: &cuponID, MontoDescuento: 100, Detalle: json.RawMessage(`{"nota":"hola","codigo":"otro"}`),
	})
	require.NoError(t, err)
	require.Same(t, insertado, res)

	var det map[string]any
	require.NoError(t, json.Unmarshal(res.DetalleObj, &det))
	assert.Equal(t, "hola", det["nota"])
	assert.Equal(t, "VERANO10", det["codigo"], "los datos del servidor prevalecen")
	assert.Equal(t, "cupon", det["tipo"])
	assert.JSONEq(t, string(res.DetalleObj), res.Detalle, "BeforeInsert debe serializar el detalle para la columna jsonb")
}

func TestDescuentoService_AplicarDescuento_OfertaDeserializaDias(t *testing.T) {
	ofertaID := int64(2)
	s := NewDescuentoService(ormerSinDescuentos(
		func(interface{}) (int64, error) { return 1, nil },
		func(m interface{}, _ ...string) error {
			if o, ok := m.(*models.Oferta); ok {
				o.Titulo = "2x1"
				o.DiasSemana = "{Lunes,Martes}"
			}
			return nil
		},
	))
	res, err := s.AplicarDescuento(context.Background(), 5, &models.AplicarDescuentoRequest{PkIdOferta: &ofertaID})
	require.NoError(t, err)
	assert.Equal(t, []string{"Lunes", "Martes"}, res.PkIdOferta.DiasSemanaArray)
}

func TestDescuentoService_ObtenerDescuentosPedido_VacioYRelaciones(t *testing.T) {
	var sel []interface{}
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				relatedSelFn: func(p ...interface{}) orm.QuerySeter {
					sel = p
					return &mockDescuentoQuerySeter{}
				},
			}
		},
	})
	res, err := service.ObtenerDescuentosPedido(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, res, "nunca nil para serializar []")
	assert.Empty(t, res)
	assert.Equal(t, []interface{}{"PkIdPedido", "PkIdCupon", "PkIdOferta"}, sel, "no debe cargar el cliente del pedido")
}

func TestDescuentoService_ObtenerDescuentosPedido_AfterLoad(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				allFn: func(c interface{}, _ ...string) (int64, error) {
					*(c.(*[]*models.PedidoDescuentoAplicado)) = []*models.PedidoDescuentoAplicado{
						{Detalle: `{"tipo":"oferta"}`, PkIdOferta: &models.Oferta{DiasSemana: "{Sábado}"}},
						{Detalle: `{"tipo":"cupon"}`},
					}
					return 2, nil
				},
			}
		},
	})
	res, err := service.ObtenerDescuentosPedido(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.JSONEq(t, `{"tipo":"oferta"}`, string(res[0].DetalleObj))
	assert.Equal(t, []string{"Sábado"}, res[0].PkIdOferta.DiasSemanaArray)
	assert.Nil(t, res[1].PkIdOferta)
}
