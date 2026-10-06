package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"restaurante/internal/dberr"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// Errores tipados de DescuentoService; el controlador los traduce a códigos HTTP.
var (
	// ErrOfertaNoEncontrada: la oferta indicada no existe (ErrPedidoNoEncontrado y
	// ErrCuponNoEncontrado se comparten con CuponService).
	ErrOfertaNoEncontrada = errors.New("oferta no encontrada")
	// ErrOfertaNoAplicable: la oferta no está vigente o no aplica al pedido.
	ErrOfertaNoAplicable = errors.New("oferta no aplicable")
	// ErrDescuentoYaAplicado: el pedido ya tiene un descuento aplicado.
	ErrDescuentoYaAplicado = errors.New("ya existe un descuento aplicado para este pedido")
	// ErrPedidoPagado: el pedido ya está pagado y su total no puede cambiar.
	ErrPedidoPagado = errors.New("el pedido ya está pagado")
	// ErrDescuentoInvalido: la solicitud no es válida (fuente o detalle).
	ErrDescuentoInvalido = errors.New("solicitud de descuento inválida")
)

type DescuentoService struct {
	ormer orm.Ormer
}

func NewDescuentoService(ormer orm.Ormer) *DescuentoService {
	return &DescuentoService{ormer: ormer}
}

