package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"restaurante/database"
	"restaurante/internal/dberr"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type cuponQuerySeter interface {
	Filter(string, ...interface{}) cuponQuerySeter
	One(interface{}, ...string) error
	Count() (int64, error)
	RelatedSel(...interface{}) cuponQuerySeter
	All(interface{}, ...string) (int64, error)
	// ForUpdate agrega SELECT ... FOR UPDATE; solo bloquea dentro de una transacción.
	ForUpdate() cuponQuerySeter
}

// Errores tipados de CuponService; el controlador los traduce a códigos HTTP.
var (
	ErrCuponNoEncontrado  = errors.New("cupón no encontrado")
	ErrPedidoNoEncontrado = errors.New("pedido no encontrado")
	ErrPedidoRequerido    = errors.New("pedidoId es requerido")
	// ErrPedidoAjeno: el pedido no pertenece al cliente que actúa.
	ErrPedidoAjeno = errors.New("el pedido no pertenece al cliente")
	// ErrPedidoCerrado: el pedido está cancelado o terminado y ya no admite descuentos.
	ErrPedidoCerrado = errors.New("el pedido está cancelado o terminado")
	// ErrCuponNoAplicable: el cupón no cumple las reglas (inactivo, fuera de vigencia, monto mínimo, cliente, productos).
	ErrCuponNoAplicable = errors.New("cupón no aplicable")
	// ErrCuponConflicto: el cupón ya fue redimido o está agotado (máximo de usos, límite por cliente, ya usado en el pedido).
	ErrCuponConflicto = errors.New("cupón no redimible")
)

type cuponOrmer interface {
	QueryTable(string) cuponQuerySeter
	Insert(interface{}) (int64, error)
}

// CuponService valida y redime cupones. La redención se ejecuta dentro de una
// transacción que bloquea (FOR UPDATE) el pedido y la fila del cupón, de modo
// que los topes (maxUsos, limitePorCliente) se comprueban bajo el bloqueo.
type CuponService struct {
	ormer cuponOrmer
	// runTx ejecuta fn dentro de una transacción con un servicio ligado a ella;
	// nil ejecuta fn sin transacción sobre el mismo ormer (pruebas).
	runTx func(ctx context.Context, fn func(*CuponService) error) error
}

func NewCuponService(ormer cuponOrmer) *CuponService {
	return &CuponService{ormer: ormer}
}

// NewCuponServiceFromOrm crea el servicio sobre un orm.Ormer real; RedimirCupon
// abre una transacción con DoTx.
func NewCuponServiceFromOrm(o orm.Ormer) *CuponService {
	return &CuponService{
		ormer: ormerAdapter{queryTable: func(name string) orm.QuerySeter { return o.QueryTable(name) }, insert: o.Insert},
		runTx: func(ctx context.Context, fn func(*CuponService) error) error {
			return doTx(ctx, o, func(tx orm.TxOrmer) error { return fn(newCuponServiceFromTx(tx)) })
		},
	}
}

