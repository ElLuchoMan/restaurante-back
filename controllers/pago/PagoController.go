package pago

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"restaurante/controllers/login"
	"restaurante/internal/authz"
	"restaurante/internal/httpx"
	"restaurante/internal/montopedido"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PagoController struct {
	web.Controller
}

const (
	msgIDInvalido    = "El parámetro 'id' es inválido o está ausente"
	msgMetodoNoExist = "Método de pago no encontrado"
	msgPagoNoExiste  = "Pago no encontrado"

	msgPedidoNoExiste = "Pedido no encontrado"
	msgSinProductos   = "El pedido no tiene productos: agréguelos antes de crear el pago"
	msgPedidoConPago  = "El pedido ya tiene un pago asignado"

	msgMontoPagado    = "No se puede modificar el monto de un pago ya PAGADO"
	msgMontoDescuento = "No se puede modificar el monto: el pedido del pago tiene un descuento aplicado y el monto ya lo refleja"
)

// normalizeEstado devuelve el estado en mayúsculas y si pertenece al enum.
func normalizeEstado(e string) (string, bool) {
	e = strings.ToUpper(strings.TrimSpace(e))
	switch e {
	case models.EstadoPagoPagado, models.EstadoPagoPendiente, models.EstadoPagoNoPago:
		return e, true
	}
	return e, false
}

// sameDay compara solo año, mes y día (UTC).
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

// firstNonNil devuelve el primer puntero no nulo (permite aceptar alias).
func firstNonNil(vals ...*string) *string {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

func (c *PagoController) fail(status int, msg string, err error) {
	httpx.Fail(&c.Controller, status, msg, err)
}

// writeError responde 409 si err es un conflicto de PostgreSQL y 500 en otro caso.
func (c *PagoController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "pagos."+op, err, nil)
	if httpx.IsPGConflict(err) {
		c.fail(http.StatusConflict, msg+": conflicto con datos existentes", err)
		return
	}
	c.fail(http.StatusInternalServerError, msg, err)
}

