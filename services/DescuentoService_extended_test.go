package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/client/orm/clauses/order_clause"
	"github.com/beego/beego/v2/core/utils"
	"github.com/stretchr/testify/assert"
)

type mockDescuentoOrmer struct {
	readFn       func(interface{}, ...string) error
	insertFn     func(interface{}) (int64, error)
	queryTableFn func(string) orm.QuerySeter
}

func (m *mockDescuentoOrmer) Read(md interface{}, cols ...string) error {
	if m.readFn != nil {
		return m.readFn(md, cols...)
	}
	return orm.ErrNoRows
}

func (m *mockDescuentoOrmer) Insert(md interface{}) (int64, error) {
	if m.insertFn != nil {
		return m.insertFn(md)
	}
	return 1, nil
}

func (m *mockDescuentoOrmer) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	if m.queryTableFn != nil {
		if tableName, ok := ptrStructOrTableName.(string); ok {
			return m.queryTableFn(tableName)
		}
	}
	return &mockDescuentoQuerySeter{}
}

func (m *mockDescuentoOrmer) ReadForUpdate(md interface{}, cols ...string) error {
	return m.Read(md, cols...)
}
func (m *mockDescuentoOrmer) ReadOrCreate(md interface{}, col1 string, cols ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockDescuentoOrmer) InsertMulti(int, interface{}) (int64, error)  { return 0, nil }
func (m *mockDescuentoOrmer) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockDescuentoOrmer) Delete(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockDescuentoOrmer) LoadRelated(md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) QueryM2M(interface{}, string) orm.QueryM2Mer { return nil }
func (m *mockDescuentoOrmer) Begin() (orm.TxOrmer, error)                 { return nil, nil }
func (m *mockDescuentoOrmer) BeginWithCtx(context.Context) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockDescuentoOrmer) BeginWithOpts(opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockDescuentoOrmer) BeginWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockDescuentoOrmer) DoTx(task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockDescuentoOrmer) DoTxWithCtx(ctx context.Context, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockDescuentoOrmer) DoTxWithOpts(opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockDescuentoOrmer) DoTxWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockDescuentoOrmer) Raw(string, ...interface{}) orm.RawSeter { return nil }
func (m *mockDescuentoOrmer) Driver() orm.Driver                      { return nil }
func (m *mockDescuentoOrmer) DBStats() *sql.DBStats                   { return nil }
func (m *mockDescuentoOrmer) ReadWithCtx(context.Context, interface{}, ...string) error {
	return m.Read(nil)
}
func (m *mockDescuentoOrmer) ReadForUpdateWithCtx(context.Context, interface{}, ...string) error {
	return m.Read(nil)
}
func (m *mockDescuentoOrmer) ReadOrCreateWithCtx(context.Context, interface{}, string, ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockDescuentoOrmer) InsertWithCtx(context.Context, interface{}) (int64, error) {
	return m.Insert(nil)
}
func (m *mockDescuentoOrmer) InsertMultiWithCtx(context.Context, int, interface{}) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) InsertOrUpdate(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockDescuentoOrmer) InsertOrUpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) UpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) DeleteWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) LoadRelatedWithCtx(ctx context.Context, md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoOrmer) QueryM2MWithCtx(context.Context, interface{}, string) orm.QueryM2Mer {
	return nil
}
func (m *mockDescuentoOrmer) QueryTableWithCtx(context.Context, interface{}) orm.QuerySeter {
	return m.QueryTable(nil)
}
func (m *mockDescuentoOrmer) RawWithCtx(context.Context, string, ...interface{}) orm.RawSeter {
	return nil
}

type mockDescuentoQuerySeter struct {
	filterFn     func(string, ...interface{}) orm.QuerySeter
	countFn      func() (int64, error)
	relatedSelFn func(...interface{}) orm.QuerySeter
	allFn        func(interface{}, ...string) (int64, error)
}

func (m *mockDescuentoQuerySeter) Filter(expr string, args ...interface{}) orm.QuerySeter {
	if m.filterFn != nil {
		return m.filterFn(expr, args...)
	}
	return m
}

func (m *mockDescuentoQuerySeter) Count() (int64, error) {
	if m.countFn != nil {
		return m.countFn()
	}
	return 0, nil
}

