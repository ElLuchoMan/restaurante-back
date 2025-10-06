package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/client/orm/clauses/order_clause"
	"github.com/beego/beego/v2/core/utils"
	"github.com/stretchr/testify/assert"
)

// Mock para OfertaService
type mockOfertaOrmer struct {
	queryTableFn func(string) orm.QuerySeter
}

func (m *mockOfertaOrmer) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	if m.queryTableFn != nil {
		if tableName, ok := ptrStructOrTableName.(string); ok {
			return m.queryTableFn(tableName)
		}
	}
	return &mockOfertaQuerySeter{}
}

// Implementar métodos restantes de orm.Ormer
func (m *mockOfertaOrmer) Read(interface{}, ...string) error          { return orm.ErrNoRows }
func (m *mockOfertaOrmer) ReadForUpdate(interface{}, ...string) error { return orm.ErrNoRows }
func (m *mockOfertaOrmer) ReadOrCreate(interface{}, string, ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockOfertaOrmer) Insert(interface{}) (int64, error)            { return 0, nil }
func (m *mockOfertaOrmer) InsertMulti(int, interface{}) (int64, error)  { return 0, nil }
func (m *mockOfertaOrmer) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockOfertaOrmer) Delete(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockOfertaOrmer) LoadRelated(md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) QueryM2M(interface{}, string) orm.QueryM2Mer { return nil }
func (m *mockOfertaOrmer) Begin() (orm.TxOrmer, error)                 { return nil, nil }
func (m *mockOfertaOrmer) BeginWithCtx(context.Context) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockOfertaOrmer) BeginWithOpts(opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockOfertaOrmer) BeginWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockOfertaOrmer) DoTx(task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockOfertaOrmer) DoTxWithCtx(ctx context.Context, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockOfertaOrmer) DoTxWithOpts(opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockOfertaOrmer) DoTxWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockOfertaOrmer) Raw(string, ...interface{}) orm.RawSeter { return nil }
func (m *mockOfertaOrmer) Driver() orm.Driver                      { return nil }
func (m *mockOfertaOrmer) DBStats() *sql.DBStats                   { return nil }
func (m *mockOfertaOrmer) ReadWithCtx(context.Context, interface{}, ...string) error {
	return orm.ErrNoRows
}
func (m *mockOfertaOrmer) ReadForUpdateWithCtx(context.Context, interface{}, ...string) error {
	return orm.ErrNoRows
}
func (m *mockOfertaOrmer) ReadOrCreateWithCtx(context.Context, interface{}, string, ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockOfertaOrmer) InsertWithCtx(context.Context, interface{}) (int64, error) { return 0, nil }
func (m *mockOfertaOrmer) InsertMultiWithCtx(context.Context, int, interface{}) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) InsertOrUpdate(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockOfertaOrmer) InsertOrUpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) UpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) DeleteWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) LoadRelatedWithCtx(ctx context.Context, md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockOfertaOrmer) QueryM2MWithCtx(context.Context, interface{}, string) orm.QueryM2Mer {
	return nil
}
func (m *mockOfertaOrmer) QueryTableWithCtx(context.Context, interface{}) orm.QuerySeter {
	return &mockOfertaQuerySeter{}
}
func (m *mockOfertaOrmer) RawWithCtx(context.Context, string, ...interface{}) orm.RawSeter {
	return nil
}

type mockOfertaQuerySeter struct {
	filterFn func(string, ...interface{}) orm.QuerySeter
	allFn    func(interface{}, ...string) (int64, error)
}

func (m *mockOfertaQuerySeter) Filter(expr string, args ...interface{}) orm.QuerySeter {
	if m.filterFn != nil {
		return m.filterFn(expr, args...)
	}
	return m
}

func (m *mockOfertaQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	if m.allFn != nil {
		return m.allFn(container, cols...)
	}
	return 0, nil
}