// metodoExists comprueba que el método de pago exista: (existe, errorDeBD).
func metodoExists(o orm.QueryExecutor, id int64) (bool, error) {
	err := o.Read(&models.MetodoPago{PK_ID_METODO_PAGO: id})
	if errors.Is(err, orm.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// parseFecha/parseHora validan los formatos de petición (YYYY-MM-DD y HH:MM[:SS]).
func parseFecha(s string) (time.Time, error) { return models.ParseDateToNoonUTC(strings.TrimSpace(s)) }

// parseHora valida HH:MM[:SS]. Devuelve la hora con año 2000 (no el año 1 de
// models.ParseTimeToUTC) para que FormatTimeWithLMT, que corrige años < 1900,
// no le sume 9:52:32 al serializar la respuesta.
func parseHora(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if len(s) != len("15:04:05") && len(s) != len("15:04") {
		return time.Time{}, errors.New("formato esperado HH:MM o HH:MM:SS")
	}
	t, err := models.ParseTimeToUTC(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(2000, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
}

// @Title GetAll
// @Summary Obtener todos los pagos con filtros
// @Description Devuelve los pagos, con filtros opcionales por fecha exacta, día, mes, año, estado y método de pago. El personal ve todos los pagos; un Cliente solo los de sus propios pedidos (los demás nunca aparecen). Cada pago incluye `metodoPagoId` como objeto (las relaciones embebidas solo garantizan su id). En la respuesta `fechaPago` va como DD-MM-YYYY, `horaPago` como HH:MM:SS y `updatedAt` como DD-MM-YYYY HH:MM:SS (Bogotá). Sin resultados, `data` es `[]` (HTTP 200).
// @Tags pagos
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Fecha exacta (YYYY-MM-DD)"
// @Param   dia      query   int      false   "Día del mes (1-31)"
// @Param   mes      query   int      false   "Mes (1-12)"
// @Param   anio     query   int      false   "Año (YYYY)"
// @Param   estado   query   string   false   "Estado del pago" Enums(PAGADO,PENDIENTE,NO_PAGO)
// @Param   metodo_pago     query   int      false   "ID del método de pago (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.PagoDoc} "Lista de pagos (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "Algún filtro tiene formato inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token de un Cliente no identifica a un cliente"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos [get]
func (c *PagoController) GetAll() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	qs := orm.NewOrm().QueryTable(new(models.Pago))
	if !claims.IsStaff() {
		// Un Cliente solo ve los pagos de sus propios pedidos. El documento viene del token (entero), no del usuario.
		doc, ok := authz.ResolveCliente(&c.Controller, claims, 0)
		if !ok {
			return
		}
		qs = qs.FilterRaw("PK_ID_PAGO", "IN (SELECT pk_id_pago FROM pedido WHERE pk_documento_cliente = "+strconv.FormatInt(doc, 10)+")")
	}

	var fecha time.Time
	hasFecha := false
	if raw := c.GetString("fecha"); raw != "" {
		f, err := parseFecha(raw)
		if err != nil {
			c.fail(http.StatusBadRequest, "Parámetro 'fecha' inválido (use YYYY-MM-DD)", err)
			return
		}
		fecha, hasFecha = f, true
	}
	// fecha/dia/mes/anio se filtran en memoria (el ORM no expone EXTRACT y la
	// columna es DATE).
	var dia, mes, anio int
	for _, p := range []struct {
		key      string
		dst      *int
		min, max int
	}{{"dia", &dia, 1, 31}, {"mes", &mes, 1, 12}, {"anio", &anio, 1, 9999}} {
		if c.GetString(p.key) == "" {
			continue
		}
		v, err := c.GetInt(p.key)
		if err != nil || v < p.min || v > p.max {
			c.fail(http.StatusBadRequest, "Parámetro '"+p.key+"' inválido", errors.New("valor fuera de rango o no numérico"))
			return
		}
		*p.dst = v
	}
	if estado := c.GetString("estado"); estado != "" {
		e, ok := normalizeEstado(estado)
		if !ok {
			c.fail(http.StatusBadRequest, "Parámetro 'estado' inválido (PAGADO, PENDIENTE o NO_PAGO)", nil)
			return
		}
		qs = qs.Filter("ESTADO_PAGO", e)
	}
	if c.GetString("metodo_pago") != "" {
		m, err := httpx.PositiveInt64Param(&c.Controller, "metodo_pago")
		if err != nil {
			c.fail(http.StatusBadRequest, "Parámetro 'metodo_pago' inválido", err)
			return
		}
		qs = qs.Filter("PK_ID_METODO_PAGO", m)
	}

	var pagos []models.Pago
	if _, err := qs.OrderBy("PK_ID_PAGO").All(&pagos); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.getall.db_error", err, nil)
		c.fail(http.StatusInternalServerError, "Error al obtener pagos de la base de datos", err)
		return
	}
	filtrados := make([]models.Pago, 0, len(pagos))
	for _, p := range pagos {
		if (hasFecha && !sameDay(p.FECHA, fecha)) ||
			(dia > 0 && p.FECHA.UTC().Day() != dia) ||
			(mes > 0 && int(p.FECHA.UTC().Month()) != mes) ||
			(anio > 0 && p.FECHA.UTC().Year() != anio) {
			continue
		}
		filtrados = append(filtrados, p)
	}
	httpx.Send(&c.Controller, http.StatusOK, "Pagos obtenidos exitosamente", filtrados)
}

// @Title GetById
// @Summary Obtener pago por ID
// @Description Devuelve un pago por ID. Un Cliente solo puede ver pagos de sus propios pedidos (otro pago responde 404, igual que si no existiera); el personal ve cualquiera. `fechaPago` va como DD-MM-YYYY y `horaPago` como HH:MM:SS.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del pago (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.PagoDoc} "Pago encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado (para un Cliente, también si no es de uno de sus pedidos)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos/search [get]
func (c *PagoController) GetById() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	pago := models.Pago{PK_ID_PAGO: id}
	if err := o.Read(&pago); err != nil {
		c.readError("getbyid", id, err)
		return
	}
	if !claims.IsStaff() {
		// Un pago ajeno (o sin pedido del cliente) responde igual que uno inexistente.
		n, err := o.QueryTable(new(models.Pedido)).Filter("PK_ID_PAGO", id).Filter("PK_DOCUMENTO_CLIENTE", claims.Documento).Count()
		if err != nil {
			c.readError("getbyid.owner", id, err)
			return
		}
		if claims.Documento <= 0 || n == 0 {
			c.fail(http.StatusNotFound, msgPagoNoExiste, nil)
			return
		}
	}
	httpx.Send(&c.Controller, http.StatusOK, "Pago encontrado", pago)
}

// readError traduce un fallo de Read: ErrNoRows -> 404, cualquier otro -> 500.
func (c *PagoController) readError(op string, id int64, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		c.fail(http.StatusNotFound, msgPagoNoExiste, nil)
		return
	}
	logging.LogControllerError(c.Ctx, "pagos."+op+".db_error", err, map[string]interface{}{"id": id})
	c.fail(http.StatusInternalServerError, "Error al consultar el pago", err)
}

