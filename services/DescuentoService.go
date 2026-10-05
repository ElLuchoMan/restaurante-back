package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// Errores tipados de DescuentoService; el controlador los traduce a códigos HTTP.
var (
	// ErrOfertaNoEncontrada: la oferta indicada no existe (ErrPedidoNoEncontrado y
	// ErrCuponNoEncontrado se comparten con CuponService).
	ErrOfertaNoEncontrada = errors.New("oferta no encontrada")
	// ErrDescuentoYaAplicado: el pedido ya tiene un descuento aplicado.
	ErrDescuentoYaAplicado = errors.New("ya existe un descuento aplicado para este pedido")
	// ErrDescuentoInvalido: la solicitud no es válida (fuente, monto o detalle).
	ErrDescuentoInvalido = errors.New("solicitud de descuento inválida")
)

type DescuentoService struct {
	ormer orm.Ormer
}

func NewDescuentoService(ormer orm.Ormer) *DescuentoService {
	return &DescuentoService{ormer: ormer}
}

// AplicarDescuento registra un único descuento (cupón u oferta) sobre un pedido.
// `detalle` del request, si es un objeto JSON, se conserva y se le agregan los
// datos del cupón/oferta (tipo, codigo/titulo...), que prevalecen.
func (s *DescuentoService) AplicarDescuento(ctx context.Context, pedidoId int64, req *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {

	if (req.PkIdCupon == nil && req.PkIdOferta == nil) || (req.PkIdCupon != nil && req.PkIdOferta != nil) {
		return nil, fmt.Errorf("%w: debe especificar exactamente uno de cupón o oferta", ErrDescuentoInvalido)
	}
	if req.MontoDescuento < 0 {
		return nil, fmt.Errorf("%w: montoDescuento no puede ser negativo", ErrDescuentoInvalido)
	}

	detalle := map[string]interface{}{}
	if len(req.Detalle) > 0 {
		if err := json.Unmarshal(req.Detalle, &detalle); err != nil || detalle == nil {
			return nil, fmt.Errorf("%w: detalle debe ser un objeto JSON", ErrDescuentoInvalido)
		}
	}

	pedido := &models.Pedido{PK_ID_PEDIDO: pedidoId}
	err := s.ormer.Read(pedido)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, ErrPedidoNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar pedido: %w", err)
	}

	count, err := s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Count()
	if err != nil {
		return nil, fmt.Errorf("error al verificar descuentos existentes: %w", err)
	}
	if count > 0 {
		return nil, ErrDescuentoYaAplicado
	}

	descuentoAplicado := &models.PedidoDescuentoAplicado{
		PkIdPedido:     pedido,
		MontoDescuento: req.MontoDescuento,
	}

	if req.PkIdCupon != nil {
		cupon := &models.Cupon{PkIdCupon: *req.PkIdCupon}
		err := s.ormer.Read(cupon)
		if err != nil {
			if err == orm.ErrNoRows {
				return nil, ErrCuponNoEncontrado
			}
			return nil, fmt.Errorf("error al buscar cupón: %w", err)
		}
		descuentoAplicado.PkIdCupon = cupon

		detalle["tipo"] = "cupon"
		detalle["codigo"] = cupon.Codigo
		detalle["scope"] = cupon.Scope
	}

	if req.PkIdOferta != nil {
		oferta := &models.Oferta{PkIdOferta: *req.PkIdOferta}
		err := s.ormer.Read(oferta)
		if err != nil {
			if err == orm.ErrNoRows {
				return nil, ErrOfertaNoEncontrada
			}
			return nil, fmt.Errorf("error al buscar oferta: %w", err)
		}
		oferta.AfterLoad()
		descuentoAplicado.PkIdOferta = oferta

		detalle["tipo"] = "oferta"
		detalle["titulo"] = oferta.Titulo
	}

	// El mapa solo contiene tipos JSON, así que siempre se puede serializar.
	detalleJSON, _ := json.Marshal(detalle)
	descuentoAplicado.DetalleObj = json.RawMessage(detalleJSON)
	descuentoAplicado.BeforeInsert()

	_, err = s.ormer.Insert(descuentoAplicado)
	if err != nil {
		return nil, fmt.Errorf("error al registrar descuento aplicado: %w", err)
	}

	return descuentoAplicado, nil
}

// ObtenerDescuentosPedido lista los descuentos del pedido (nunca nil). Carga solo
// las relaciones pedido, cupón y oferta (no el cliente del pedido).
func (s *DescuentoService) ObtenerDescuentosPedido(ctx context.Context, pedidoId int64) ([]*models.PedidoDescuentoAplicado, error) {
	descuentos := []*models.PedidoDescuentoAplicado{}
	_, err := s.ormer.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).RelatedSel("PkIdPedido", "PkIdCupon", "PkIdOferta").All(&descuentos)
	if err != nil {
		return nil, fmt.Errorf("error al obtener descuentos del pedido: %w", err)
	}
	for _, d := range descuentos {
		d.AfterLoad()
		if d.PkIdOferta != nil {
			d.PkIdOferta.AfterLoad()
		}
	}

	return descuentos, nil
}