// Implementar métodos restantes de orm.QuerySeter
func (m *mockOfertaQuerySeter) Limit(interface{}, ...interface{}) orm.QuerySeter { return m }
func (m *mockOfertaQuerySeter) Offset(interface{}) orm.QuerySeter                { return m }
func (m *mockOfertaQuerySeter) OrderBy(...string) orm.QuerySeter                 { return m }
func (m *mockOfertaQuerySeter) Distinct() orm.QuerySeter                         { return m }
func (m *mockOfertaQuerySeter) SetCond(*orm.Condition) orm.QuerySeter            { return m }
func (m *mockOfertaQuerySeter) GetCond() *orm.Condition                          { return nil }
func (m *mockOfertaQuerySeter) One(interface{}, ...string) error                 { return orm.ErrNoRows }
func (m *mockOfertaQuerySeter) Exist() bool                                      { return false }
func (m *mockOfertaQuerySeter) Update(orm.Params) (int64, error)                 { return 0, nil }
func (m *mockOfertaQuerySeter) Delete() (int64, error)                           { return 0, nil }
func (m *mockOfertaQuerySeter) PrepareInsert() (orm.Inserter, error)             { return nil, nil }
func (m *mockOfertaQuerySeter) Values(*[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) ValuesList(*[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) ValuesFlat(*orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) RowsToMap(*orm.Params, string, string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) RowsToStruct(interface{}, string, string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) GroupBy(...string) orm.QuerySeter                   { return m }
func (m *mockOfertaQuerySeter) ForUpdate() orm.QuerySeter                          { return m }
func (m *mockOfertaQuerySeter) Aggregate(string) orm.QuerySeter                    { return m }
func (m *mockOfertaQuerySeter) FilterRaw(string, string) orm.QuerySeter            { return m }
func (m *mockOfertaQuerySeter) Exclude(string, ...interface{}) orm.QuerySeter      { return m }
func (m *mockOfertaQuerySeter) OrderClauses(...*order_clause.Order) orm.QuerySeter { return m }
func (m *mockOfertaQuerySeter) ForceIndex(...string) orm.QuerySeter                { return m }
func (m *mockOfertaQuerySeter) UseIndex(...string) orm.QuerySeter                  { return m }
func (m *mockOfertaQuerySeter) IgnoreIndex(...string) orm.QuerySeter               { return m }
func (m *mockOfertaQuerySeter) RelatedSel(...interface{}) orm.QuerySeter           { return m }
func (m *mockOfertaQuerySeter) Count() (int64, error)                              { return 0, nil }

// Métodos con contexto
func (m *mockOfertaQuerySeter) OneWithCtx(context.Context, interface{}, ...string) error {
	return m.One(nil)
}
func (m *mockOfertaQuerySeter) AllWithCtx(ctx context.Context, container interface{}, cols ...string) (int64, error) {
	return m.All(container, cols...)
}
func (m *mockOfertaQuerySeter) ValuesWithCtx(context.Context, *[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) ValuesListWithCtx(context.Context, *[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) ValuesFlatWithCtx(context.Context, *orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) CountWithCtx(ctx context.Context) (int64, error) {
	return m.Count()
}
func (m *mockOfertaQuerySeter) ExistWithCtx(context.Context) bool { return false }
func (m *mockOfertaQuerySeter) UpdateWithCtx(context.Context, orm.Params) (int64, error) {
	return 0, nil
}
func (m *mockOfertaQuerySeter) DeleteWithCtx(context.Context) (int64, error) { return 0, nil }
func (m *mockOfertaQuerySeter) PrepareInsertWithCtx(context.Context) (orm.Inserter, error) {
	return nil, nil
}

// Tests para ValidarReglasNegocioOferta - Casos de horarios

func TestOfertaService_ValidarReglasNegocioOferta_HorarioIncompletoSoloInicio(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		HoraInicio:      &horaInicio,
		HoraFin:         nil, // Sin hora fin
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "horario")
}

func TestOfertaService_ValidarReglasNegocioOferta_HorarioIncompletoSoloFin(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaFin := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		HoraInicio:      nil,
		HoraFin:         &horaFin, // Sin hora inicio
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "horario")
}

func TestOfertaService_ValidarReglasNegocioOferta_HorarioFinAntes(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC) // Antes que inicio

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		HoraInicio:      &horaInicio,
		HoraFin:         &horaFin,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hora")
}

func TestOfertaService_ValidarReglasNegocioOferta_HorarioIgual(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC) // Igual

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		HoraInicio:      &horaInicio,
		HoraFin:         &horaFin,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hora")
}

func TestOfertaService_ValidarReglasNegocioOferta_HorarioValido(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		HoraInicio:      &horaInicio,
		HoraFin:         &horaFin,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.NoError(t, err)
}

// Tests para ValidarReglasNegocioOferta - Días de la semana

func TestOfertaService_ValidarReglasNegocioOferta_DiaInvalido(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		DiasSemana:      "Lunes,Martes,InvalidDay",
		DiasSemanaArray: []string{"Lunes", "Martes", "InvalidDay"},
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "día de la semana inválido")
}

