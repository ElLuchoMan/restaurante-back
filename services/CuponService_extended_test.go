package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/client/orm/clauses/order_clause"
	"github.com/stretchr/testify/assert"
)

type dummyQuerySeter struct {
	filterCalls  int
	relatedCalls int
	oneCalled    bool
	countCalled  bool
}

func (d *dummyQuerySeter) Filter(string, ...interface{}) orm.QuerySeter             { d.filterCalls++; return d }
func (d *dummyQuerySeter) FilterRaw(string, string) orm.QuerySeter                  { return d }
func (d *dummyQuerySeter) Exclude(string, ...interface{}) orm.QuerySeter            { return d }
func (d *dummyQuerySeter) SetCond(*orm.Condition) orm.QuerySeter                    { return d }
func (d *dummyQuerySeter) GetCond() *orm.Condition                                  { return nil }
func (d *dummyQuerySeter) Limit(interface{}, ...interface{}) orm.QuerySeter         { return d }
func (d *dummyQuerySeter) Offset(interface{}) orm.QuerySeter                        { return d }
func (d *dummyQuerySeter) GroupBy(...string) orm.QuerySeter                         { return d }
func (d *dummyQuerySeter) OrderBy(...string) orm.QuerySeter                         { return d }
func (d *dummyQuerySeter) OrderClauses(...*order_clause.Order) orm.QuerySeter       { return d }
func (d *dummyQuerySeter) ForceIndex(...string) orm.QuerySeter                      { return d }
func (d *dummyQuerySeter) UseIndex(...string) orm.QuerySeter                        { return d }
func (d *dummyQuerySeter) IgnoreIndex(...string) orm.QuerySeter                     { return d }
func (d *dummyQuerySeter) RelatedSel(...interface{}) orm.QuerySeter                 { d.relatedCalls++; return d }
func (d *dummyQuerySeter) Distinct() orm.QuerySeter                                 { return d }
func (d *dummyQuerySeter) ForUpdate() orm.QuerySeter                                { return d }
func (d *dummyQuerySeter) Count() (int64, error)                                    { d.countCalled = true; return 5, nil }
func (d *dummyQuerySeter) CountWithCtx(context.Context) (int64, error)              { return d.Count() }
func (d *dummyQuerySeter) Exist() bool                                              { return true }
func (d *dummyQuerySeter) ExistWithCtx(context.Context) bool                        { return true }
func (d *dummyQuerySeter) Update(orm.Params) (int64, error)                         { return 0, nil }
func (d *dummyQuerySeter) UpdateWithCtx(context.Context, orm.Params) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) Delete() (int64, error)                                   { return 0, nil }
func (d *dummyQuerySeter) DeleteWithCtx(context.Context) (int64, error)             { return 0, nil }
func (d *dummyQuerySeter) PrepareInsert() (orm.Inserter, error)                     { return nil, nil }
func (d *dummyQuerySeter) PrepareInsertWithCtx(context.Context) (orm.Inserter, error) {
	return nil, nil
}
func (d *dummyQuerySeter) All(interface{}, ...string) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) AllWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (d *dummyQuerySeter) One(interface{}, ...string) error { d.oneCalled = true; return nil }
func (d *dummyQuerySeter) OneWithCtx(context.Context, interface{}, ...string) error {
	return d.One(nil)
}
func (d *dummyQuerySeter) Values(*[]orm.Params, ...string) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) ValuesWithCtx(context.Context, *[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (d *dummyQuerySeter) ValuesList(*[]orm.ParamsList, ...string) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) ValuesListWithCtx(context.Context, *[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (d *dummyQuerySeter) ValuesFlat(*orm.ParamsList, string) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) ValuesFlatWithCtx(context.Context, *orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (d *dummyQuerySeter) RowsToMap(*orm.Params, string, string) (int64, error)    { return 0, nil }
func (d *dummyQuerySeter) RowsToStruct(interface{}, string, string) (int64, error) { return 0, nil }
func (d *dummyQuerySeter) Aggregate(string) orm.QuerySeter                         { return d }

func TestCuponService_ValidarCupon_OrmerNil(t *testing.T) {
	service := NewCuponService(nil)
	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_ValidarCupon_CuponNoEncontrado(t *testing.T) {
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						return orm.ErrNoRows
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cupón no encontrado", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_ErrorBuscarCupon(t *testing.T) {
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						return fmt.Errorf("db error")
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_ValidarCupon_CuponInactivo(t *testing.T) {
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{Activo: false}
						return nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cupón inactivo", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_CuponFueraVigencia(t *testing.T) {
	now := time.Now()
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:      true,
							FechaInicio: now.Add(24 * time.Hour),
							FechaFin:    now.Add(48 * time.Hour),
						}
						return nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cupón fuera del período de vigencia", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_MaxUsosSuperado(t *testing.T) {
	now := time.Now()
	maxUsos := 1
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:      true,
							FechaInicio: now.Add(-time.Hour),
							FechaFin:    now.Add(time.Hour),
							MaxUsos:     &maxUsos,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) {
						return 1, nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cupón ha alcanzado el límite máximo de usos", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_ErrorContarUsos(t *testing.T) {
	now := time.Now()
	maxUsos := 2
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:      true,
							FechaInicio: now.Add(-time.Hour),
							FechaFin:    now.Add(time.Hour),
							MaxUsos:     &maxUsos,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) {
						return 0, fmt.Errorf("count error")
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_ValidarCupon_LimiteClienteSuperado(t *testing.T) {
	now := time.Now()
	limite := 1
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:           true,
							FechaInicio:      now.Add(-time.Hour),
							FechaFin:         now.Add(time.Hour),
							LimitePorCliente: &limite,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) {
						return 1, nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cliente ha alcanzado el límite de usos para este cupón", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_ErrorContarCliente(t *testing.T) {
	now := time.Now()
	limite := 2
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:           true,
							FechaInicio:      now.Add(-time.Hour),
							FechaFin:         now.Add(time.Hour),
							LimitePorCliente: &limite,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) {
						return 0, fmt.Errorf("cliente count error")
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{Codigo: "TEST", ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_ValidarCupon_MontoMinimo(t *testing.T) {
	now := time.Now()
	minimo := int64(500)
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:      true,
							FechaInicio: now.Add(-time.Hour),
							FechaFin:    now.Add(time.Hour),
							MontoMinimo: &minimo,
							Scope:       models.CuponScopeGlobal,
						}
						return nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 1,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 100},
		},
	})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Contains(t, *resp.Motivo, "El monto mínimo requerido es")
	}
}

func TestCuponService_ValidarCupon_ScopeClienteInvalido(t *testing.T) {
	now := time.Now()
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:             true,
							FechaInicio:        now.Add(-time.Hour),
							FechaFin:           now.Add(time.Hour),
							Scope:              models.CuponScopeCliente,
							PkDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: 99},
						}
						return nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 1,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 100},
		},
	})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "Cupón no válido para este cliente", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_ScopeProductoSinAplicables(t *testing.T) {
	now := time.Now()
	product := &models.Producto{PK_ID_PRODUCTO: 99}
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:       true,
							FechaInicio:  now.Add(-time.Hour),
							FechaFin:     now.Add(time.Hour),
							Scope:        models.CuponScopeProducto,
							PkIdProducto: product,
						}
						return nil
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 1,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 100},
		},
	})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "No hay productos aplicables para este cupón", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_ScopeCategoriaSinAplicables(t *testing.T) {
	now := time.Now()
	categoria := &models.Categoria{PK_ID_CATEGORIA: 5}
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:        true,
							FechaInicio:   now.Add(-time.Hour),
							FechaFin:      now.Add(time.Hour),
							Scope:         models.CuponScopeCategoria,
							PkIdCategoria: categoria,
						}
						return nil
					},
				}
			},
			"producto": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						return fmt.Errorf("producto no encontrado")
					},
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 1,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 100},
		},
	})
	assert.NoError(t, err)
	assert.False(t, resp.Aplicable)
	if assert.NotNil(t, resp.Motivo) {
		assert.Equal(t, "No hay productos aplicables para este cupón", *resp.Motivo)
	}
}