// @Title Create
// @Summary Crear un nuevo pago
// @Description Lo pueden hacer el personal y un Cliente, con reglas distintas. CLIENTE: debe enviar `pedidoId` (de su propio pedido; ajeno o inexistente responde 404) y solo puede crear pagos PENDIENTE (403 con otro estado); el pago se crea y se LIGA al pedido en la MISMA transacción (con el pedido bloqueado), así nunca queda un pago huérfano que otro cliente pueda adivinar y asignar. El `monto` del cuerpo se IGNORA (puede omitirse): el pago nace con el monto calculado por el servidor, que se devuelve en la respuesta; responde 409 si el pedido no tiene productos o ya tiene un pago (no se crea nada). Para comprar de una sola vez use `POST /pedidos/checkout`. PERSONAL: el pago NO se liga al pedido (eso se hace con `POST /pedidos/asignar-pago`); puede indicar `pedidoId` y dejar `monto` en 0/omitido para usar el calculado, o fijar un `monto` manual > 0 (ajustes de mostrador; negativo responde 400); sin `pedidoId` el `monto` manual es obligatorio y debe ser > 0. Campos obligatorios: `fechaPago` YYYY-MM-DD, `horaPago` HH:MM[:SS], `estadoPago` PAGADO|PENDIENTE|NO_PAGO y `metodoPagoId` de un método existente (404 si no existe); `updatedBy` es opcional. EL SERVIDOR MANDA EL MONTO. Fórmula: `monto = MAX(0, SUM(detalle_pedido.precio x cantidad) - descuentos ya aplicados al pedido)`; no existen cargos de domicilio ni propina. La respuesta devuelve `fechaPago` como DD-MM-YYYY.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   body  body   models.PagoCreateRequest true  "Datos del pago a crear"
// @Success 201 {object} models.ApiResponse{data=models.PagoDoc} "Pago creado"
// @Failure 400 {object} models.ApiResponse "JSON inválido o campos ausentes/inválidos (un Cliente sin `pedidoId`; personal con `monto` negativo o sin `pedidoId` y `monto` <= 0)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente intentó crear un pago que no es PENDIENTE"
// @Failure 404 {object} models.ApiResponse "El método de pago o el pedido indicado no existe (para un Cliente, también si el pedido es ajeno)"
// @Failure 409 {object} models.ApiResponse "El pedido no tiene productos o (Cliente) ya tiene un pago asignado, o conflicto de unicidad en base de datos"
// @Failure 500 {object} models.ApiResponse "Error al crear el pago"
// @Security BearerAuth
// @Router /pagos [post]
func (c *PagoController) Post() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	staff := claims.IsStaff()
	var in models.PagoCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Error al decodificar la solicitud", err)
		return
	}
	if in.FechaPago == "" {
		c.fail(http.StatusBadRequest, "El campo fechaPago no puede estar vacío", nil)
		return
	}
	fecha, err := parseFecha(in.FechaPago)
	if err != nil {
		c.fail(http.StatusBadRequest, "Formato de fecha inválido (use YYYY-MM-DD)", err)
		return
	}
	if in.HoraPago == "" {
		c.fail(http.StatusBadRequest, "El campo horaPago no puede estar vacío", nil)
		return
	}
	hora, err := parseHora(in.HoraPago)
	if err != nil {
		c.fail(http.StatusBadRequest, "Formato de hora inválido, debe ser HH:mm:ss", err)
		return
	}
	switch {
	case !staff && in.PedidoId <= 0:
		c.fail(http.StatusBadRequest, "El campo pedidoId es obligatorio para un cliente: el monto lo calcula el servidor desde el pedido", nil)
		return
	case staff && in.Monto < 0:
		c.fail(http.StatusBadRequest, "El campo monto no puede ser negativo", nil)
		return
	case staff && in.PedidoId <= 0 && in.Monto == 0:
		c.fail(http.StatusBadRequest, "El campo monto es obligatorio y debe ser un entero mayor que 0 (o indique pedidoId para usar el monto calculado)", nil)
		return
	}
	if in.EstadoPago == "" {
		c.fail(http.StatusBadRequest, "El campo estadoPago es obligatorio", nil)
		return
	}
	estado, ok := normalizeEstado(in.EstadoPago)
	if !ok {
		c.fail(http.StatusBadRequest, "Estado de pago inválido", errors.New("el estado debe ser 'PAGADO', 'PENDIENTE' o 'NO_PAGO'"))
		return
	}
	if in.MetodoPagoId <= 0 {
		c.fail(http.StatusBadRequest, "El campo metodoPagoId es obligatorio y debe ser un número válido", nil)
		return
	}

	if !staff && estado != models.EstadoPagoPendiente {
		c.fail(http.StatusForbidden, "Un cliente solo puede crear pagos en estado PENDIENTE", nil)
		return
	}

	o := orm.NewOrm()
	if ok, err := metodoExists(o, in.MetodoPagoId); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.metodo_error", err, map[string]interface{}{"metodoPagoId": in.MetodoPagoId})
		c.fail(http.StatusInternalServerError, "Error al validar el método de pago", err)
		return
	} else if !ok {
		c.fail(http.StatusNotFound, msgMetodoNoExist, nil)
		return
	}

	pago := models.Pago{
		FECHA:             fecha,
		HORA:              hora,
		ESTADO_PAGO:       estado,
		PK_ID_METODO_PAGO: &models.MetodoPago{PK_ID_METODO_PAGO: in.MetodoPagoId},
		UPDATED_AT:        time.Now().UTC(),
	}
	if in.UpdatedBy != "" {
		pago.UPDATED_BY = &in.UpdatedBy
	}
	if !staff {
		if c.crearPagoLigado(o, claims, in.PedidoId, &pago) {
			httpx.Send(&c.Controller, http.StatusCreated, "Pago creado correctamente", pago)
		}
		return
	}

	monto := in.Monto
	if in.PedidoId > 0 {
		calculado, ok := c.montoDelPedido(o, in.PedidoId)
		if !ok {
			return
		}
		// El personal puede fijar un monto manual (ajuste de mostrador); sin él rige el calculado.
		if in.Monto == 0 {
			monto = calculado
		}
	}
	pago.MONTO = monto
	if _, err := o.Insert(&pago); err != nil {
		c.writeError("post.insert_error", "Error al crear el pago", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Pago creado correctamente", pago)
}

// crearPagoLigado crea el pago de un Cliente y lo liga a su pedido en UNA
// transacción, con el pedido bloqueado (FOR UPDATE): 404 si el pedido no existe
// o es ajeno, 409 si ya tiene pago o no tiene productos, 500 si falla la base
// de datos. El monto sale siempre del servidor. Si responde un error no queda
// nada escrito y devuelve false.
func (c *PagoController) crearPagoLigado(o orm.Ormer, claims *login.Claims, pedidoID int64, pago *models.Pago) bool {
	ctx := map[string]interface{}{"pedidoId": pedidoID}
	tx, err := o.Begin()
	if err != nil {
		c.writeError("post.tx_begin_error", "Error al crear el pago", err)
		return false
	}
	abortar := func(status int, msg string, err error) bool {
		_ = tx.Rollback()
		if status == http.StatusInternalServerError {
			logging.LogControllerError(c.Ctx, "pagos.post.ligar_error", err, ctx)
		}
		c.fail(status, msg, err)
		return false
	}
	var cliente, pagoActual int64
	err = tx.Raw("SELECT COALESCE(pk_documento_cliente, 0), COALESCE(pk_id_pago, 0) FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", pedidoID).QueryRow(&cliente, &pagoActual)
	switch {
	case errors.Is(err, orm.ErrNoRows) || (err == nil && !authz.EsDuenio(claims, cliente)):
		return abortar(http.StatusNotFound, msgPedidoNoExiste, nil)
	case err != nil:
		return abortar(http.StatusInternalServerError, "Error al consultar el pedido", err)
	case pagoActual > 0:
		return abortar(http.StatusConflict, msgPedidoConPago, nil)
	}
	m, err := montopedido.Calcular(tx, pedidoID)
	if err != nil {
		return abortar(http.StatusInternalServerError, "Error al calcular el monto del pedido", err)
	}
	if m.Lineas == 0 {
		return abortar(http.StatusConflict, msgSinProductos, nil)
	}
	pago.MONTO = m.Total
	if _, err := tx.Insert(pago); err != nil {
		_ = tx.Rollback()
		c.writeError("post.insert_error", "Error al crear el pago", err)
		return false
	}
	if _, err := tx.Raw("UPDATE pedido SET pk_id_pago = ?, updated_at = ? WHERE pk_id_pedido = ?", pago.PK_ID_PAGO, pago.UPDATED_AT, pedidoID).Exec(); err != nil {
		return abortar(http.StatusInternalServerError, "Error al ligar el pago al pedido", err)
	}
	if err := tx.Commit(); err != nil {
		c.writeError("post.tx_commit_error", "Error al crear el pago", err)
		return false
	}
	return true
}

// montoDelPedido calcula en el servidor el monto del pedido (ver montopedido).
// Responde 404 si el pedido no existe, 409 si no tiene productos y 500 si falla
// la base de datos. Devuelve false si ya respondió.
func (c *PagoController) montoDelPedido(o orm.QueryExecutor, pedidoID int64) (int64, bool) {
	ctx := map[string]interface{}{"pedidoId": pedidoID}
	var id int64
	err := o.Raw("SELECT pk_id_pedido FROM pedido WHERE pk_id_pedido = ?", pedidoID).QueryRow(&id)
	if errors.Is(err, orm.ErrNoRows) {
		c.fail(http.StatusNotFound, msgPedidoNoExiste, nil)
		return 0, false
	}
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.pedido_error", err, ctx)
		c.fail(http.StatusInternalServerError, "Error al consultar el pedido", err)
		return 0, false
	}
	m, err := montopedido.Calcular(o, pedidoID)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.monto_error", err, ctx)
		c.fail(http.StatusInternalServerError, "Error al calcular el monto del pedido", err)
		return 0, false
	}
	if m.Lineas == 0 {
		c.fail(http.StatusConflict, msgSinProductos, nil)
		return 0, false
	}
	return m.Total, true
}