func TestOfertaService_ValidarReglasNegocioOferta_DiasValidos(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
		DiasSemana:      "Lunes,Martes,Miércoles",
		DiasSemanaArray: []string{"Lunes", "Martes", "Miércoles"},
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.NoError(t, err)
}

// Tests para obtenerDiaSemanaEspanol - Caso default

func TestOfertaService_ObtenerDiaSemanaEspanol_Default(t *testing.T) {
	service := &OfertaService{}

	// Usar un valor inválido de Weekday (fuera de rango 0-6)
	result := service.obtenerDiaSemanaEspanol(time.Weekday(99))
	assert.Equal(t, "", result)
}

// Tests para CalcularDescuentoOferta

func TestOfertaService_CalcularDescuentoOferta_ErrorObtenerProductos(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							return 0, assert.AnError
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 20,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.Error(t, err)
	assert.Equal(t, int64(0), descuento)
	assert.Contains(t, err.Error(), "error al obtener productos de la oferta")
}

func TestOfertaService_CalcularDescuentoOferta_Porcentaje(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 20, // 20%
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2}, // 20000 total
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(4000), descuento) // 20% de 20000
}

func TestOfertaService_CalcularDescuentoOferta_Monto(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoMonto,
		ValorDescuento: 5000,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2}, // 20000 total
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(5000), descuento)
}

func TestOfertaService_CalcularDescuentoOferta_MontoMayorQueTotal(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoMonto,
		ValorDescuento: 25000, // Mayor que el total
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 5000, Cantidad: 2}, // 10000 total
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), descuento) // No puede ser mayor que el total
}

func TestOfertaService_CalcularDescuentoOferta_TipoInvalido(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  "INVALIDO",
		ValorDescuento: 10,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), descuento)
}

func TestOfertaService_CalcularDescuentoOferta_ProductoNoEnOferta(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 999} // Otro producto
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 20,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2}, // Producto no en oferta
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), descuento) // Sin productos aplicables, descuento = 0
}

func TestOfertaService_CalcularDescuentoOferta_SinProductosEnOferta(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							// Sin productos en oferta
							return 0, nil
						},
					}
				},
			}
		},
	})

	oferta := &models.Oferta{
		PkIdOferta:     1,
		TipoDescuento:  models.TipoDescuentoPorcentaje,
		ValorDescuento: 20,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento, err := service.CalcularDescuentoOferta(oferta, items)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), descuento)
}

// Tests para ObtenerOfertasActivas

func TestOfertaService_ObtenerOfertasActivas_ErrorObtenerOfertas(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							return 0, assert.AnError
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, ofertas)
	assert.Contains(t, err.Error(), "error al obtener ofertas")
}

func TestOfertaService_ObtenerOfertasActivas_SinOfertas(t *testing.T) {
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							// Sin ofertas
							return 0, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, nil)
	assert.NoError(t, err)
	assert.Empty(t, ofertas)
}

