package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type cuponQuerySeter interface {
	Filter(string, ...interface{}) cuponQuerySeter
	One(interface{}, ...string) error
	Count() (int64, error)
	RelatedSel(...interface{}) cuponQuerySeter
	All(interface{}, ...string) (int64, error)
}

// Errores tipados de CuponService; el controlador los traduce a códigos HTTP.
var (
	ErrCuponNoEncontrado   = errors.New("cupón no encontrado")
	ErrClienteNoEncontrado = errors.New("cliente no encontrado")
	ErrPedidoNoEncontrado  = errors.New("pedido no encontrado")
	ErrPedidoRequerido     = errors.New("pedidoId es requerido")
	// ErrCuponNoAplicable: el cupón no cumple las reglas (inactivo, fuera de vigencia, monto mínimo, cliente, productos).
	ErrCuponNoAplicable = errors.New("cupón no aplicable")
	// ErrCuponConflicto: el cupón ya fue redimido o está agotado (máximo de usos, límite por cliente, ya usado en el pedido).
	ErrCuponConflicto = errors.New("cupón no redimible")
)

type cuponOrmer interface {
	QueryTable(string) cuponQuerySeter
	Insert(interface{}) (int64, error)
}

type CuponService struct {
	ormer cuponOrmer
}

func NewCuponService(ormer cuponOrmer) *CuponService {
	return &CuponService{ormer: ormer}
}

// NewCuponServiceFromOrm crea el servicio sobre un orm.Ormer real.
func NewCuponServiceFromOrm(o orm.Ormer) *CuponService {
	return NewCuponService(NewCuponOrmerFromFuncs(func(name string) orm.QuerySeter { return o.QueryTable(name) }, o.Insert))
}

func NewCuponOrmerFromFuncs(queryTable func(string) orm.QuerySeter, insertFunc func(interface{}) (int64, error)) cuponOrmer {
	if queryTable == nil || insertFunc == nil {
		return nil
	}
	return beegoCuponOrmer{
		queryTable: func(name string) cuponQuerySeter {
			return wrapOrmQuerySeter(queryTable(name))
		},
		insertFunc: insertFunc,
	}
}

type beegoCuponOrmer struct {
	queryTable func(string) cuponQuerySeter
	insertFunc func(interface{}) (int64, error)
}

func (o beegoCuponOrmer) QueryTable(name string) cuponQuerySeter {
	if o.queryTable == nil {
		return beegoCuponQuerySeter{}
	}
	return o.queryTable(name)
}

type beegoCuponQuerySeter struct {
	filterFunc     func(string, ...interface{}) cuponQuerySeter
	oneFunc        func(interface{}, ...string) error
	countFunc      func() (int64, error)
	relatedSelFunc func(...interface{}) cuponQuerySeter
	allFunc        func(interface{}, ...string) (int64, error)
}

func (q beegoCuponQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	if q.allFunc == nil {
		return 0, nil
	}
	return q.allFunc(container, cols...)
}

func (q beegoCuponQuerySeter) Filter(field string, args ...interface{}) cuponQuerySeter {
	if q.filterFunc == nil {
		return q
	}
	return q.filterFunc(field, args...)
}

func (q beegoCuponQuerySeter) One(container interface{}, cols ...string) error {
	if q.oneFunc == nil {
		return orm.ErrNoRows
	}
	return q.oneFunc(container, cols...)
}

func (q beegoCuponQuerySeter) Count() (int64, error) {
	if q.countFunc == nil {
		return 0, nil
	}
	return q.countFunc()
}

func (q beegoCuponQuerySeter) RelatedSel(params ...interface{}) cuponQuerySeter {
	if q.relatedSelFunc == nil {
		return q
	}
	return q.relatedSelFunc(params...)
}

func wrapOrmQuerySeter(qs orm.QuerySeter) cuponQuerySeter {
	if qs == nil {
		return beegoCuponQuerySeter{}
	}
	return beegoCuponQuerySeter{
		filterFunc: func(field string, args ...interface{}) cuponQuerySeter {
			return wrapOrmQuerySeter(qs.Filter(field, args...))
		},
		oneFunc: func(container interface{}, cols ...string) error {
			return qs.One(container, cols...)
		},
		countFunc: func() (int64, error) {
			return qs.Count()
		},
		relatedSelFunc: func(params ...interface{}) cuponQuerySeter {
			return wrapOrmQuerySeter(qs.RelatedSel(params...))
		},
		allFunc: func(container interface{}, cols ...string) (int64, error) {
			return qs.All(container, cols...)
		},
	}
}

