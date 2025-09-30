package services

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type DescuentoService struct {
	ormer orm.Ormer
}

func NewDescuentoService(ormer orm.Ormer) *DescuentoService {
	return &DescuentoService{ormer: ormer}
}

// AplicarDescuento registra la aplicación de un descuento a un pedido
func (s *DescuentoService) AplicarDescuento(ctx context.Context, pedidoId int64, req *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {
	// Validar exclusividad: solo cupón o solo oferta
	if (req.PkIdCupon == nil && req.PkIdOferta == nil) || (req.PkIdCupon != nil && req.PkIdOferta != nil) {
		return nil, fmt.Errorf("debe especificar exactamente uno de cupón o oferta")
	}

	// Verificar que el pedido existe
	pedido := &models.Pedido{PK_ID_PEDIDO: pedidoId}
	err := s.ormer.Read(pedido)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("pedido no encontrado")
		}
		return nil, fmt.Errorf("error al buscar pedido: %w", err)
	}

	// Verificar que no existe ya un descuento aplicado para este pedido
	count, err := s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Count()
	if err != nil {
		return nil, fmt.Errorf("error al verificar descuentos existentes: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("ya existe un descuento aplicado para este pedido")
	}

	// Crear el registro de descuento aplicado
	descuentoAplicado := &models.PedidoDescuentoAplicado{
		PkIdPedido:     pedido,
		MontoDescuento: req.MontoDescuento,
		DetalleObj:     req.Detalle,
	}

	// Verificar y asignar cupón o oferta
	if req.PkIdCupon != nil {
		cupon := &models.Cupon{PkIdCupon: *req.PkIdCupon}
		err := s.ormer.Read(cupon)
		if err != nil {
			if err == orm.ErrNoRows {
				return nil, fmt.Errorf("cupón no encontrado")
			}
			return nil, fmt.Errorf("error al buscar cupón: %w", err)
		}
		descuentoAplicado.PkIdCupon = cupon

		// Crear detalle específico para cupón
		detalleCupon := map[string]interface{}{
			"tipo":   "cupon",
			"codigo": cupon.Codigo,
			"scope":  cupon.Scope,
		}
		detalleJSON, _ := json.Marshal(detalleCupon)
		descuentoAplicado.DetalleObj = json.RawMessage(detalleJSON)
	}

	if req.PkIdOferta != nil {
		oferta := &models.Oferta{PkIdOferta: *req.PkIdOferta}
		err := s.ormer.Read(oferta)
		if err != nil {
			if err == orm.ErrNoRows {
				return nil, fmt.Errorf("oferta no encontrada")
			}
			return nil, fmt.Errorf("error al buscar oferta: %w", err)
		}
		descuentoAplicado.PkIdOferta = oferta

		// Crear detalle específico para oferta
		detalleOferta := map[string]interface{}{
			"tipo":   "oferta",
			"titulo": oferta.Titulo,
		}
		detalleJSON, _ := json.Marshal(detalleOferta)
		descuentoAplicado.DetalleObj = json.RawMessage(detalleJSON)
	}

	_, err = s.ormer.Insert(descuentoAplicado)
	if err != nil {
		return nil, fmt.Errorf("error al registrar descuento aplicado: %w", err)
	}

	return descuentoAplicado, nil
}

// ObtenerDescuentosPedido obtiene todos los descuentos aplicados a un pedido
func (s *DescuentoService) ObtenerDescuentosPedido(ctx context.Context, pedidoId int64) ([]*models.PedidoDescuentoAplicado, error) {
	var descuentos []*models.PedidoDescuentoAplicado
	_, err := s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).RelatedSel().All(&descuentos)
	if err != nil {
		return nil, fmt.Errorf("error al obtener descuentos del pedido: %w", err)
	}

	return descuentos, nil
}

// ValidarExclusividadDescuento verifica que no se apliquen múltiples descuentos al mismo pedido
func (s *DescuentoService) ValidarExclusividadDescuento(ctx context.Context, pedidoId int64, cuponId *int64, ofertaId *int64) error {
	// Verificar si ya existe un descuento para este pedido
	var count int64
	count, err := s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Count()
	if err != nil {
		return fmt.Errorf("error al verificar descuentos existentes: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("ya existe un descuento aplicado para este pedido")
	}

	// Si es un cupón, verificar que no se haya aplicado ya el mismo cupón al mismo pedido
	if cuponId != nil {
		count, err = s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Filter("pk_id_cupon", *cuponId).Count()
		if err != nil {
			return fmt.Errorf("error al verificar cupón duplicado: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("este cupón ya ha sido aplicado a este pedido")
		}
	}

	// Si es una oferta, verificar que no se haya aplicado ya la misma oferta al mismo pedido
	if ofertaId != nil {
		count, err = s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Filter("pk_id_oferta", *ofertaId).Count()
		if err != nil {
			return fmt.Errorf("error al verificar oferta duplicada: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("esta oferta ya ha sido aplicada a este pedido")
		}
	}

	return nil
}