func TestOfertaService_ObtenerOfertasActivas_OfertaSimple(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta 1",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemanaArray: []string{},
										},
									}
								}
								return 1, nil
							},
						}
					},
				}
			}
			// Segunda llamada: obtener productos
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, ofertas, 1)
	assert.Equal(t, int64(1), ofertas[0].OfertaId)
	assert.Equal(t, "Oferta 1", ofertas[0].Titulo)
	assert.Len(t, ofertas[0].ProductosIds, 1)
	assert.Equal(t, int64(1), ofertas[0].ProductosIds[0])
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorDiaSemana(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	// Fecha de consulta: Lunes
	fechaConsulta := time.Date(2025, 10, 6, 0, 0, 0, 0, time.UTC) // Lunes

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas (2 ofertas: una válida para Lunes, otra para Martes)
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta Lunes",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemana:      "Lunes",
											DiasSemanaArray: []string{"Lunes"},
										},
										{
											PkIdOferta:      2,
											Titulo:          "Oferta Martes",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  15,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemana:      "Martes",
											DiasSemanaArray: []string{"Martes"},
										},
									}
								}
								return 2, nil
							},
						}
					},
				}
			}
			// Segunda llamada: obtener productos (para oferta 1)
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, &fechaConsulta, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, ofertas, 1)
	assert.Equal(t, "Oferta Lunes", ofertas[0].Titulo)
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorHorario(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	// Hora de consulta dentro del rango (15:00)
	horaConsultaValida := time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas con horario
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta con horario",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											HoraInicio:      &horaInicio,
											HoraFin:         &horaFin,
											DiasSemanaArray: []string{},
										},
									}
								}
								return 1, nil
							},
						}
					},
				}
			}
			// Segunda llamada: obtener productos
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, &horaConsultaValida, nil)
	assert.NoError(t, err)
	assert.Len(t, ofertas, 1)
	assert.Equal(t, "Oferta con horario", ofertas[0].Titulo)
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorHorarioFueraDerango(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	// Hora de consulta fuera del rango (antes: 08:00)
	horaConsultaInvalida := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)

	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertas, ok := container.(*[]*models.Oferta); ok {
								*ofertas = []*models.Oferta{
									{
										PkIdOferta:      1,
										Titulo:          "Oferta con horario",
										TipoDescuento:   models.TipoDescuentoPorcentaje,
										ValorDescuento:  20,
										FechaInicio:     fechaInicio,
										FechaFin:        fechaFin,
										PkIdRestaurante: restaurante,
										HoraInicio:      &horaInicio,
										HoraFin:         &horaFin,
										DiasSemanaArray: []string{},
									},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, &horaConsultaInvalida, nil)
	assert.NoError(t, err)
	assert.Empty(t, ofertas) // Filtrada por horario
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorHorarioDespues(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	horaInicio := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	horaFin := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	// Hora de consulta fuera del rango (después: 22:00)
	horaConsultaInvalida := time.Date(2025, 1, 1, 22, 0, 0, 0, time.UTC)

	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertas, ok := container.(*[]*models.Oferta); ok {
								*ofertas = []*models.Oferta{
									{
										PkIdOferta:      1,
										Titulo:          "Oferta con horario",
										TipoDescuento:   models.TipoDescuentoPorcentaje,
										ValorDescuento:  20,
										FechaInicio:     fechaInicio,
										FechaFin:        fechaFin,
										PkIdRestaurante: restaurante,
										HoraInicio:      &horaInicio,
										HoraFin:         &horaFin,
										DiasSemanaArray: []string{},
									},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, &horaConsultaInvalida, nil)
	assert.NoError(t, err)
	assert.Empty(t, ofertas) // Filtrada por horario
}

func TestOfertaService_ObtenerOfertasActivas_ErrorObtenerProductos(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta 1",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemanaArray: []string{},
										},
									}
								}
								return 1, nil
							},
						}
					},
				}
			}
			// Segunda llamada: error al obtener productos
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							return 0, assert.AnError
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, ofertas)
	assert.Contains(t, err.Error(), "error al obtener productos de la oferta")
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorProducto(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	productoIdFiltro := int64(1)

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta 1",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemanaArray: []string{},
										},
									}
								}
								return 1, nil
							},
						}
					},
				}
			}
			// Segunda llamada: obtener productos (incluye el producto filtrado)
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, &productoIdFiltro)
	assert.NoError(t, err)
	assert.Len(t, ofertas, 1)
	assert.Equal(t, "Oferta 1", ofertas[0].Titulo)
}

func TestOfertaService_ObtenerOfertasActivas_FiltradoPorProductoNoEnOferta(t *testing.T) {
	fechaInicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	fechaFin := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}
	productoIdFiltro := int64(999) // Producto no en oferta

	callCount := 0
	service := NewOfertaService(&mockOfertaOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			callCount++
			if callCount == 1 {
				// Primera llamada: obtener ofertas
				return &mockOfertaQuerySeter{
					filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
						return &mockOfertaQuerySeter{
							allFn: func(container interface{}, cols ...string) (int64, error) {
								if ofertas, ok := container.(*[]*models.Oferta); ok {
									*ofertas = []*models.Oferta{
										{
											PkIdOferta:      1,
											Titulo:          "Oferta 1",
											TipoDescuento:   models.TipoDescuentoPorcentaje,
											ValorDescuento:  20,
											FechaInicio:     fechaInicio,
											FechaFin:        fechaFin,
											PkIdRestaurante: restaurante,
											DiasSemanaArray: []string{},
										},
									}
								}
								return 1, nil
							},
						}
					},
				}
			}
			// Segunda llamada: obtener productos (producto diferente)
			return &mockOfertaQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockOfertaQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if ofertaProds, ok := container.(*[]*models.OfertaProducto); ok {
								producto := &models.Producto{PK_ID_PRODUCTO: 1}
								*ofertaProds = []*models.OfertaProducto{
									{PkIdProducto: producto},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	ofertas, err := service.ObtenerOfertasActivas(context.Background(), 1, nil, nil, &productoIdFiltro)
	assert.NoError(t, err)
	assert.Empty(t, ofertas) // Filtrada porque producto no está en oferta
}