func (o beegoCuponOrmer) Insert(model interface{}) (int64, error) {
	if o.insertFunc == nil {
		return 0, fmt.Errorf("insert function no configurada")
	}
	return o.insertFunc(model)
}

// ahoraBogota devuelve la hora actual en zona America/Bogota.
func ahoraBogota() time.Time {
	loc := database.BogotaZone
	if loc == nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}
	return time.Now().In(loc)
}

// soloDia descarta la hora de t (usando su fecha en UTC, como FormatDateUTC).
func soloDia(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ValidarCupon evalúa si un cupón es aplicable a la solicitud. Las reglas
// incumplidas se informan con Aplicable=false y un Motivo (no es error).
func (s *CuponService) ValidarCupon(ctx context.Context, req *models.ValidarCuponRequest) (*models.ValidarCuponResponse, error) {
	resp, _, err := s.validar(ctx, req)
	return resp, err
}

// validar implementa ValidarCupon. El segundo resultado indica que el rechazo
// es un conflicto (usos agotados), no una regla de aplicabilidad.
func (s *CuponService) validar(_ context.Context, req *models.ValidarCuponRequest) (*models.ValidarCuponResponse, bool, error) {
	if s.ormer == nil {
		return nil, false, fmt.Errorf("ormer no configurado")
	}

	rechazo := func(motivo string, conflicto bool) (*models.ValidarCuponResponse, bool, error) {
		return &models.ValidarCuponResponse{Aplicable: false, Motivo: stringPtr(motivo)}, conflicto, nil
	}

	cupon := &models.Cupon{}
	err := s.ormer.QueryTable("cupon").Filter("codigo", req.Codigo).One(cupon)
	if err != nil {
		if err == orm.ErrNoRows {
			return rechazo("Cupón no encontrado", false)
		}
		return nil, false, fmt.Errorf("error al buscar cupón: %w", err)
	}

	if !cupon.Activo {
		return rechazo("Cupón inactivo", false)
	}

	ahora := ahoraBogota()
	hoy := time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 0, 0, 0, 0, time.UTC)
	if hoy.Before(soloDia(cupon.FechaInicio)) || hoy.After(soloDia(cupon.FechaFin)) {
		return rechazo("Cupón fuera del período de vigencia", false)
	}

	if cupon.MaxUsos != nil {
		usosActuales, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Count()
		if err != nil {
			return nil, false, fmt.Errorf("error al contar usos del cupón: %w", err)
		}
		if usosActuales >= int64(*cupon.MaxUsos) {
			return rechazo("Cupón ha alcanzado el límite máximo de usos", true)
		}
	}

	if cupon.LimitePorCliente != nil {
		usosCliente, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Filter("pk_documento_cliente", req.ClienteId).Count()
		if err != nil {
			return nil, false, fmt.Errorf("error al contar usos del cliente: %w", err)
		}
		if usosCliente >= int64(*cupon.LimitePorCliente) {
			return rechazo("Cliente ha alcanzado el límite de usos para este cupón", true)
		}
	}

	if cupon.Scope == models.CuponScopeCliente && cupon.PkDocumentoCliente != nil {
		if cupon.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE != req.ClienteId {
			return rechazo("Cupón no válido para este cliente", false)
		}
	}

	montoTotal := int64(0)
	productosAplicables := []int64{}
	for _, item := range req.Items {
		montoTotal += item.Precio * int64(item.Cantidad)

		if s.esProductoAplicable(cupon, item.ProductoId) {
			productosAplicables = append(productosAplicables, item.ProductoId)
		}
	}

	if cupon.MontoMinimo != nil && montoTotal < *cupon.MontoMinimo {
		return rechazo(fmt.Sprintf("El monto mínimo requerido es %d", *cupon.MontoMinimo), false)
	}

	if (cupon.Scope == models.CuponScopeProducto || cupon.Scope == models.CuponScopeCategoria) && len(productosAplicables) == 0 {
		return rechazo("No hay productos aplicables para este cupón", false)
	}

	montoDescuento := s.calcularDescuento(cupon, montoTotal, req.Items, productosAplicables)

	return &models.ValidarCuponResponse{
		Aplicable:      true,
		MontoDescuento: montoDescuento,
	}, false, nil
}

