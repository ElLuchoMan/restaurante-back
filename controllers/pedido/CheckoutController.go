package pedido

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"restaurante/controllers/login"
	"restaurante/internal/authz"
	"restaurante/internal/httpx"
	"restaurante/internal/inventario"
	"restaurante/internal/montopedido"
	"restaurante/internal/notify"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

const (
	msgCheckoutInventario = "Inventario insuficiente para uno o más productos"
	msgMetodoNoEncontrado = "Método de pago no encontrado"
)

// checkoutError es un fallo del checkout: de validación (status + msg), de
// inventario (409 con `data`) o de base de datos (op != ""; 409 si es un
// conflicto de PostgreSQL y 500 en otro caso).
type checkoutError struct {
	status int
	msg    string
	data   any
	op     string
	err    error
}

func rechazo(status int, msg string) *checkoutError { return &checkoutError{status: status, msg: msg} }

func falloBD(op, msg string, err error) *checkoutError {
	return &checkoutError{op: op, msg: msg, err: err}
}

func (c *PedidoController) responderCheckout(e *checkoutError) {
	switch {
	case e.op != "":
		c.dbError(e.op, e.msg, e.err, nil)
	case e.data != nil:
		httpx.Send(&c.Controller, e.status, e.msg, e.data)
	default:
		c.fail(e.status, e.msg, nil)
	}
}

// checkoutPlan es el cuerpo de POST /pedidos/checkout ya validado y normalizado.
type checkoutPlan struct {
	restaurante int64
	cliente     int64 // 0 = pedido de mostrador (solo personal)
	metodoPago  int64
	items       map[int64]int
	domicilio   *planDomicilio
	pago        planPago
	montoManual int64 // > 0 solo si lo envió el personal
	creadoPor   string
	hoy         time.Time // fecha de hoy en Bogotá (a mediodía UTC) para el pedido
}

type planDomicilio struct {
	direccion, telefono string
	fecha               time.Time
	observaciones       *string
}

type planPago struct {
	fecha, hora time.Time
	estado      string
}