// AplicarDescuento aplica un único descuento (cupón u oferta) al pedido del
// cliente, todo en UNA transacción: bloquea el pedido (y su pago) y el cupón con
// SELECT ... FOR UPDATE, valida, CALCULA el monto en el servidor con el detalle
// del pedido (nunca se acepta un monto del cliente), redime el cupón, registra
// el PedidoDescuentoAplicado y descuenta el monto de `pago.monto` del pedido
// (si tiene pago). Si algo falla no queda nada a medias.
//
// `detalle` del request, si es un objeto JSON, se conserva y se le agregan los
// datos del cupón/oferta (tipo, codigo/titulo...), que prevalecen.
func (s *DescuentoService) AplicarDescuento(ctx context.Context, pedidoId, clienteId int64, req *models.AplicarDescuentoRequest) (*models.DescuentoAplicadoResponse, error) {
	if (req.PkIdCupon == nil && req.PkIdOferta == nil) || (req.PkIdCupon != nil && req.PkIdOferta != nil) {
		return nil, fmt.Errorf("%w: debe especificar exactamente uno de cupón o oferta", ErrDescuentoInvalido)
	}

	detalle := map[string]interface{}{}
	if len(req.Detalle) > 0 {
		if err := json.Unmarshal(req.Detalle, &detalle); err != nil || detalle == nil {
			return nil, fmt.Errorf("%w: detalle debe ser un objeto JSON", ErrDescuentoInvalido)
		}
	}

	var resp *models.DescuentoAplicadoResponse
	err := doTx(ctx, s.ormer, func(tx orm.TxOrmer) error {
		var err error
		resp, err = aplicarEnTx(tx, pedidoId, clienteId, req, detalle)
		return err
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// aplicarEnTx ejecuta AplicarDescuento dentro de la transacción tx. Orden de
// bloqueo: pedido, pago, cupón (el mismo en toda la base, para evitar deadlocks).
func aplicarEnTx(tx orm.TxOrmer, pedidoId, clienteId int64, req *models.AplicarDescuentoRequest, detalle map[string]interface{}) (*models.DescuentoAplicadoResponse, error) {
	cupones := newCuponServiceFromTx(tx)

	pedido, err := cupones.bloquearPedido(pedidoId, clienteId)
	if err != nil {
		return nil, err
	}

	var pago *models.Pago
	if pedido.PK_ID_PAGO != nil {
		pago = &models.Pago{}
		if err = tx.QueryTable("pago").Filter("pk_id_pago", pedido.PK_ID_PAGO.PK_ID_PAGO).ForUpdate().One(pago); err != nil {
			return nil, fmt.Errorf("error al buscar el pago del pedido: %w", err)
		}
		if pago.ESTADO_PAGO == models.EstadoPagoPagado {
			return nil, ErrPedidoPagado
		}
	}

	count, err := tx.QueryTable("pedido_descuento_aplicado").Filter("pk_id_pedido", pedidoId).Count()
	if err != nil {
		return nil, fmt.Errorf("error al verificar descuentos existentes: %w", err)
	}
	if count > 0 {
		return nil, ErrDescuentoYaAplicado
	}

	items, err := cupones.itemsDePedido(pedidoId)
	if err != nil {
		return nil, err
	}
	var subtotal int64
	for _, it := range items {
		subtotal += it.Precio * int64(it.Cantidad)
	}

	aplicado := &models.PedidoDescuentoAplicado{PkIdPedido: &models.Pedido{PK_ID_PEDIDO: pedidoId}}
	var monto int64

	if req.PkIdCupon != nil {
		redencion, err := cupones.redimirEn("pk_id_cupon", *req.PkIdCupon, clienteId, pedido, items)
		if err != nil {
			return nil, err
		}
		monto = redencion.MontoDescuento
		aplicado.PkIdCupon = redencion.PkIdCupon
		detalle["tipo"] = "cupon"
		detalle["codigo"] = redencion.PkIdCupon.Codigo
		detalle["scope"] = redencion.PkIdCupon.Scope
	} else {
		oferta, err := evaluarOferta(tx, *req.PkIdOferta, pedido, items)
		if err != nil {
			return nil, err
		}
		monto = oferta.monto
		aplicado.PkIdOferta = oferta.oferta
		detalle["tipo"] = "oferta"
		detalle["titulo"] = oferta.oferta.Titulo
	}
	aplicado.MontoDescuento = monto

	// El mapa solo contiene tipos JSON, así que siempre se puede serializar.
	detalleJSON, _ := json.Marshal(detalle)
	aplicado.DetalleObj = json.RawMessage(detalleJSON)
	aplicado.BeforeInsert()

	if _, err = tx.Insert(aplicado); err != nil {
		if dberr.IsUnique(err) {
			return nil, ErrDescuentoYaAplicado
		}
		return nil, fmt.Errorf("error al registrar descuento aplicado: %w", err)
	}

	resp := &models.DescuentoAplicadoResponse{
		Descuento:      aplicado,
		Subtotal:       subtotal,
		MontoDescuento: monto,
		Total:          subtotal - monto,
	}
	if pago != nil {
		pago.MONTO = pago.MONTO - monto
		if pago.MONTO < 0 {
			pago.MONTO = 0
		}
		pago.UPDATED_AT = ahoraBogota()
		if _, err = tx.Update(pago, "MONTO", "UPDATED_AT"); err != nil {
			return nil, fmt.Errorf("error al actualizar el total del pago: %w", err)
		}
		resp.Total = pago.MONTO
		resp.PagoId = &pago.PK_ID_PAGO
	}
	return resp, nil
}

// ofertaEvaluada es el resultado de evaluarOferta.
type ofertaEvaluada struct {
	oferta *models.Oferta
	monto  int64
}

// evaluarOferta carga la oferta, comprueba vigencia (período, día y horario en
// hora de Bogotá), que rija en el restaurante del pedido y calcula el descuento
// sobre los productos del pedido que pertenecen a la oferta.
func evaluarOferta(tx orm.TxOrmer, ofertaID int64, pedido *models.Pedido, items []models.ValidarCuponItemRequest) (*ofertaEvaluada, error) {
	oferta := &models.Oferta{PkIdOferta: ofertaID}
	if err := tx.Read(oferta); err != nil {
		if err == orm.ErrNoRows {
			return nil, ErrOfertaNoEncontrada
		}
		return nil, fmt.Errorf("error al buscar oferta: %w", err)
	}
	oferta.AfterLoad()

	ofertas := NewOfertaService(tx)
	if motivo := ofertas.MotivoNoVigente(oferta, ahoraBogota()); motivo != "" {
		return nil, fmt.Errorf("%w: %s", ErrOfertaNoAplicable, motivo)
	}
	if pedido.PK_ID_RESTAURANTE != nil && oferta.PkIdRestaurante != nil && oferta.PkIdRestaurante.PK_ID_RESTAURANTE != pedido.PK_ID_RESTAURANTE.PK_ID_RESTAURANTE {
		return nil, fmt.Errorf("%w: la oferta no rige en el restaurante del pedido", ErrOfertaNoAplicable)
	}

	monto, err := ofertas.CalcularDescuentoOferta(oferta, items)
	if err != nil {
		return nil, err
	}
	if monto <= 0 {
		return nil, fmt.Errorf("%w: ningún producto del pedido pertenece a la oferta", ErrOfertaNoAplicable)
	}
	return &ofertaEvaluada{oferta: oferta, monto: monto}, nil
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