// itemsDePedido arma los ítems (producto, cantidad, precio) del pedido a partir de detalle_pedido.
func (s *CuponService) itemsDePedido(pedidoId int64) ([]models.ValidarCuponItemRequest, error) {
	var detalles []*models.DetallePedido
	if _, err := s.ormer.QueryTable("detalle_pedido").Filter("pk_id_pedido", pedidoId).All(&detalles); err != nil {
		return nil, fmt.Errorf("error al obtener el detalle del pedido: %w", err)
	}
	items := make([]models.ValidarCuponItemRequest, 0, len(detalles))
	for _, d := range detalles {
		if d.PKIDProducto == nil {
			continue
		}
		items = append(items, models.ValidarCuponItemRequest{
			ProductoId: d.PKIDProducto.PK_ID_PRODUCTO,
			Cantidad:   d.Cantidad,
			Precio:     d.Precio,
		})
	}
	return items, nil
}

// existe indica si hay al menos una fila de tabla con campo = valor.
func (s *CuponService) existe(tabla, campo string, valor int64) (bool, error) {
	n, err := s.ormer.QueryTable(tabla).Filter(campo, valor).Count()
	return n > 0, err
}

// RedimirCupon registra la redención de un cupón para un cliente y un pedido
// existentes. Los montos se calculan con el detalle del pedido. Errores:
// ErrPedidoRequerido, ErrCuponNoEncontrado, ErrClienteNoEncontrado,
// ErrPedidoNoEncontrado, ErrCuponConflicto (agotado / ya redimido) y
// ErrCuponNoAplicable (reglas incumplidas); cualquier otro es un error interno.
func (s *CuponService) RedimirCupon(ctx context.Context, codigo string, req *models.RedimirCuponRequest) (*models.CuponRedencion, error) {
	if s.ormer == nil {
		return nil, fmt.Errorf("ormer no configurado")
	}
	if req.PedidoId == nil {
		return nil, ErrPedidoRequerido
	}

	cupon := &models.Cupon{}
	if err := s.ormer.QueryTable("cupon").Filter("codigo", codigo).One(cupon); err != nil {
		if err == orm.ErrNoRows {
			return nil, ErrCuponNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar cupón: %w", err)
	}

	ok, err := s.existe("cliente", "pk_documento_cliente", req.ClienteId)
	if err != nil {
		return nil, fmt.Errorf("error al buscar cliente: %w", err)
	}
	if !ok {
		return nil, ErrClienteNoEncontrado
	}

	ok, err = s.existe("pedido", "pk_id_pedido", *req.PedidoId)
	if err != nil {
		return nil, fmt.Errorf("error al buscar pedido: %w", err)
	}
	if !ok {
		return nil, ErrPedidoNoEncontrado
	}

	yaRedimido, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Filter("pk_id_pedido", *req.PedidoId).Count()
	if err != nil {
		return nil, fmt.Errorf("error al verificar redenciones del pedido: %w", err)
	}
	if yaRedimido > 0 {
		return nil, fmt.Errorf("%w: el cupón ya fue redimido para este pedido", ErrCuponConflicto)
	}

	items, err := s.itemsDePedido(*req.PedidoId)
	if err != nil {
		return nil, err
	}

	validacion, conflicto, err := s.validar(ctx, &models.ValidarCuponRequest{
		PedidoId:  req.PedidoId,
		ClienteId: req.ClienteId,
		Codigo:    codigo,
		Items:     items,
	})
	if err != nil {
		return nil, fmt.Errorf("error al validar cupón: %w", err)
	}
	if !validacion.Aplicable {
		if conflicto {
			return nil, fmt.Errorf("%w: %s", ErrCuponConflicto, *validacion.Motivo)
		}
		return nil, fmt.Errorf("%w: %s", ErrCuponNoAplicable, *validacion.Motivo)
	}

	redencion := &models.CuponRedencion{
		PkIdCupon:          cupon,
		PkDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: req.ClienteId},
		PkIdPedido:         &models.Pedido{PK_ID_PEDIDO: *req.PedidoId},
		MontoDescuento:     validacion.MontoDescuento,
	}

	if _, err = s.ormer.Insert(redencion); err != nil {
		return nil, fmt.Errorf("error al registrar redención: %w", err)
	}

	return redencion, nil
}