func (m *mockDescuentoQuerySeter) RelatedSel(params ...interface{}) orm.QuerySeter {
	if m.relatedSelFn != nil {
		return m.relatedSelFn(params...)
	}
	return m
}

func (m *mockDescuentoQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	if m.allFn != nil {
		return m.allFn(container, cols...)
	}
	return 0, nil
}

func (m *mockDescuentoQuerySeter) Limit(limit interface{}, args ...interface{}) orm.QuerySeter {
	return m
}
func (m *mockDescuentoQuerySeter) Offset(offset interface{}) orm.QuerySeter { return m }
func (m *mockDescuentoQuerySeter) OrderBy(exprs ...string) orm.QuerySeter   { return m }
func (m *mockDescuentoQuerySeter) Distinct() orm.QuerySeter                 { return m }
func (m *mockDescuentoQuerySeter) SetCond(*orm.Condition) orm.QuerySeter    { return m }
func (m *mockDescuentoQuerySeter) GetCond() *orm.Condition                  { return nil }
func (m *mockDescuentoQuerySeter) One(container interface{}, cols ...string) error {
	return orm.ErrNoRows
}
func (m *mockDescuentoQuerySeter) Exist() bool                             { return false }
func (m *mockDescuentoQuerySeter) Update(values orm.Params) (int64, error) { return 0, nil }
func (m *mockDescuentoQuerySeter) Delete() (int64, error)                  { return 0, nil }
func (m *mockDescuentoQuerySeter) PrepareInsert() (orm.Inserter, error)    { return nil, nil }
func (m *mockDescuentoQuerySeter) Values(results *[]orm.Params, exprs ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) ValuesList(results *[]orm.ParamsList, exprs ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) ValuesFlat(result *orm.ParamsList, expr string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) RowsToMap(result *orm.Params, keyCol, valueCol string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) RowsToStruct(ptrStruct interface{}, keyCol, valueCol string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) GroupBy(exprs ...string) orm.QuerySeter             { return m }
func (m *mockDescuentoQuerySeter) ForUpdate() orm.QuerySeter                          { return m }
func (m *mockDescuentoQuerySeter) Aggregate(s string) orm.QuerySeter                  { return m }
func (m *mockDescuentoQuerySeter) FilterRaw(string, string) orm.QuerySeter            { return m }
func (m *mockDescuentoQuerySeter) Exclude(string, ...interface{}) orm.QuerySeter      { return m }
func (m *mockDescuentoQuerySeter) OrderClauses(...*order_clause.Order) orm.QuerySeter { return m }
func (m *mockDescuentoQuerySeter) ForceIndex(...string) orm.QuerySeter                { return m }
func (m *mockDescuentoQuerySeter) UseIndex(...string) orm.QuerySeter                  { return m }
func (m *mockDescuentoQuerySeter) IgnoreIndex(...string) orm.QuerySeter               { return m }

func (m *mockDescuentoQuerySeter) OneWithCtx(context.Context, interface{}, ...string) error {
	return m.One(nil)
}
func (m *mockDescuentoQuerySeter) AllWithCtx(ctx context.Context, container interface{}, cols ...string) (int64, error) {
	return m.All(container, cols...)
}
func (m *mockDescuentoQuerySeter) ValuesWithCtx(context.Context, *[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) ValuesListWithCtx(context.Context, *[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) ValuesFlatWithCtx(context.Context, *orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) CountWithCtx(ctx context.Context) (int64, error) {
	return m.Count()
}
func (m *mockDescuentoQuerySeter) ExistWithCtx(context.Context) bool { return false }
func (m *mockDescuentoQuerySeter) UpdateWithCtx(context.Context, orm.Params) (int64, error) {
	return 0, nil
}
func (m *mockDescuentoQuerySeter) DeleteWithCtx(context.Context) (int64, error) { return 0, nil }
func (m *mockDescuentoQuerySeter) PrepareInsertWithCtx(context.Context) (orm.Inserter, error) {
	return nil, nil
}

func TestDescuentoService_AplicarDescuento_NoDescuentoEspecificado(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{})

	req := &models.AplicarDescuentoRequest{
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "debe especificar exactamente uno de cupón o oferta")
}

func TestDescuentoService_AplicarDescuento_AmbosCuponYOferta(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{})

	cuponId := int64(1)
	ofertaId := int64(2)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		PkIdOferta:     &ofertaId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "debe especificar exactamente uno de cupón o oferta")
}

