package services

import (
	"testing"

	"restaurante/models"

	"github.com/stretchr/testify/assert"
)

type mockCuponOrmer struct {
	tables   map[string]func() cuponQuerySeter
	insertFn func(interface{}) (int64, error)
}

func (m *mockCuponOrmer) QueryTable(name string) cuponQuerySeter {
	if m.tables != nil {
		if factory, ok := m.tables[name]; ok && factory != nil {
			return factory()
		}
	}
	return &mockQuerySeter{}
}

func (m *mockCuponOrmer) Insert(model interface{}) (int64, error) {
	if m.insertFn != nil {
		return m.insertFn(model)
	}
	return 0, nil
}

type mockQuerySeter struct {
	filterHook     func(string, ...interface{}) cuponQuerySeter
	oneHook        func(interface{}, ...string) error
	countHook      func() (int64, error)
	relatedSelHook func(...interface{}) cuponQuerySeter
}

func (m *mockQuerySeter) Filter(field string, args ...interface{}) cuponQuerySeter {
	if m.filterHook != nil {
		return m.filterHook(field, args...)
	}
	return m
}

func (m *mockQuerySeter) One(dest interface{}, cols ...string) error {
	if m.oneHook != nil {
		return m.oneHook(dest, cols...)
	}
	return nil
}

func (m *mockQuerySeter) Count() (int64, error) {
	if m.countHook != nil {
		return m.countHook()
	}
	return 0, nil
}

func (m *mockQuerySeter) RelatedSel(params ...interface{}) cuponQuerySeter {
	if m.relatedSelHook != nil {
		return m.relatedSelHook(params...)
	}
	return m
}

func TestNewCuponService(t *testing.T) {

	service := NewCuponService(nil)
	assert.NotNil(t, service)
}

func TestCuponService_ValidarReglasNegocioCupon_TipoDescuentoPorcentaje(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "PORCENTAJE",
		ValorDescuento: 10,
		Scope:          "GLOBAL",
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.NoError(t, err)

	cupon.ValorDescuento = 150
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "porcentaje")
}

func TestCuponService_ValidarReglasNegocioCupon_TipoDescuentoMonto(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "MONTO",
		ValorDescuento: 5000,
		Scope:          "GLOBAL",
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.NoError(t, err)

	cupon.ValorDescuento = -1000
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monto")
}

func TestCuponService_CalcularDescuento_Porcentaje(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "PORCENTAJE",
		ValorDescuento: 10,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento := service.calcularDescuento(cupon, 20000, items, []int64{1})
	assert.Equal(t, int64(2000), descuento)
}

func TestCuponService_CalcularDescuento_Monto(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "MONTO",
		ValorDescuento: 5000,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento := service.calcularDescuento(cupon, 20000, items, []int64{1})
	assert.Equal(t, int64(5000), descuento)
}

func TestCuponService_CalcularDescuento_MontoMayorQueTotal(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "MONTO",
		ValorDescuento: 25000,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento := service.calcularDescuento(cupon, 20000, items, []int64{1})
	assert.Equal(t, int64(20000), descuento)
}

func TestCuponService_EsProductoAplicable_Global(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		Scope: "GLOBAL",
	}

	aplicable := service.esProductoAplicable(cupon, 123)
	assert.True(t, aplicable)
}

func TestCuponService_EsProductoAplicable_Producto(t *testing.T) {
	service := &CuponService{}

	producto := &models.Producto{PK_ID_PRODUCTO: 123}
	cupon := &models.Cupon{
		Scope:        "PRODUCTO",
		PkIdProducto: producto,
	}

	aplicable := service.esProductoAplicable(cupon, 123)
	assert.True(t, aplicable)

	aplicable = service.esProductoAplicable(cupon, 456)
	assert.False(t, aplicable)
}

func TestCuponService_EsProductoAplicable_Cliente(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		Scope: "CLIENTE",
	}

	aplicable := service.esProductoAplicable(cupon, 123)
	assert.True(t, aplicable)
}

func TestStringPtrFunction(t *testing.T) {
	result := stringPtr("test")
	assert.NotNil(t, result)
	assert.Equal(t, "test", *result)
}

func TestCuponService_ValidarReglasNegocioCupon_ScopeProducto(t *testing.T) {
	service := &CuponService{}

	producto := &models.Producto{PK_ID_PRODUCTO: 123}
	cupon := &models.Cupon{
		TipoDescuento:  "PORCENTAJE",
		ValorDescuento: 10,
		Scope:          "PRODUCTO",
		PkIdProducto:   producto,
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.NoError(t, err)

	cupon.PkIdProducto = nil
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "producto")

	cupon.PkIdProducto = producto
	cupon.PkIdCategoria = &models.Categoria{PK_ID_CATEGORIA: 9}
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PRODUCTO")
}

func TestCuponService_ValidarReglasNegocioCupon_ScopeCategoria(t *testing.T) {
	service := &CuponService{}

	categoria := &models.Categoria{PK_ID_CATEGORIA: 5}
	cupon := &models.Cupon{
		TipoDescuento:  "PORCENTAJE",
		ValorDescuento: 10,
		Scope:          "CATEGORIA",
		PkIdCategoria:  categoria,
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.NoError(t, err)

	cupon.PkIdCategoria = nil
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "categoría")

	cupon.PkIdCategoria = categoria
	cupon.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: 22}
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CATEGORIA")
}

func TestCuponService_ValidarReglasNegocioCupon_ScopeCliente(t *testing.T) {
	service := &CuponService{}

	cliente := &models.Cliente{PK_DOCUMENTO_CLIENTE: 123}
	cupon := &models.Cupon{
		TipoDescuento:      "PORCENTAJE",
		ValorDescuento:     10,
		Scope:              "CLIENTE",
		PkDocumentoCliente: cliente,
	}

	err := service.ValidarReglasNegocioCupon(cupon)
	assert.NoError(t, err)

	cupon.PkDocumentoCliente = nil
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cliente")

	cupon.PkDocumentoCliente = cliente
	cupon.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: 33}
	err = service.ValidarReglasNegocioCupon(cupon)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CLIENTE")
}

func TestCuponService_CalcularDescuento_TipoInvalido(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		TipoDescuento:  "INVALIDO",
		ValorDescuento: 10,
	}

	items := []models.ValidarCuponItemRequest{
		{ProductoId: 1, Precio: 10000, Cantidad: 2},
	}

	descuento := service.calcularDescuento(cupon, 20000, items, []int64{1})
	assert.Equal(t, int64(0), descuento)
}

func TestCuponService_EsProductoAplicable_ScopeInvalido(t *testing.T) {
	service := &CuponService{}

	cupon := &models.Cupon{
		Scope: "INVALIDO",
	}

	aplicable := service.esProductoAplicable(cupon, 123)
	assert.False(t, aplicable)
}