func TestCuponService_ValidarCupon_Success(t *testing.T) {
	now := time.Now()
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							Activo:         true,
							FechaInicio:    now.Add(-time.Hour),
							FechaFin:       now.Add(time.Hour),
							Scope:          models.CuponScopeGlobal,
							TipoDescuento:  models.TipoDescuentoPorcentaje,
							ValorDescuento: 10,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) { return 0, nil },
				}
			},
		},
	})

	resp, err := service.ValidarCupon(context.Background(), &models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 1,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 2, Precio: 100},
		},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Aplicable)
	assert.Equal(t, int64(20), resp.MontoDescuento)
}

func TestCuponService_RedimirCupon_OrmerNil(t *testing.T) {
	service := NewCuponService(nil)
	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_RedimirCupon_CuponNoEncontrado(t *testing.T) {
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						return fmt.Errorf("no encontrado")
					},
				}
			},
		},
	})

	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_RedimirCupon_ValidacionError(t *testing.T) {
	now := time.Now()
	maxUsos := 1
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							PkIdCupon:   10,
							Activo:      true,
							FechaInicio: now.Add(-time.Hour),
							FechaFin:    now.Add(time.Hour),
							MaxUsos:     &maxUsos,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) {
						return 0, fmt.Errorf("error count")
					},
				}
			},
		},
	})

	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_RedimirCupon_NoAplicable(t *testing.T) {
	now := time.Now()
	minimo := int64(500)
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							PkIdCupon:   10,
							Activo:      true,
							FechaInicio: now.Add(-time.Hour),
							FechaFin:    now.Add(time.Hour),
							Scope:       models.CuponScopeGlobal,
							MontoMinimo: &minimo,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) { return 0, nil },
				}
			},
		},
	})

	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_RedimirCupon_InsertError(t *testing.T) {
	now := time.Now()
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							PkIdCupon:      10,
							Activo:         true,
							FechaInicio:    now.Add(-time.Hour),
							FechaFin:       now.Add(time.Hour),
							Scope:          models.CuponScopeGlobal,
							TipoDescuento:  models.TipoDescuentoPorcentaje,
							ValorDescuento: 10,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) { return 0, nil },
				}
			},
		},
		insertFn: func(model interface{}) (int64, error) {
			return 0, fmt.Errorf("insert error")
		},
	})

	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCuponService_RedimirCupon_Success(t *testing.T) {
	now := time.Now()
	pedidoID := int64(22)
	service := NewCuponService(&mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"cupon": func() cuponQuerySeter {
				return &mockQuerySeter{
					oneHook: func(dest interface{}, _ ...string) error {
						cupon := dest.(*models.Cupon)
						*cupon = models.Cupon{
							PkIdCupon:      10,
							Activo:         true,
							FechaInicio:    now.Add(-time.Hour),
							FechaFin:       now.Add(time.Hour),
							Scope:          models.CuponScopeGlobal,
							TipoDescuento:  models.TipoDescuentoMonto,
							ValorDescuento: 5,
						}
						return nil
					},
				}
			},
			"cupon_redencion": func() cuponQuerySeter {
				return &mockQuerySeter{
					countHook: func() (int64, error) { return 0, nil },
				}
			},
		},
		insertFn: func(model interface{}) (int64, error) {
			return 1, nil
		},
	})

	resp, err := service.RedimirCupon(context.Background(), "CODE", &models.RedimirCuponRequest{ClienteId: 1, PedidoId: &pedidoID})
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int64(0), resp.MontoDescuento)
		if assert.NotNil(t, resp.PkIdPedido) {
			assert.Equal(t, pedidoID, resp.PkIdPedido.PK_ID_PEDIDO)
		}
	}
}