// montoModificable comprueba que el monto del pago pueda cambiar: un pago ya
// PAGADO no admite cambios y tampoco el de un pedido con descuentos aplicados
// (AplicarDescuento ya restó el descuento de `monto`, sobrescribirlo lo borraría).
// Responde 409 (o 500 si falla la consulta) y devuelve false si no se puede.
func (c *PagoController) montoModificable(tx orm.QueryExecutor, pago *models.Pago) bool {
	if pago.ESTADO_PAGO == models.EstadoPagoPagado {
		c.fail(http.StatusConflict, msgMontoPagado, nil)
		return false
	}
	var n int64
	err := tx.Raw(`SELECT COUNT(*) FROM pedido_descuento_aplicado d JOIN pedido p ON p.pk_id_pedido = d.pk_id_pedido WHERE p.pk_id_pago = ?`, pago.PK_ID_PAGO).QueryRow(&n)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.put.descuento_error", err, map[string]interface{}{"id": pago.PK_ID_PAGO})
		c.fail(http.StatusInternalServerError, "Error al validar los descuentos del pedido", err)
		return false
	}
	if n > 0 {
		c.fail(http.StatusConflict, msgMontoDescuento, nil)
		return false
	}
	return true
}

// @Title Update
// @Summary Actualizar un pago
// @Description Solo personal (trabajador o administrador): un Cliente recibe 403. Regla del `monto`: no se puede cambiar el monto de un pago ya PAGADO (409) ni el de un pago cuyo pedido tiene descuentos aplicados (409, porque el monto ya refleja el descuento y se perdería); enviar el mismo monto que ya tiene no es un cambio. Actualización parcial (merge): los campos ausentes del cuerpo se conservan; el cuerpo puede ser parcial (incluso `{}`). Claves: `fechaPago` (YYYY-MM-DD), `horaPago` (HH:MM[:SS]), `monto` (entero > 0), `estadoPago`, `metodoPagoId` (debe existir, 404 si no) y `updatedBy`. Por compatibilidad se aceptan también `fecha` y `hora` como alias. Solo `updatedBy` es anulable (null lo limpia); null en cualquier otro campo responde 400.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del pago (entero positivo)"
// @Param   body  body   models.PagoUpdateRequest true  "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.PagoDoc} "Pago actualizado"
// @Failure 400 {object} models.ApiResponse "id inválido, JSON inválido, null en campo no anulable o valores inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere un usuario trabajador"
// @Failure 404 {object} models.ApiResponse "Pago o método de pago no encontrado"
// @Failure 409 {object} models.ApiResponse "El pago ya está PAGADO o su pedido tiene descuentos aplicados (no se puede cambiar `monto`), o conflicto de unicidad en base de datos"
// @Failure 500 {object} models.ApiResponse "Error al actualizar el pago"
// @Security BearerAuth
// @Router /pagos [put]
func (c *PagoController) Put() {
	if _, ok := authz.RequireStaff(&c.Controller); !ok {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	// El pago se lee y se actualiza en una transacción con el pago bloqueado, igual que
	// AplicarDescuento: así no se puede pisar el monto que un descuento acaba de rebajar.
	tx, err := orm.NewOrm().Begin()
	if err != nil {
		c.writeError("put.tx_begin_error", "Error al actualizar el pago", err)
		return
	}
	confirmada := false
	defer func() {
		if !confirmada {
			_ = tx.Rollback()
		}
	}()
	var pago models.Pago
	if err := tx.QueryTable(new(models.Pago)).Filter("PK_ID_PAGO", id).ForUpdate().One(&pago); err != nil {
		c.readError("put", id, err)
		return
	}

	var in models.PagoUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in, "updatedBy"); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.put.bad_json", err, map[string]interface{}{"id": id})
		c.fail(http.StatusBadRequest, "Error al decodificar la solicitud", err)
		return
	}

	// Solo se actualizan las columnas enviadas (no se reescribe el resto).
	cols := []string{"UPDATED_AT"}
	if fs := firstNonNil(in.FechaPago, in.Fecha); fs != nil {
		f, err := parseFecha(*fs)
		if err != nil {
			c.fail(http.StatusBadRequest, "Formato de fecha inválido (YYYY-MM-DD)", err)
			return
		}
		pago.FECHA = f
		cols = append(cols, "FECHA")
	}
	if hs := firstNonNil(in.HoraPago, in.Hora); hs != nil {
		h, err := parseHora(*hs)
		if err != nil {
			c.fail(http.StatusBadRequest, "Formato de hora inválido (HH:MM[:SS])", err)
			return
		}
		pago.HORA = h
		cols = append(cols, "HORA")
	}
	if in.Monto != nil {
		if *in.Monto <= 0 {
			c.fail(http.StatusBadRequest, "El campo monto debe ser un entero mayor que 0", nil)
			return
		}
		if *in.Monto != pago.MONTO {
			if !c.montoModificable(tx, &pago) {
				return
			}
			pago.MONTO = *in.Monto
			cols = append(cols, "MONTO")
		}
	}
	if in.EstadoPago != nil {
		estado, ok := normalizeEstado(*in.EstadoPago)
		if !ok {
			c.fail(http.StatusBadRequest, "Estado de pago inválido. Debe ser 'PAGADO', 'PENDIENTE' o 'NO_PAGO'", nil)
			return
		}
		pago.ESTADO_PAGO = estado
		cols = append(cols, "ESTADO_PAGO")
	}
	if in.MetodoPagoId != nil {
		if *in.MetodoPagoId <= 0 {
			c.fail(http.StatusBadRequest, "El campo metodoPagoId debe ser un número válido", nil)
			return
		}
		ok, err := metodoExists(tx, *in.MetodoPagoId)
		if err != nil {
			logging.LogControllerError(c.Ctx, "pagos.put.metodo_error", err, map[string]interface{}{"id": id})
			c.fail(http.StatusInternalServerError, "Error al validar el método de pago", err)
			return
		}
		if !ok {
			c.fail(http.StatusNotFound, msgMetodoNoExist, nil)
			return
		}
		pago.PK_ID_METODO_PAGO = &models.MetodoPago{PK_ID_METODO_PAGO: *in.MetodoPagoId}
		cols = append(cols, "PK_ID_METODO_PAGO")
	}
	if in.UpdatedBy != nil {
		pago.UPDATED_BY = in.UpdatedBy
		cols = append(cols, "UPDATED_BY")
	} else if httpx.IsNull(c.Ctx.Input.RequestBody, "updatedBy") {
		pago.UPDATED_BY = nil
		cols = append(cols, "UPDATED_BY")
	}
	pago.UPDATED_AT = time.Now().UTC()

	if _, err := tx.Update(&pago, cols...); err != nil {
		c.writeError("put.update_error", "Error al actualizar el pago", err)
		return
	}
	if err := tx.Commit(); err != nil {
		c.writeError("put.tx_commit_error", "Error al actualizar el pago", err)
		return
	}
	confirmada = true
	httpx.Send(&c.Controller, http.StatusOK, "Pago actualizado correctamente", pago)
}

// @Title Delete
// @Summary Eliminar un pago
// @Description Solo personal (trabajador o administrador): un Cliente recibe 403. Elimina un pago. Si está asociado a un pedido responde 409.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del pago (entero positivo)"
// @Success 200 {object} models.ApiResponse "Pago eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere un usuario trabajador"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Failure 409 {object} models.ApiResponse "El pago está asociado a un pedido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos [delete]
func (c *PagoController) Delete() {
	if _, ok := authz.RequireStaff(&c.Controller); !ok {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	n, err := orm.NewOrm().Delete(&models.Pago{PK_ID_PAGO: id})
	if err != nil {
		c.writeError("delete.delete_error", "Error al eliminar el pago", err)
		return
	}
	if n == 0 {
		c.fail(http.StatusNotFound, msgPagoNoExiste, nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Pago eliminado", nil)
}