func TestDescuentoService_AplicarDescuento_PedidoNoEncontrado(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return orm.ErrNoRows
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 999, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "pedido no encontrado")
}

func TestDescuentoService_AplicarDescuento_ErrorBuscarPedido(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return assert.AnError
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al buscar pedido")
}

func TestDescuentoService_AplicarDescuento_ErrorVerificarDescuentosExistentes(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			return nil
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, assert.AnError
						},
					}
				},
			}
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al verificar descuentos existentes")
}

func TestDescuentoService_AplicarDescuento_DescuentoYaExiste(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 1, nil
						},
					}
				},
			}
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ya existe un descuento aplicado para este pedido")
}

func TestDescuentoService_AplicarDescuento_CuponNoEncontrado(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {

				return nil
			}

			return orm.ErrNoRows
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	cuponId := int64(999)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cupón no encontrado")
}

func TestDescuentoService_AplicarDescuento_ErrorBuscarCupon(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {
				return nil
			}
			return assert.AnError
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al buscar cupón")
}

func TestDescuentoService_AplicarDescuento_OfertaNoEncontrada(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {
				return nil
			}
			return orm.ErrNoRows
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertaId := int64(999)
	req := &models.AplicarDescuentoRequest{
		PkIdOferta:     &ofertaId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "oferta no encontrada")
}

func TestDescuentoService_AplicarDescuento_ErrorBuscarOferta(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {
				return nil
			}
			return assert.AnError
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertaId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdOferta:     &ofertaId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al buscar oferta")
}

func TestDescuentoService_AplicarDescuento_ErrorInsert(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {

				if pedido, ok := md.(*models.Pedido); ok {
					pedido.PK_ID_PEDIDO = 1
				}
				return nil
			}

			if cupon, ok := md.(*models.Cupon); ok {
				cupon.PkIdCupon = 1
				cupon.Codigo = "TEST"
				cupon.Scope = "GLOBAL"
			}
			return nil
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
		insertFn: func(md interface{}) (int64, error) {
			return 0, assert.AnError
		},
	})

	cuponId := int64(1)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al registrar descuento aplicado")
}

func TestDescuentoService_AplicarDescuento_ExitoCupon(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {
				if pedido, ok := md.(*models.Pedido); ok {
					pedido.PK_ID_PEDIDO = 1
				}
				return nil
			}
			if cupon, ok := md.(*models.Cupon); ok {
				cupon.PkIdCupon = 1
				cupon.Codigo = "DESCUENTO10"
				cupon.Scope = "GLOBAL"
			}
			return nil
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
		insertFn: func(md interface{}) (int64, error) {
			return 1, nil
		},
	})

	cuponId := int64(1)
	detalle := json.RawMessage(`{"info":"test"}`)
	req := &models.AplicarDescuentoRequest{
		PkIdCupon:      &cuponId,
		MontoDescuento: 1000,
		Detalle:        detalle,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1000), result.MontoDescuento)
	assert.NotNil(t, result.PkIdCupon)
	assert.Equal(t, int64(1), result.PkIdCupon.PkIdCupon)
}

func TestDescuentoService_AplicarDescuento_ExitoOferta(t *testing.T) {
	readCallCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		readFn: func(md interface{}, cols ...string) error {
			readCallCount++
			if readCallCount == 1 {
				if pedido, ok := md.(*models.Pedido); ok {
					pedido.PK_ID_PEDIDO = 1
				}
				return nil
			}
			if oferta, ok := md.(*models.Oferta); ok {
				oferta.PkIdOferta = 1
				oferta.Titulo = "Oferta 2x1"
			}
			return nil
		},
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
		insertFn: func(md interface{}) (int64, error) {
			return 1, nil
		},
	})

	ofertaId := int64(1)
	detalle := json.RawMessage(`{"info":"test"}`)
	req := &models.AplicarDescuentoRequest{
		PkIdOferta:     &ofertaId,
		MontoDescuento: 2000,
		Detalle:        detalle,
	}

	result, err := service.AplicarDescuento(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2000), result.MontoDescuento)
	assert.NotNil(t, result.PkIdOferta)
	assert.Equal(t, int64(1), result.PkIdOferta.PkIdOferta)
}

