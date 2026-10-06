package services

import (
	"context"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