// parseHoraPago valida HH:MM[:SS] y devuelve la hora con año 2000 (como /pagos,
// para que FormatTimeWithLMT no la corrija al serializar).
func parseHoraPago(s string) (time.Time, error) {
	if len(s) != len("15:04:05") && len(s) != len("15:04") {
		return time.Time{}, errors.New("formato esperado HH:MM o HH:MM:SS")
	}
	t, err := models.ParseTimeToUTC(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(2000, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
}

// planearCheckout valida el cuerpo y quién llama. Devuelve el plan o el rechazo
// (400 cuerpo inválido, 403 cliente ajeno o estado de pago no permitido).
func planearCheckout(claims *login.Claims, in *models.CheckoutRequest, ahora time.Time) (*checkoutPlan, *checkoutError) {
	staff := claims.IsStaff()
	plan := &checkoutPlan{restaurante: in.RestauranteId, metodoPago: in.Pago.MetodoPagoId}
	if !staff {
		// Un Cliente siempre compra a su nombre: el documento sale del token y otro distinto en el body es 403.
		var body int64
		if in.DocumentoCliente != nil {
			body = *in.DocumentoCliente
		}
		switch {
		case claims.Documento <= 0:
			return nil, rechazo(http.StatusForbidden, "El token no identifica a un cliente")
		case body != 0 && body != claims.Documento:
			return nil, rechazo(http.StatusForbidden, "No puede actuar en nombre de otro cliente")
		}
		plan.cliente = claims.Documento
	} else if in.DocumentoCliente != nil {
		if *in.DocumentoCliente <= 0 {
			return nil, rechazo(http.StatusBadRequest, "documentoCliente debe ser un entero positivo")
		}
		plan.cliente = *in.DocumentoCliente
	}
	if in.RestauranteId < 0 {
		return nil, rechazo(http.StatusBadRequest, "restauranteId debe ser un entero positivo")
	}
	if len(in.Productos) == 0 {
		return nil, rechazo(http.StatusBadRequest, "Debe enviar al menos un producto")
	}
	plan.items = make(map[int64]int, len(in.Productos))
	for i, it := range in.Productos {
		if it.ProductoId <= 0 || it.Cantidad <= 0 {
			return nil, rechazo(http.StatusBadRequest, fmt.Sprintf("productos[%d]: productoId y cantidad deben ser enteros positivos", i))
		}
		plan.items[it.ProductoId] += it.Cantidad
	}
	if in.Pago.MetodoPagoId <= 0 {
		return nil, rechazo(http.StatusBadRequest, "pago.metodoPagoId es obligatorio y debe ser un entero positivo")
	}
	if staff && in.Pago.Monto < 0 {
		return nil, rechazo(http.StatusBadRequest, "pago.monto no puede ser negativo")
	}
	if staff {
		plan.montoManual = in.Pago.Monto
	}
	plan.creadoPor = fmt.Sprintf("Usuario %d", claims.Documento)

	hoy := time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 12, 0, 0, 0, time.UTC)
	plan.hoy = hoy
	plan.pago = planPago{
		fecha:  hoy,
		hora:   time.Date(2000, 1, 1, ahora.Hour(), ahora.Minute(), ahora.Second(), 0, time.UTC),
		estado: models.EstadoPagoPendiente,
	}
	var err error
	if f := strings.TrimSpace(in.Pago.FechaPago); f != "" {
		if plan.pago.fecha, err = models.ParseDateToNoonUTC(f); err != nil {
			return nil, rechazo(http.StatusBadRequest, "pago.fechaPago inválida (use YYYY-MM-DD)")
		}
	}
	if h := strings.TrimSpace(in.Pago.HoraPago); h != "" {
		if plan.pago.hora, err = parseHoraPago(h); err != nil {
			return nil, rechazo(http.StatusBadRequest, "pago.horaPago inválida (use HH:MM o HH:MM:SS)")
		}
	}
	if e := strings.ToUpper(strings.TrimSpace(in.Pago.EstadoPago)); e != "" {
		switch e {
		case models.EstadoPagoPagado, models.EstadoPagoPendiente, models.EstadoPagoNoPago:
			plan.pago.estado = e
		default:
			return nil, rechazo(http.StatusBadRequest, "pago.estadoPago inválido (PAGADO, PENDIENTE o NO_PAGO)")
		}
	}
	if !staff && plan.pago.estado != models.EstadoPagoPendiente {
		return nil, rechazo(http.StatusForbidden, "Un cliente solo puede crear pagos en estado PENDIENTE")
	}

	if d := in.Domicilio; d != nil {
		dom := &planDomicilio{direccion: strings.TrimSpace(d.Direccion), telefono: strings.TrimSpace(d.Telefono), fecha: hoy, observaciones: d.Observaciones}
		if dom.direccion == "" || dom.telefono == "" {
			return nil, rechazo(http.StatusBadRequest, "domicilio.direccion y domicilio.telefono son obligatorios")
		}
		if f := strings.TrimSpace(d.FechaDomicilio); f != "" {
			if dom.fecha, err = models.ParseDateToNoonUTC(f); err != nil {
				return nil, rechazo(http.StatusBadRequest, "domicilio.fechaDomicilio inválida (use YYYY-MM-DD)")
			}
		}
		plan.domicilio = dom
	}
	return plan, nil
}

// existeEnTx responde 404 (msg) si la fila column = id no existe en tx y un fallo de BD si la consulta falla.
func existeEnTx(tx orm.TxOrmer, model interface{}, column string, id int64, msg string) *checkoutError {
	ok, err := exists(tx, model, column, id)
	if err != nil {
		return falloBD("checkout.referencia_error", "Error al validar los datos referenciados", err)
	}
	if !ok {
		return rechazo(http.StatusNotFound, msg)
	}
	return nil
}

func validarReferencias(tx orm.TxOrmer, p *checkoutPlan) *checkoutError {
	if p.restaurante > 0 {
		if e := existeEnTx(tx, new(models.Restaurante), "PK_ID_RESTAURANTE", p.restaurante, "Restaurante no encontrado"); e != nil {
			return e
		}
	}
	if e := existeEnTx(tx, new(models.MetodoPago), "PK_ID_METODO_PAGO", p.metodoPago, msgMetodoNoEncontrado); e != nil {
		return e
	}
	if p.cliente > 0 {
		return existeEnTx(tx, new(models.Cliente), "PK_DOCUMENTO_CLIENTE", p.cliente, "Cliente no encontrado")
	}
	return nil
}

// reservarInventario bloquea y descuenta las unidades del plan (404 producto
// inexistente, 409 inventario insuficiente con el detalle por producto).
func reservarInventario(tx orm.TxOrmer, items map[int64]int) *checkoutError {
	res, err := inventario.Bloquear(tx, items)
	if err != nil {
		return falloBD("checkout.inventario_error", "Error al validar inventario", err)
	}
	if len(res.NoExisten) > 0 {
		ids := make([]string, len(res.NoExisten))
		for i, id := range res.NoExisten {
			ids[i] = fmt.Sprint(id)
		}
		return rechazo(http.StatusNotFound, "Producto no encontrado: "+strings.Join(ids, ", "))
	}
	if len(res.Insuficientes) > 0 {
		return &checkoutError{status: http.StatusConflict, msg: msgCheckoutInventario, data: res.Insuficientes}
	}
	if _, err := inventario.Descontar(tx, items); err != nil {
		return falloBD("checkout.descontar_error", "Error al descontar el inventario", err)
	}
	return nil
}

func insertarDomicilio(tx orm.TxOrmer, p *checkoutPlan) (*models.Domicilio, *checkoutError) {
	cols := []string{"direccion", "fecha", "telefono", "created_by"}
	vals := []interface{}{p.domicilio.direccion, p.domicilio.fecha, p.domicilio.telefono, p.creadoPor}
	if p.domicilio.observaciones != nil {
		cols, vals = append(cols, "observaciones"), append(vals, *p.domicilio.observaciones)
	}
	query := fmt.Sprintf("INSERT INTO domicilio (%s) VALUES (%s) RETURNING pk_id_domicilio", strings.Join(cols, ","), strings.TrimSuffix(strings.Repeat("?,", len(vals)), ","))
	var dom models.Domicilio
	if err := tx.Raw(query, vals...).QueryRow(&dom.ID); err != nil {
		return nil, falloBD("checkout.domicilio_error", "Error al crear el domicilio", err)
	}
	// Se relee para devolver lo que fija la base (estado por defecto, entregado, marcas de tiempo).
	if err := tx.Read(&dom); err != nil {
		return nil, falloBD("checkout.domicilio_read_error", "Error al leer el domicilio creado", err)
	}
	return &dom, nil
}

// ejecutarCheckout hace todo el trabajo de escritura dentro de tx. No confirma
// ni deshace: quien llama decide (cualquier error implica rollback).
func ejecutarCheckout(tx orm.TxOrmer, p *checkoutPlan) (*models.CheckoutResult, *checkoutError) {
	if e := validarReferencias(tx, p); e != nil {
		return nil, e
	}
	if e := reservarInventario(tx, p.items); e != nil {
		return nil, e
	}
	ahora := time.Now().UTC()
	pedido := models.Pedido{
		FECHA:         p.hoy,
		HORA:          ahora,
		ESTADO_PEDIDO: models.EstadoPedidoIniciado,
		UPDATED_AT:    ahora,
	}
	if p.domicilio != nil {
		dom, e := insertarDomicilio(tx, p)
		if e != nil {
			return nil, e
		}
		pedido.DELIVERY, pedido.PK_ID_DOMICILIO = true, dom
	}
	if p.restaurante > 0 {
		pedido.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: p.restaurante}
	}
	if p.cliente > 0 {
		pedido.PK_DOCUMENTO_CLIENTE = &models.Cliente{PK_DOCUMENTO_CLIENTE: p.cliente}
	}
	if _, err := tx.Insert(&pedido); err != nil {
		return nil, falloBD("checkout.pedido_error", "Error al crear el pedido", err)
	}
	for _, pid := range inventario.SortedIDs(p.items) {
		detalle := models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedido.PK_ID_PEDIDO}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: pid}, Cantidad: p.items[pid]}
		if _, err := tx.Insert(&detalle); err != nil {
			return nil, falloBD("checkout.detalle_error", "Error al guardar los productos del pedido", err)
		}
	}
	// El monto sale de los detalles recién insertados (el precio lo fija la base): el body nunca lo decide.
	m, err := montopedido.Calcular(tx, pedido.PK_ID_PEDIDO)
	if err != nil {
		return nil, falloBD("checkout.monto_error", "Error al calcular el monto del pedido", err)
	}
	if p.montoManual > 0 {
		m.Total = p.montoManual
	}
	pago := models.Pago{
		FECHA:             p.pago.fecha,
		HORA:              p.pago.hora,
		MONTO:             m.Total,
		ESTADO_PAGO:       p.pago.estado,
		PK_ID_METODO_PAGO: &models.MetodoPago{PK_ID_METODO_PAGO: p.metodoPago},
		UPDATED_AT:        ahora,
	}
	if _, err := tx.Insert(&pago); err != nil {
		return nil, falloBD("checkout.pago_error", "Error al crear el pago", err)
	}
	pedido.PK_ID_PAGO = &pago
	if _, err := tx.Update(&pedido, "PK_ID_PAGO", "UPDATED_AT"); err != nil {
		return nil, falloBD("checkout.enlazar_pago_error", "Error al asignar el pago al pedido", err)
	}
	return &models.CheckoutResult{Pedido: pedido, Monto: m.Total}, nil
}

