package services

import (
	"context"
	"fmt"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type CuponService struct {
	ormer orm.Ormer
}

func NewCuponService(ormer orm.Ormer) *CuponService {
	return &CuponService{ormer: ormer}
}

// ValidarCupon valida si un cupón puede ser aplicado a un pedido
func (s *CuponService) ValidarCupon(ctx context.Context, req *models.ValidarCuponRequest) (*models.ValidarCuponResponse, error) {
	// Buscar el cupón por código
	cupon := &models.Cupon{}
	err := s.ormer.QueryTable("cupon").Filter("codigo", req.Codigo).One(cupon)
	if err != nil {
		if err == orm.ErrNoRows {
			return &models.ValidarCuponResponse{
				Aplicable: false,
				Motivo:    stringPtr("Cupón no encontrado"),
			}, nil
		}
		return nil, fmt.Errorf("error al buscar cupón: %w", err)
	}

	// Validar si el cupón está activo
	if !cupon.Activo {
		return &models.ValidarCuponResponse{
			Aplicable: false,
			Motivo:    stringPtr("Cupón inactivo"),
		}, nil
	}

	// Validar fechas de vigencia
	now := time.Now()
	if now.Before(cupon.FechaInicio) || now.After(cupon.FechaFin) {
		return &models.ValidarCuponResponse{
			Aplicable: false,
			Motivo:    stringPtr("Cupón fuera del período de vigencia"),
		}, nil
	}

	// Validar límite de usos globales
	if cupon.MaxUsos != nil {
		var usosActuales int64
		_, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Count()
		if err != nil {
			return nil, fmt.Errorf("error al contar usos del cupón: %w", err)
		}
		if usosActuales >= int64(*cupon.MaxUsos) {
			return &models.ValidarCuponResponse{
				Aplicable: false,
				Motivo:    stringPtr("Cupón ha alcanzado el límite máximo de usos"),
			}, nil
		}
	}

	// Validar límite de usos por cliente
	if cupon.LimitePorCliente != nil {
		var usosCliente int64
		_, err := s.ormer.QueryTable("cupon_redencion").Filter("pk_id_cupon", cupon.PkIdCupon).Filter("pk_documento_cliente", req.ClienteId).Count()
		if err != nil {
			return nil, fmt.Errorf("error al contar usos del cliente: %w", err)
		}
		if usosCliente >= int64(*cupon.LimitePorCliente) {
			return &models.ValidarCuponResponse{
				Aplicable: false,
				Motivo:    stringPtr("Cliente ha alcanzado el límite de usos para este cupón"),
			}, nil
		}
	}

	// Calcular monto total del pedido
	montoTotal := int64(0)
	productosAplicables := []int64{}

	for _, item := range req.Items {
		montoTotal += item.Precio * int64(item.Cantidad)

		// Verificar si el producto es aplicable según el scope del cupón
		if s.esProductoAplicable(cupon, item.ProductoId) {
			productosAplicables = append(productosAplicables, item.ProductoId)
		}
	}

	// Validar monto mínimo
	if cupon.MontoMinimo != nil && montoTotal < *cupon.MontoMinimo {
		return &models.ValidarCuponResponse{
			Aplicable: false,
			Motivo:    stringPtr(fmt.Sprintf("El monto mínimo requerido es %d", *cupon.MontoMinimo)),
		}, nil
	}

	// Validar scope específico
	if cupon.Scope == models.CuponScopeCliente && cupon.PkDocumentoCliente != nil {
		if cupon.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE != req.ClienteId {
			return &models.ValidarCuponResponse{
				Aplicable: false,
				Motivo:    stringPtr("Cupón no válido para este cliente"),
			}, nil
		}
	}

	// Si el scope es por producto o categoría, verificar que haya productos aplicables
	if (cupon.Scope == models.CuponScopeProducto || cupon.Scope == models.CuponScopeCategoria) && len(productosAplicables) == 0 {
		return &models.ValidarCuponResponse{
			Aplicable: false,
			Motivo:    stringPtr("No hay productos aplicables para este cupón"),
		}, nil
	}

	// Calcular descuento
	montoDescuento := s.calcularDescuento(cupon, montoTotal, req.Items, productosAplicables)

	return &models.ValidarCuponResponse{
		Aplicable:      true,
		MontoDescuento: montoDescuento,
	}, nil
}

// RedimirCupon registra la redención de un cupón
func (s *CuponService) RedimirCupon(ctx context.Context, codigo string, req *models.RedimirCuponRequest) (*models.CuponRedencion, error) {
	// Buscar el cupón
	cupon := &models.Cupon{}
	err := s.ormer.QueryTable("cupon").Filter("codigo", codigo).One(cupon)
	if err != nil {
		return nil, fmt.Errorf("cupón no encontrado: %w", err)
	}

	// Validar que el cupón pueda ser redimido (reutilizar lógica de validación)
	validacionReq := &models.ValidarCuponRequest{
		ClienteId: req.ClienteId,
		Codigo:    codigo,
		Items:     []models.ValidarCuponItemRequest{}, // Se necesitarían los items del pedido real
	}

	validacion, err := s.ValidarCupon(ctx, validacionReq)
	if err != nil {
		return nil, fmt.Errorf("error al validar cupón: %w", err)
	}

	if !validacion.Aplicable {
		return nil, fmt.Errorf("cupón no aplicable: %s", *validacion.Motivo)
	}

	// Crear la redención
	redencion := &models.CuponRedencion{
		PkIdCupon:          cupon,
		PkDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: req.ClienteId},
		MontoDescuento:     validacion.MontoDescuento,
	}

	if req.PedidoId != nil {
		redencion.PkIdPedido = &models.Pedido{PK_ID_PEDIDO: *req.PedidoId}
	}

	_, err = s.ormer.Insert(redencion)
	if err != nil {
		return nil, fmt.Errorf("error al registrar redención: %w", err)
	}

	return redencion, nil
}