func TestDescuentoService_ObtenerDescuentosPedido_Error(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						relatedSelFn: func(params ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								allFn: func(container interface{}, cols ...string) (int64, error) {
									return 0, assert.AnError
								},
							}
						},
					}
				},
			}
		},
	})

	result, err := service.ObtenerDescuentosPedido(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al obtener descuentos del pedido")
}

func TestDescuentoService_ObtenerDescuentosPedido_Exito(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						relatedSelFn: func(params ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								allFn: func(container interface{}, cols ...string) (int64, error) {

									if descuentos, ok := container.(*[]*models.PedidoDescuentoAplicado); ok {
										*descuentos = []*models.PedidoDescuentoAplicado{
											{MontoDescuento: 1000},
											{MontoDescuento: 2000},
										}
									}
									return 2, nil
								},
							}
						},
					}
				},
			}
		},
	})

	result, err := service.ObtenerDescuentosPedido(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

func TestDescuentoService_ValidarExclusividadDescuento_ErrorCount(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, assert.AnError
						},
					}
				},
			}
		},
	})

	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al verificar descuentos existentes")
}

func TestDescuentoService_ValidarExclusividadDescuento_YaExiste(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 1, nil
						},
					}
				},
			}
		},
	})

	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ya existe un descuento aplicado para este pedido")
}

func TestDescuentoService_ValidarExclusividadDescuento_CuponDuplicado_Error(t *testing.T) {
	callCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			qs := &mockDescuentoQuerySeter{}

			qs.filterFn = func(expr string, args ...interface{}) orm.QuerySeter {
				qs2 := &mockDescuentoQuerySeter{}
				qs2.filterFn = func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							if callCount == 2 {
								return 0, assert.AnError
							}
							return 0, nil
						},
					}
				}
				qs2.countFn = func() (int64, error) {
					return 0, nil
				}
				return qs2
			}
			return qs
		},
	})

	cuponId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, &cuponId, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al verificar cupón duplicado")
}

func TestDescuentoService_ValidarExclusividadDescuento_CuponYaAplicado(t *testing.T) {
	callCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								countFn: func() (int64, error) {
									if callCount == 2 {
										return 1, nil
									}
									return 0, nil
								},
							}
						},
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	cuponId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, &cuponId, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "este cupón ya ha sido aplicado a este pedido")
}

func TestDescuentoService_ValidarExclusividadDescuento_OfertaYaAplicada(t *testing.T) {
	callCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								countFn: func() (int64, error) {
									if callCount == 2 {
										return 1, nil
									}
									return 0, nil
								},
							}
						},
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertaId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, &ofertaId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "esta oferta ya ha sido aplicada a este pedido")
}

func TestDescuentoService_ValidarExclusividadDescuento_OfertaDuplicada_Error(t *testing.T) {
	callCount := 0
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								countFn: func() (int64, error) {
									if callCount == 2 {
										return 0, assert.AnError
									}
									return 0, nil
								},
							}
						},
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertaId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, &ofertaId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al verificar oferta duplicada")
}

func TestDescuentoService_ValidarExclusividadDescuento_Exito_SinDescuentos(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, nil)
	assert.NoError(t, err)
}

func TestDescuentoService_ValidarExclusividadDescuento_Exito_ConCupon(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								countFn: func() (int64, error) {
									return 0, nil
								},
							}
						},
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	cuponId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, &cuponId, nil)
	assert.NoError(t, err)
}

func TestDescuentoService_ValidarExclusividadDescuento_Exito_ConOferta(t *testing.T) {
	service := NewDescuentoService(&mockDescuentoOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockDescuentoQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockDescuentoQuerySeter{
						filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
							return &mockDescuentoQuerySeter{
								countFn: func() (int64, error) {
									return 0, nil
								},
							}
						},
						countFn: func() (int64, error) {
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertaId := int64(1)
	err := service.ValidarExclusividadDescuento(context.Background(), 1, nil, &ofertaId)
	assert.NoError(t, err)
}