// @Title Checkout
// @Summary Crear pedido, productos, pago y domicilio en una sola operación atómica
// @Description Reemplaza la secuencia `POST /domicilios` + `POST /pedidos` + `POST /producto_pedido` + `POST /pagos` + `POST /pedidos/asignar-pago`. TODO ocurre en UNA transacción: se validan las referencias (restaurante si se envía, método de pago y cliente), se bloquea y descuenta el inventario (`FOR UPDATE`), se crea el domicilio (si viene `domicilio`; entonces el pedido es `delivery`), el pedido (INICIADO, fecha y hora del servidor en Bogotá), sus detalles y el pago, y se enlazan pago y domicilio al pedido. Si algo falla se deshace todo y no queda ningún pedido, domicilio ni pago huérfano (reintentar es seguro). EL SERVIDOR MANDA EL MONTO: `pago.monto` = `MAX(0, SUM(precio x cantidad) - descuentos)` calculado con los precios vigentes; el `monto` del body se IGNORA salvo el del personal (ajuste manual > 0; negativo responde 400). `productos` (al menos uno) lleva `cantidad` > 0 (las líneas repetidas se suman). `pago.fechaPago`/`pago.horaPago` por defecto son ahora en Bogotá y `pago.estadoPago` por defecto PENDIENTE (un Cliente solo puede PENDIENTE: 403). Quién compra: un Cliente siempre a su nombre (el documento sale del token; `documentoCliente` distinto responde 403); el personal puede enviar `documentoCliente` (404 si no existe) o dejarlo vacío (pedido de mostrador). Responde 201 con el pedido completo (misma forma que `GET /pedidos`, con `pagoId` y `domicilioId` ya enlazados) más `monto`. Las notificaciones push (cliente y trabajadores, best-effort y en segundo plano) se envían solo después de confirmar la transacción.
// @Tags pedido
// @Accept json
// @Produce json
// @Param body body models.CheckoutRequest true "Productos, pago y domicilio opcional"
// @Success 201 {object} models.ApiResponse{data=models.CheckoutRespuestaDoc} "Pedido creado con su pago (y domicilio) y el monto calculado por el servidor"
// @Failure 400 {object} models.ApiResponse "JSON inválido, sin productos, productoId o cantidad no positivos, método de pago ausente, fecha/hora/estado de pago inválidos, domicilio sin dirección o teléfono, o monto negativo (personal)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente envió otro documentoCliente, su token no identifica a un cliente o pidió un estado de pago distinto de PENDIENTE"
// @Failure 404 {object} models.ApiResponse "No existe el restaurante, el método de pago, el cliente o algún producto"
// @Failure 409 {object} models.ApiResponse{data=[]models.InventarioInsuficienteDoc} "Inventario insuficiente (`data` lista {productoId, requerido, disponible}) o conflicto con datos existentes"
// @Failure 500 {object} models.ApiResponse "Error al crear el pedido (no se persistió nada)"
// @Security BearerAuth
// @Router /pedidos/checkout [post]
func (c *PedidoController) Checkout() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	var in models.CheckoutRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.checkout.bad_json", err, nil)
		c.fail(http.StatusBadRequest, "Datos inválidos (JSON mal formado)", err)
		return
	}
	plan, e := planearCheckout(claims, &in, ahoraBogota())
	if e != nil {
		c.responderCheckout(e)
		return
	}
	tx, err := orm.NewOrm().Begin()
	if err != nil {
		c.dbError("checkout.tx_begin_error", "Error al crear el pedido", err, nil)
		return
	}
	confirmada := false
	defer func() {
		if !confirmada {
			_ = tx.Rollback()
		}
	}()
	res, e := ejecutarCheckout(tx, plan)
	if e != nil {
		c.responderCheckout(e)
		return
	}
	if err := tx.Commit(); err != nil {
		c.dbError("checkout.tx_commit_error", "Error al confirmar el pedido", err, map[string]interface{}{"pedido_id": res.Pedido.PK_ID_PEDIDO})
		return
	}
	confirmada = true
	// Solo después del commit: si el pedido no existe en la base no se avisa a nadie.
	notify.Enviar(notify.Evento{Tipo: notify.PedidoCreado, PedidoID: res.Pedido.PK_ID_PEDIDO, Cliente: clienteDe(&res.Pedido), DomicilioID: domicilioDe(&res.Pedido)})
	httpx.Send(&c.Controller, http.StatusCreated, "Pedido creado exitosamente", *res)
}