func TestCuponService_EsProductoAplicable_Categoria_Success(t *testing.T) {
	categoria := &models.Categoria{PK_ID_CATEGORIA: 5}
	service := &CuponService{ormer: &mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"producto": func() cuponQuerySeter {
				producto := &mockQuerySeter{}
				producto.oneHook = func(dest interface{}, _ ...string) error {
					prod := dest.(*models.Producto)
					prod.PK_ID_PRODUCTO = 10
					prod.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: 7}
					return nil
				}
				producto.relatedSelHook = func(...interface{}) cuponQuerySeter { return producto }
				return producto
			},
			"subcategoria": func() cuponQuerySeter {
				qs := &mockQuerySeter{}
				qs.oneHook = func(dest interface{}, _ ...string) error {
					sub := dest.(*models.Subcategoria)
					sub.PK_ID_CATEGORIA = categoria
					return nil
				}
				qs.relatedSelHook = func(...interface{}) cuponQuerySeter { return qs }
				return qs
			},
		},
	}}

	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: categoria}
	assert.True(t, service.esProductoAplicable(cupon, 10))
}

func TestCuponService_EsProductoAplicable_Categoria_SubcategoriaNil(t *testing.T) {
	categoria := &models.Categoria{PK_ID_CATEGORIA: 5}
	service := &CuponService{ormer: &mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"producto": func() cuponQuerySeter {
				producto := &mockQuerySeter{}
				producto.oneHook = func(dest interface{}, _ ...string) error {
					prod := dest.(*models.Producto)
					prod.PK_ID_PRODUCTO = 10
					prod.PK_ID_SUBCATEGORIA = nil
					return nil
				}
				producto.relatedSelHook = func(...interface{}) cuponQuerySeter { return producto }
				return producto
			},
		},
	}}

	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: categoria}
	assert.False(t, service.esProductoAplicable(cupon, 10))
}