// esProductoAplicable verifica si un producto es aplicable según el scope del cupón
func (s *CuponService) esProductoAplicable(cupon *models.Cupon, productoId int64) bool {
	switch cupon.Scope {
	case models.CuponScopeGlobal:
		return true
	case models.CuponScopeProducto:
		return cupon.PkIdProducto != nil && cupon.PkIdProducto.PK_ID_PRODUCTO == productoId
	case models.CuponScopeCategoria:
		if cupon.PkIdCategoria == nil {
			return false
		}
		// Buscar el producto y verificar su categoría a través de subcategoría
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

		return subcategoria.PK_ID_CATEGORIA.PK_ID_CATEGORIA == cupon.PkIdCategoria.PK_ID_CATEGORIA
	case models.CuponScopeCliente:
		return true // Ya se validó el cliente en ValidarCupon
	}
	return false
}

// calcularDescuento calcula el monto de descuento según el tipo y valor del cupón
func (s *CuponService) calcularDescuento(cupon *models.Cupon, montoTotal int64, items []models.ValidarCuponItemRequest, productosAplicables []int64) int64 {
	var montoAplicable int64

	if cupon.Scope == models.CuponScopeGlobal || cupon.Scope == models.CuponScopeCliente {
		montoAplicable = montoTotal
	} else {
		// Solo productos aplicables
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

// ValidarReglasNegocioCupon valida las reglas de negocio al crear/actualizar un cupón
func (s *CuponService) ValidarReglasNegocioCupon(cupon *models.Cupon) error {
	// Validar tipo de descuento y valor
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

	// Validar fechas
	if cupon.FechaFin.Before(cupon.FechaInicio) {
		return fmt.Errorf("la fecha de fin debe ser posterior a la fecha de inicio")
	}

	// Validar scope vs targets
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