func (s *CuponService) esProductoAplicable(cupon *models.Cupon, productoId int64) bool {
	switch cupon.Scope {
	case models.CuponScopeGlobal:
		return true
	case models.CuponScopeProducto:
		return cupon.PkIdProducto != nil && cupon.PkIdProducto.PK_ID_PRODUCTO == productoId
	case models.CuponScopeCategoria:
		if s.ormer == nil {
			return false
		}
		if cupon.PkIdCategoria == nil {
			return false
		}

		producto := &models.Producto{}
		err := s.ormer.QueryTable("producto").Filter("pk_id_producto", productoId).RelatedSel().One(producto)
		if err != nil {
			return false
		}
		if producto.PK_ID_SUBCATEGORIA == nil {
			return false
		}

		subcategoria := &models.Subcategoria{}
		err = s.ormer.QueryTable("subcategoria").Filter("pk_id_subcategoria", producto.PK_ID_SUBCATEGORIA.PK_ID_SUBCATEGORIA).RelatedSel().One(subcategoria)
		if err != nil {
			return false
		}

		if subcategoria.PK_ID_CATEGORIA == nil {
			return false
		}

		return subcategoria.PK_ID_CATEGORIA.PK_ID_CATEGORIA == cupon.PkIdCategoria.PK_ID_CATEGORIA
	case models.CuponScopeCliente:
		return true
	}
	return false
}

func (s *CuponService) calcularDescuento(cupon *models.Cupon, montoTotal int64, items []models.ValidarCuponItemRequest, productosAplicables []int64) int64 {
	var montoAplicable int64

	if cupon.Scope == models.CuponScopeGlobal || cupon.Scope == models.CuponScopeCliente {
		montoAplicable = montoTotal
	} else {

		for _, item := range items {
			for _, prodId := range productosAplicables {
				if item.ProductoId == prodId {
					montoAplicable += item.Precio * int64(item.Cantidad)
					break
				}
			}
		}
	}

	switch cupon.TipoDescuento {
	case models.TipoDescuentoPorcentaje:
		return (montoAplicable * cupon.ValorDescuento) / 100
	case models.TipoDescuentoMonto:
		if cupon.ValorDescuento > montoAplicable {
			return montoAplicable
		}
		return cupon.ValorDescuento
	}
	return 0
}

func (s *CuponService) ValidarReglasNegocioCupon(cupon *models.Cupon) error {

	switch cupon.TipoDescuento {
	case models.TipoDescuentoPorcentaje:
		if cupon.ValorDescuento < 1 || cupon.ValorDescuento > 100 {
			return fmt.Errorf("el porcentaje de descuento debe estar entre 1 y 100")
		}
	case models.TipoDescuentoMonto:
		if cupon.ValorDescuento < 0 {
			return fmt.Errorf("el monto de descuento debe ser mayor o igual a 0")
		}
	}

	if cupon.FechaFin.Before(cupon.FechaInicio) {
		return fmt.Errorf("la fecha de fin debe ser posterior a la fecha de inicio")
	}

	if cupon.MaxUsos != nil && *cupon.MaxUsos < 0 {
		return fmt.Errorf("maxUsos no puede ser negativo")
	}
	if cupon.LimitePorCliente != nil && *cupon.LimitePorCliente < 0 {
		return fmt.Errorf("limitePorCliente no puede ser negativo")
	}
	if cupon.MontoMinimo != nil && *cupon.MontoMinimo < 0 {
		return fmt.Errorf("montoMinimo no puede ser negativo")
	}

	switch cupon.Scope {
	case models.CuponScopeProducto:
		if cupon.PkIdProducto == nil {
			return fmt.Errorf("debe especificar un producto para cupones con scope PRODUCTO")
		}
		if cupon.PkIdCategoria != nil || cupon.PkDocumentoCliente != nil {
			return fmt.Errorf("no debe especificar categoría o cliente para cupones con scope PRODUCTO")
		}
	case models.CuponScopeCategoria:
		if cupon.PkIdCategoria == nil {
			return fmt.Errorf("debe especificar una categoría para cupones con scope CATEGORIA")
		}
		if cupon.PkIdProducto != nil || cupon.PkDocumentoCliente != nil {
			return fmt.Errorf("no debe especificar producto o cliente para cupones con scope CATEGORIA")
		}
	case models.CuponScopeCliente:
		if cupon.PkDocumentoCliente == nil {
			return fmt.Errorf("debe especificar un cliente para cupones con scope CLIENTE")
		}
		if cupon.PkIdProducto != nil || cupon.PkIdCategoria != nil {
			return fmt.Errorf("no debe especificar producto o categoría para cupones con scope CLIENTE")
		}
	case models.CuponScopeGlobal:
		if cupon.PkIdProducto != nil || cupon.PkIdCategoria != nil || cupon.PkDocumentoCliente != nil {
			return fmt.Errorf("no debe especificar producto, categoría o cliente para cupones con scope GLOBAL")
		}
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}