func TestCuponService_EsProductoAplicable_Categoria_SubcategoriaError(t *testing.T) {
	categoria := &models.Categoria{PK_ID_CATEGORIA: 5}
	service := &CuponService{ormer: &mockCuponOrmer{
		tables: map[string]func() cuponQuerySeter{
			"producto": func() cuponQuerySeter {
				producto := &mockQuerySeter{}
				producto.oneHook = func(dest interface{}, _ ...string) error {
					prod := dest.(*models.Producto)
					prod.PK_ID_PRODUCTO = 10
					prod.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: 7}
					return nil
				}
				producto.relatedSelHook = func(...interface{}) cuponQuerySeter { return producto }
				return producto
			},
			"subcategoria": func() cuponQuerySeter {
				qs := &mockQuerySeter{}
				qs.oneHook = func(dest interface{}, _ ...string) error {
					return fmt.Errorf("subcategoria error")
				}
				qs.relatedSelHook = func(...interface{}) cuponQuerySeter { return qs }
				return qs
			},
		},
	}}

	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: categoria}
	assert.False(t, service.esProductoAplicable(cupon, 10))
}

func TestCuponService_EsProductoAplicable_Categoria_SinOrmer(t *testing.T) {
	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: &models.Categoria{PK_ID_CATEGORIA: 1}}
	service := &CuponService{}
	assert.False(t, service.esProductoAplicable(cupon, 1))
}

func TestCuponService_EsProductoAplicable_Categoria_SinCategoria(t *testing.T) {
	cupon := &models.Cupon{Scope: models.CuponScopeCategoria, PkIdCategoria: nil}
	service := &CuponService{ormer: &mockCuponOrmer{}}
	assert.False(t, service.esProductoAplicable(cupon, 1))
}

func TestBeegoCuponQuerySeterNilBehaviors(t *testing.T) {
	qs := beegoCuponQuerySeter{}
	filtered := qs.Filter("field")
	_, ok := filtered.(beegoCuponQuerySeter)
	assert.True(t, ok)

	err := qs.One(&models.Cupon{})
	assert.ErrorIs(t, err, orm.ErrNoRows)

	count, err := qs.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	related := qs.RelatedSel()
	_, ok = related.(beegoCuponQuerySeter)
	assert.True(t, ok)
}