// doTx abre una transacción, ejecuta task y la confirma. A diferencia de
// orm.DoTx (que solo registra en el log un COMMIT fallido y devuelve nil),
// devuelve el error del COMMIT: un descuento o redención no confirmados nunca se
// reportan como exitosos. Ante error o pánico en task hace ROLLBACK.
func doTx(ctx context.Context, o orm.TxBeginner, task func(orm.TxOrmer) error) error {
	tx, err := o.BeginWithCtx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err = task(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// newCuponServiceFromTx liga un servicio a una transacción abierta.
func newCuponServiceFromTx(tx orm.TxOrmer) *CuponService {
	return &CuponService{ormer: ormerAdapter{queryTable: func(name string) orm.QuerySeter { return tx.QueryTable(name) }, insert: tx.Insert}}
}

// inTx ejecuta fn en una transacción (o directamente si no hay runTx).
func (s *CuponService) inTx(ctx context.Context, fn func(*CuponService) error) error {
	if s.runTx == nil {
		return fn(s)
	}
	return s.runTx(ctx, fn)
}

// ormerAdapter adapta un orm.Ormer / orm.TxOrmer a cuponOrmer.
type ormerAdapter struct {
	queryTable func(string) orm.QuerySeter
	insert     func(interface{}) (int64, error)
}

func (a ormerAdapter) QueryTable(name string) cuponQuerySeter { return qsAdapter{a.queryTable(name)} }
func (a ormerAdapter) Insert(m interface{}) (int64, error)    { return a.insert(m) }

// qsAdapter adapta orm.QuerySeter a cuponQuerySeter.
type qsAdapter struct{ qs orm.QuerySeter }

func (q qsAdapter) Filter(field string, args ...interface{}) cuponQuerySeter {
	return qsAdapter{q.qs.Filter(field, args...)}
}
func (q qsAdapter) One(container interface{}, cols ...string) error {
	return q.qs.One(container, cols...)
}
func (q qsAdapter) Count() (int64, error) { return q.qs.Count() }
func (q qsAdapter) RelatedSel(params ...interface{}) cuponQuerySeter {
	return qsAdapter{q.qs.RelatedSel(params...)}
}
func (q qsAdapter) All(container interface{}, cols ...string) (int64, error) {
	return q.qs.All(container, cols...)
}
func (q qsAdapter) ForUpdate() cuponQuerySeter { return qsAdapter{q.qs.ForUpdate()} }

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
// incumplidas se informan con Aplicable=false y un Motivo (no es error). Con
// PedidoId los ítems salen del detalle real del pedido (que debe pertenecer al
// cliente: ErrPedidoNoEncontrado / ErrPedidoAjeno) y se ignora req.Items.
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

	items := req.Items
	if req.PedidoId != nil {
		pedido, err := s.cargarPedido(*req.PedidoId, false)
		if err != nil {
			return nil, false, err
		}
		if err = verificarPropietario(pedido, req.ClienteId); err != nil {
			return nil, false, err
		}
		if items, err = s.itemsDePedido(pedido.PK_ID_PEDIDO); err != nil {
			return nil, false, err
		}
	}

	cupon := &models.Cupon{}
	err := s.ormer.QueryTable("cupon").Filter("codigo", req.Codigo).One(cupon)
	if err != nil {
		if err == orm.ErrNoRows {
			return &models.ValidarCuponResponse{Aplicable: false, Motivo: stringPtr("Cupón no encontrado")}, false, nil
		}
		return nil, false, fmt.Errorf("error al buscar cupón: %w", err)
	}
	return s.evaluar(cupon, req.ClienteId, items)
}

// evaluar aplica las reglas del cupón (estado, vigencia, topes de uso, cliente,
// monto mínimo y productos) y calcula el descuento. Los conteos de uso se leen
// con el ormer del servicio: dentro de RedimirCupon es el de la transacción con
// la fila del cupón ya bloqueada.
func (s *CuponService) evaluar(cupon *models.Cupon, clienteID int64, items []models.ValidarCuponItemRequest) (*models.ValidarCuponResponse, bool, error) {
	rechazo := func(motivo string, conflicto bool) (*models.ValidarCuponResponse, bool, error) {
		return &models.ValidarCuponResponse{Aplicable: false, Motivo: stringPtr(motivo)}, conflicto, nil
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
		usosCliente, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Filter("pk_documento_cliente", clienteID).Count()
		if err != nil {
			return nil, false, fmt.Errorf("error al contar usos del cliente: %w", err)
		}
		if usosCliente >= int64(*cupon.LimitePorCliente) {
			return rechazo("Cliente ha alcanzado el límite de usos para este cupón", true)
		}
	}

	if cupon.Scope == models.CuponScopeCliente && cupon.PkDocumentoCliente != nil {
		if cupon.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE != clienteID {
			return rechazo("Cupón no válido para este cliente", false)
		}
	}

	montoTotal := int64(0)
	productosAplicables := []int64{}
	for _, item := range items {
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

	montoDescuento := s.calcularDescuento(cupon, montoTotal, items, productosAplicables)

	return &models.ValidarCuponResponse{
		Aplicable:      true,
		MontoDescuento: montoDescuento,
	}, false, nil
}

// cargarPedido lee el pedido (sin relaciones); con lock lo bloquea con
// SELECT ... FOR UPDATE (solo tiene efecto dentro de una transacción).
func (s *CuponService) cargarPedido(pedidoID int64, lock bool) (*models.Pedido, error) {
	qs := s.ormer.QueryTable("pedido").Filter("pk_id_pedido", pedidoID)
	if lock {
		qs = qs.ForUpdate()
	}
	pedido := &models.Pedido{}
	if err := qs.One(pedido); err != nil {
		if err == orm.ErrNoRows {
			return nil, ErrPedidoNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar pedido: %w", err)
	}
	return pedido, nil
}

// verificarPropietario exige que el pedido pertenezca al cliente.
func verificarPropietario(pedido *models.Pedido, clienteID int64) error {
	if pedido.PK_DOCUMENTO_CLIENTE == nil || pedido.PK_DOCUMENTO_CLIENTE.PK_DOCUMENTO_CLIENTE != clienteID {
		return ErrPedidoAjeno
	}
	return nil
}

// bloquearPedido bloquea el pedido (FOR UPDATE) y exige que sea del cliente y
// que siga abierto (ni TERMINADO ni CANCELADO). Debe llamarse dentro de la
// transacción, antes de bloquear el cupón (orden único de bloqueo: pedido, cupón).
func (s *CuponService) bloquearPedido(pedidoID, clienteID int64) (*models.Pedido, error) {
	pedido, err := s.cargarPedido(pedidoID, true)
	if err != nil {
		return nil, err
	}
	if err = verificarPropietario(pedido, clienteID); err != nil {
		return nil, err
	}
	if pedido.ESTADO_PEDIDO == models.EstadoPedidoTerminado || pedido.ESTADO_PEDIDO == models.EstadoPedidoCancelado {
		return nil, ErrPedidoCerrado
	}
	return pedido, nil
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

// RedimirCupon registra la redención de un cupón para un cliente y un pedido
// suyo. Todo ocurre en una transacción: se bloquea el pedido y luego la fila del
// cupón (SELECT ... FOR UPDATE), se vuelven a comprobar activo, vigencia,
// maxUsos y limitePorCliente con conteos bajo el bloqueo, se inserta la
// redención y se confirma; dos redenciones simultáneas no pueden superar los
// topes. Los montos se calculan con el detalle del pedido. Errores:
// ErrPedidoRequerido, ErrPedidoNoEncontrado, ErrPedidoAjeno, ErrPedidoCerrado,
// ErrCuponNoEncontrado, ErrCuponConflicto (agotado / ya redimido) y
// ErrCuponNoAplicable (reglas incumplidas); cualquier otro es un error interno.
func (s *CuponService) RedimirCupon(ctx context.Context, codigo string, req *models.RedimirCuponRequest) (*models.CuponRedencion, error) {
	if s.ormer == nil {
		return nil, fmt.Errorf("ormer no configurado")
	}
	if req.PedidoId == nil {
		return nil, ErrPedidoRequerido
	}

	var redencion *models.CuponRedencion
	err := s.inTx(ctx, func(ts *CuponService) error {
		pedido, err := ts.bloquearPedido(*req.PedidoId, req.ClienteId)
		if err != nil {
			return err
		}
		items, err := ts.itemsDePedido(pedido.PK_ID_PEDIDO)
		if err != nil {
			return err
		}
		redencion, err = ts.redimirEn("codigo", codigo, req.ClienteId, pedido, items)
		return err
	})
	if err != nil {
		return nil, err
	}
	return redencion, nil
}

// redimirEn bloquea la fila del cupón (campo = valor), revalida bajo el
// bloqueo y registra la redención. El pedido ya debe estar bloqueado y validado.
// Un mismo pedido no puede redimir el mismo cupón dos veces (ErrCuponConflicto).
func (s *CuponService) redimirEn(campo string, valor interface{}, clienteID int64, pedido *models.Pedido, items []models.ValidarCuponItemRequest) (*models.CuponRedencion, error) {
	cupon := &models.Cupon{}
	if err := s.ormer.QueryTable("cupon").Filter(campo, valor).ForUpdate().One(cupon); err != nil {
		if err == orm.ErrNoRows {
			return nil, ErrCuponNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar cupón: %w", err)
	}

	yaRedimido, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Filter("pk_id_pedido", pedido.PK_ID_PEDIDO).Count()
	if err != nil {
		return nil, fmt.Errorf("error al verificar redenciones del pedido: %w", err)
	}
	if yaRedimido > 0 {
		return nil, fmt.Errorf("%w: el cupón ya fue redimido para este pedido", ErrCuponConflicto)
	}

	validacion, conflicto, err := s.evaluar(cupon, clienteID, items)
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
		PkDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: clienteID},
		PkIdPedido:         &models.Pedido{PK_ID_PEDIDO: pedido.PK_ID_PEDIDO},
		MontoDescuento:     validacion.MontoDescuento,
	}
	if _, err = s.ormer.Insert(redencion); err != nil {
		if dberr.IsUnique(err) {
			return nil, fmt.Errorf("%w: el cupón ya fue redimido para este pedido", ErrCuponConflicto)
		}
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