func TestBeegoCuponQuerySeterWithFuncs(t *testing.T) {
	called := make(map[string]bool)
	qs := beegoCuponQuerySeter{
		filterFunc: func(field string, args ...interface{}) cuponQuerySeter {
			called["filter"] = field == "foo"
			return beegoCuponQuerySeter{}
		},
		oneFunc: func(dest interface{}, cols ...string) error {
			called["one"] = true
			return nil
		},
		countFunc: func() (int64, error) {
			called["count"] = true
			return 7, nil
		},
		relatedSelFunc: func(params ...interface{}) cuponQuerySeter {
			called["related"] = true
			return beegoCuponQuerySeter{}
		},
	}

	_ = qs.Filter("foo")
	assert.True(t, called["filter"])

	assert.NoError(t, qs.One(&models.Cupon{}))
	assert.True(t, called["one"])

	count, err := qs.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(7), count)
	assert.True(t, called["count"])

	_ = qs.RelatedSel()
	assert.True(t, called["related"])
}

func TestBeegoCuponOrmer_QueryTable(t *testing.T) {
	called := false
	ormer := beegoCuponOrmer{
		queryTable: func(name string) cuponQuerySeter {
			called = name == "cupon"
			return beegoCuponQuerySeter{}
		},
	}
	_ = ormer.QueryTable("cupon")
	assert.True(t, called)
}

func TestBeegoCuponOrmer_QueryTable_Nil(t *testing.T) {
	ormer := beegoCuponOrmer{}
	qs := ormer.QueryTable("cupon")
	assert.IsType(t, beegoCuponQuerySeter{}, qs)
}

func TestBeegoCuponOrmer_Insert_NoFunc(t *testing.T) {
	_, err := beegoCuponOrmer{}.Insert(&models.Cupon{})
	assert.Error(t, err)
}

func TestWrapOrmQuerySeter(t *testing.T) {
	dummy := &dummyQuerySeter{}
	wrapped := wrapOrmQuerySeter(dummy)

	next := wrapped.Filter("filtro")
	assert.IsType(t, beegoCuponQuerySeter{}, next)
	assert.Equal(t, 1, dummy.filterCalls)

	assert.NoError(t, wrapped.One(&models.Cupon{}))
	assert.True(t, dummy.oneCalled)

	count, err := wrapped.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.True(t, dummy.countCalled)

	wrapped.RelatedSel()
	assert.Equal(t, 1, dummy.relatedCalls)
}

func TestNewCuponOrmerFromFuncs(t *testing.T) {
	calledQuery := false
	calledInsert := false

	ormer := NewCuponOrmerFromFuncs(
		func(name string) orm.QuerySeter {
			calledQuery = true
			assert.Equal(t, "cupon", name)
			return nil
		},
		func(model interface{}) (int64, error) {
			calledInsert = true
			return 1, nil
		},
	)

	service := NewCuponService(ormer)
	assert.NotNil(t, service)

	// QueryTable with nil returns default query seter
	qs := service.ormer.QueryTable("cupon")
	assert.IsType(t, beegoCuponQuerySeter{}, qs)
	_, _ = service.ormer.Insert(&models.Cupon{})

	assert.True(t, calledQuery)
	assert.True(t, calledInsert)
}

func TestNewCuponOrmerFromFuncs_Nil(t *testing.T) {
	ormer := NewCuponOrmerFromFuncs(nil, nil)
	assert.Nil(t, ormer)
}

func TestCuponService_ValidarReglasNegocioCupon_FechasInvalidas(t *testing.T) {
	service := &CuponService{}
	cupon := &models.Cupon{
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 10,
		Scope:          models.CuponScopeGlobal,
		FechaInicio:    time.Now(),
		FechaFin:       time.Now().Add(-time.Hour),
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fecha de fin")
}

func TestCuponService_ValidarReglasNegocioCupon_ScopeGlobalConTargets(t *testing.T) {
	service := &CuponService{}
	cupon := &models.Cupon{
		TipoDescuento:      models.TipoDescuentoPorcentaje,
		ValorDescuento:     10,
		Scope:              models.CuponScopeGlobal,
		FechaInicio:        time.Now(),
		FechaFin:           time.Now().Add(time.Hour),
		PkIdProducto:       &models.Producto{PK_ID_PRODUCTO: 1},
		PkIdCategoria:      &models.Categoria{PK_ID_CATEGORIA: 2},
		PkDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: 3},
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GLOBAL")
}
