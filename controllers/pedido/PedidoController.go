package pedido

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"restaurante/controllers/login"
	"restaurante/database"
	"restaurante/internal/authz"
	"restaurante/internal/httpx"
	"restaurante/internal/montopedido"
	"restaurante/internal/notify"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// loadLocation carga la zona horaria; es una variable para poder simular su fallo en tests.
var loadLocation = time.LoadLocation

type PedidoController struct {
	web.Controller
}

const (
	msgPedidoNoEncontrado    = "Pedido no encontrado"
	msgPedidoIDInvalido      = "El parámetro 'pedido_id' es obligatorio y debe ser un entero positivo"
	msgDomicilioNoEncontrado = "Domicilio no encontrado"
	msgPagoNoEncontrado      = "Pago no encontrado"
	layoutFecha              = "2006-01-02"
)

func (c *PedidoController) fail(status int, msg string, err error) {
	httpx.Fail(&c.Controller, status, msg, err)
}

// dbError registra el error y responde 409 si es un conflicto de PostgreSQL
// (unicidad o llave foránea) y 500 en cualquier otro caso.
func (c *PedidoController) dbError(op, msg string, err error, ctx map[string]interface{}) {
	logging.LogControllerError(c.Ctx, "pedidos."+op, err, ctx)
	if httpx.IsPGConflict(err) {
		c.fail(http.StatusConflict, msg+": conflicto con datos existentes", err)
		return
	}
	c.fail(http.StatusInternalServerError, msg, err)
}

// exists indica si hay una fila de model con column = id. Un error de base de
// datos se devuelve tal cual (nunca se confunde con "no existe").
func exists(o orm.Ormer, model interface{}, column string, id int64) (bool, error) {
	n, err := o.QueryTable(model).Filter(column, id).Count()
	return n > 0, err
}

// existsOrFail comprueba que exista la fila referenciada; si no existe responde
// 404 con notFoundMsg y si la consulta falla responde 500. Devuelve true si se
// puede continuar.
func (c *PedidoController) existsOrFail(o orm.Ormer, op string, model interface{}, column string, id int64, notFoundMsg string) bool {
	ok, err := exists(o, model, column, id)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos."+op+".exists_error", err, map[string]interface{}{"id": id})
		c.fail(http.StatusInternalServerError, "Error al validar los datos referenciados", err)
		return false
	}
	if !ok {
		c.fail(http.StatusNotFound, notFoundMsg, nil)
		return false
	}
	return true
}

// enOtroPedido indica (y responde 404 con notFoundMsg, o 500 si falla la
// consulta) si el recurso column = id ya pertenece a un pedido distinto de
// pedidoID. Se usa para que un Cliente no se apropie de pagos o domicilios
// ajenos; devuelve true si ya respondió.
func (c *PedidoController) enOtroPedido(o orm.Ormer, op, column string, id, pedidoID int64, notFoundMsg string) bool {
	n, err := o.QueryTable(new(models.Pedido)).Filter(column, id).Exclude("PK_ID_PEDIDO", pedidoID).Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos."+op+".owner_error", err, map[string]interface{}{"id": id})
		c.fail(http.StatusInternalServerError, "Error al validar los datos referenciados", err)
		return true
	}
	if n > 0 {
		c.fail(http.StatusNotFound, notFoundMsg, nil)
		return true
	}
	return false
}

// readPedido lee el pedido pedido_id (query param) y responde 400/404/500 si
// no se puede. Un Cliente solo accede a sus pedidos: uno ajeno responde igual
// que si no existiera (404). Devuelve nil si ya respondió.
func (c *PedidoController) readPedido(op string, o orm.Ormer, claims *login.Claims) *models.Pedido {
	id, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos."+op+".bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.fail(http.StatusBadRequest, msgPedidoIDInvalido, err)
		return nil
	}
	pedido := models.Pedido{PK_ID_PEDIDO: id}
	if err := o.Read(&pedido); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return nil
		}
		logging.LogControllerError(c.Ctx, "pedidos."+op+".read_error", err, map[string]interface{}{"pedido_id": id})
		c.fail(http.StatusInternalServerError, "Error al consultar el pedido", err)
		return nil
	}
	if !authz.EsDuenio(claims, clienteDe(&pedido)) {
		c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
		return nil
	}
	return &pedido
}

// clienteDe devuelve el documento del cliente del pedido (0 si no tiene).
func clienteDe(p *models.Pedido) int64 {
	if p.PK_DOCUMENTO_CLIENTE == nil {
		return 0
	}
	return p.PK_DOCUMENTO_CLIENTE.PK_DOCUMENTO_CLIENTE
}

// domicilioDe devuelve el id del domicilio del pedido (0 si no tiene).
func domicilioDe(p *models.Pedido) int64 {
	if p.PK_ID_DOMICILIO == nil {
		return 0
	}
	return p.PK_ID_DOMICILIO.ID
}

// notificarEstado avisa al cliente cuando el estado del pedido cambió.
func notificarEstado(p *models.Pedido, anterior models.EstadoPedido) {
	if p.ESTADO_PEDIDO != anterior {
		notify.Enviar(notify.Evento{Tipo: notify.PedidoEstado, PedidoID: p.PK_ID_PEDIDO, Cliente: clienteDe(p), Estado: p.ESTADO_PEDIDO})
	}
}

// parseFechaParam valida un filtro de fecha YYYY-MM-DD (vacío = ausente).
func parseFechaParam(raw string) error {
	if raw == "" {
		return nil
	}
	_, err := time.Parse(layoutFecha, raw)
	return err
}

// intParam lee un entero opcional en [min, max]; present=false si no vino.
func (c *PedidoController) intParam(key string, min, max int) (v int, present bool, err error) {
	raw := strings.TrimSpace(c.GetString(key))
	if raw == "" {
		return 0, false, nil
	}
	v, err = strconv.Atoi(raw)
	if err != nil || v < min || v > max {
		return 0, true, errors.New("el parámetro '" + key + "' debe ser un entero entre " + strconv.Itoa(min) + " y " + strconv.Itoa(max))
	}
	return v, true, nil
}

// @Title GetAll
// @Summary Obtener pedidos con múltiples filtros
// @Description Devuelve pedidos filtrados por fecha, rango de fechas (`desde` y `hasta` solo se aplican juntos), mes/año, cliente, tipo de método de pago y si tienen domicilio. Todos los filtros son opcionales; un filtro con formato inválido responde 400. Sin resultados responde 200 con `data` igual a `[]`. Un Cliente solo recibe sus propios pedidos (el filtro `cliente` se fuerza al documento de su token); el personal ve todos. En cada pedido `fechaPedido` va como DD-MM-YYYY, `horaPedido` como HH:MM:SS y las relaciones (`pagoId`, `domicilioId`, `restauranteId`, `documentoCliente`) como objetos en los que solo el id es fiable.
// @Tags pedido
// @Accept json
// @Produce json
// @Param fecha query string false "Fecha específica (YYYY-MM-DD)"
// @Param desde query string false "Fecha inicial del rango (YYYY-MM-DD); requiere `hasta`"
// @Param hasta query string false "Fecha final del rango (YYYY-MM-DD); requiere `desde`"
// @Param mes query int false "Mes del año (1-12)"
// @Param anio query int false "Año (YYYY); se combina con `mes` si ambos se envían"
// @Param cliente query int false "Documento del cliente (entero positivo). Un Cliente solo puede consultar el suyo (se aplica por defecto); otro documento responde 403"
// @Param metodo_pago query string false "Tipo de método de pago (p. ej. NEQUI, DAVIPLATA, EFECTIVO); sin distinguir mayúsculas"
// @Param domicilio query bool false "true: solo pedidos con domicilio; false: solo pedidos sin domicilio"
// @Success 200 {object} models.ApiResponse{data=[]models.PedidoDoc} "Pedidos obtenidos (puede ser lista vacía)"
// @Failure 400 {object} models.ApiResponse "Algún filtro tiene formato inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente pidió los pedidos de otro cliente"
// @Failure 500 {object} models.ApiResponse "Error al obtener los pedidos"
// @Security BearerAuth
// @Router /pedidos [get]
func (c *PedidoController) GetAll() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	query := `
       SELECT p.*
       FROM pedido p
       LEFT JOIN pago pa ON p.pk_id_pago = pa.pk_id_pago
       LEFT JOIN metodo_pago mp ON pa.pk_id_metodo_pago = mp.pk_id_metodo_pago
       WHERE 1 = 1
   `
	params := []interface{}{}

	fecha := strings.TrimSpace(c.GetString("fecha"))
	desde := strings.TrimSpace(c.GetString("desde"))
	hasta := strings.TrimSpace(c.GetString("hasta"))
	for _, f := range []struct{ name, val string }{{"fecha", fecha}, {"desde", desde}, {"hasta", hasta}} {
		if err := parseFechaParam(f.val); err != nil {
			c.fail(http.StatusBadRequest, "Parámetro '"+f.name+"' inválido (use YYYY-MM-DD)", err)
			return
		}
	}
	if (desde == "") != (hasta == "") {
		c.fail(http.StatusBadRequest, "Los parámetros 'desde' y 'hasta' deben enviarse juntos", nil)
		return
	}
	if desde > hasta {
		c.fail(http.StatusBadRequest, "El parámetro 'desde' no puede ser posterior a 'hasta'", nil)
		return
	}
	mes, hasMes, err := c.intParam("mes", 1, 12)
	if err != nil {
		c.fail(http.StatusBadRequest, "Parámetro 'mes' inválido", err)
		return
	}
	anio, hasAnio, err := c.intParam("anio", 1, 9999)
	if err != nil {
		c.fail(http.StatusBadRequest, "Parámetro 'anio' inválido", err)
		return
	}
	var cliente int64
	hasCliente := strings.TrimSpace(c.GetString("cliente")) != ""
	if hasCliente {
		if cliente, err = httpx.PositiveInt64Param(&c.Controller, "cliente"); err != nil {
			c.fail(http.StatusBadRequest, "Parámetro 'cliente' inválido", err)
			return
		}
	}
	if !claims.IsStaff() {
		// Un Cliente solo lista sus pedidos: el documento del token manda y otro distinto es 403.
		if cliente, ok = authz.ResolveCliente(&c.Controller, claims, cliente); !ok {
			return
		}
		hasCliente = true
	}
	var domicilio bool
	hasDomicilio := strings.TrimSpace(c.GetString("domicilio")) != ""
	if hasDomicilio {
		if domicilio, err = strconv.ParseBool(strings.TrimSpace(c.GetString("domicilio"))); err != nil {
			c.fail(http.StatusBadRequest, "Parámetro 'domicilio' inválido (true o false)", err)
			return
		}
	}
	metodoPago := strings.TrimSpace(c.GetString("metodo_pago"))

	if fecha != "" {
		query += ` AND p.fecha = ?`
		params = append(params, fecha)
	}
	if desde != "" {
		query += ` AND p.fecha BETWEEN ? AND ?`
		params = append(params, desde, hasta)
	}
	if hasMes {
		query += ` AND EXTRACT(MONTH FROM p.fecha) = ?`
		params = append(params, mes)
	}
	if hasAnio {
		query += ` AND EXTRACT(YEAR FROM p.fecha) = ?`
		params = append(params, anio)
	}
	if hasCliente {
		query += ` AND p.pk_documento_cliente = ?`
		params = append(params, cliente)
	}
	if metodoPago != "" {
		query += ` AND mp.tipo ILIKE ?`
		params = append(params, metodoPago)
	}
	if hasDomicilio {
		if domicilio {
			query += ` AND p.pk_id_domicilio IS NOT NULL`
		} else {
			query += ` AND p.pk_id_domicilio IS NULL`
		}
	}
	query += ` ORDER BY p.pk_id_pedido`

	var pedidos []models.Pedido
	if _, err := orm.NewOrm().Raw(query, params...).QueryRows(&pedidos); err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.getall.db_error", err, map[string]interface{}{
			"fecha": fecha, "desde": desde, "hasta": hasta, "mes": mes, "anio": anio, "cliente": cliente, "metodo_pago": metodoPago, "domicilio": domicilio,
		})
		c.fail(http.StatusInternalServerError, "Error al obtener los pedidos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Pedidos obtenidos exitosamente", httpx.List(pedidos))
}

// @Title PostPedido
// @Summary Crear un nuevo pedido
// @Description Crea un pedido. El servidor fija `fechaPedido`/`horaPedido` (Bogotá) y `estadoPedido`=INICIADO. Todos los campos del cuerpo son opcionales: `delivery` (por defecto false; si es true exige `pk_id_domicilio`), `pk_id_domicilio`, `restauranteId` y `documentoCliente` (si se envían deben ser enteros positivos de filas existentes: 404 si no existen). Quién crea: un Cliente siempre crea a su nombre (el documento sale del token; `documentoCliente` en el body se ignora si coincide y responde 403 si es otro); el personal (trabajador o administrador) puede crear a nombre de un cliente enviando `documentoCliente` (404 si el cliente no existe) o sin cliente (pedido de mostrador, la columna admite NULL). Responde 201 con el pedido creado. Envía en segundo plano (best-effort, sin afectar la respuesta) un push de confirmación al cliente (si tiene `documentoCliente`) y un aviso a los trabajadores.
// @Tags pedido
// @Accept json
// @Produce json
// @Param body body models.PedidoCreateRequest true "Datos del pedido"
// @Success 201 {object} models.ApiResponse{data=models.PedidoDoc} "Pedido creado"
// @Failure 400 {object} models.ApiResponse "JSON inválido, ids no positivos o `delivery` sin domicilio"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente envió un `documentoCliente` distinto al de su token"
// @Failure 404 {object} models.ApiResponse "El domicilio, el restaurante o el cliente indicado no existe"
// @Failure 409 {object} models.ApiResponse "Conflicto con datos existentes (p. ej. el domicilio ya pertenece a otro pedido)"
// @Failure 500 {object} models.ApiResponse "Error al crear el pedido"
// @Security BearerAuth
// @Router /pedidos [post]
func (c *PedidoController) Post() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	var in models.PedidoCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Datos inválidos (JSON mal formado)", err)
		return
	}
	if (in.PKIDDomicilio != nil && *in.PKIDDomicilio <= 0) ||
		(in.DocumentoCliente != nil && *in.DocumentoCliente <= 0) || in.RestauranteId < 0 {
		c.fail(http.StatusBadRequest, "pk_id_domicilio, restauranteId y documentoCliente deben ser enteros positivos", nil)
		return
	}
	if !claims.IsStaff() {
		// Un Cliente siempre crea a su nombre: el documento sale del token y otro distinto en el body es 403.
		var body int64
		if in.DocumentoCliente != nil {
			body = *in.DocumentoCliente
		}
		doc, ok := authz.ResolveCliente(&c.Controller, claims, body)
		if !ok {
			return
		}
		in.DocumentoCliente = &doc
	}
	delivery := in.Delivery != nil && *in.Delivery
	if delivery && in.PKIDDomicilio == nil {
		logging.LogControllerError(c.Ctx, "pedidos.post.delivery_missing_domicilio", nil, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "El domicilio es obligatorio cuando delivery es true", nil)
		return
	}

	bogota := database.BogotaZone
	if bogota == nil {
		if loc, err := loadLocation("America/Bogota"); err == nil {
			bogota = loc
		} else {
			bogota = time.FixedZone("UTC-5", -5*60*60)
		}
	}
	nowBogota := time.Now().In(bogota)
	pedido := models.Pedido{
		FECHA:         time.Date(nowBogota.Year(), nowBogota.Month(), nowBogota.Day(), 12, 0, 0, 0, time.UTC),
		HORA:          time.Now().UTC(),
		DELIVERY:      delivery,
		ESTADO_PEDIDO: models.EstadoPedidoIniciado,
	}

	o := orm.NewOrm()
	if in.PKIDDomicilio != nil {
		if !c.existsOrFail(o, "post", new(models.Domicilio), "ID", *in.PKIDDomicilio, msgDomicilioNoEncontrado) {
			return
		}
		pedido.PK_ID_DOMICILIO = &models.Domicilio{ID: *in.PKIDDomicilio}
	}
	if in.RestauranteId > 0 {
		if !c.existsOrFail(o, "post", new(models.Restaurante), "PK_ID_RESTAURANTE", in.RestauranteId, "Restaurante no encontrado") {
			return
		}
		pedido.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: in.RestauranteId}
	}
	if in.DocumentoCliente != nil {
		if !c.existsOrFail(o, "post", new(models.Cliente), "PK_DOCUMENTO_CLIENTE", *in.DocumentoCliente, "Cliente no encontrado") {
			return
		}
		pedido.PK_DOCUMENTO_CLIENTE = &models.Cliente{PK_DOCUMENTO_CLIENTE: *in.DocumentoCliente}
	}

	if _, err := o.Insert(&pedido); err != nil {
		c.dbError("post.insert_error", "Error al crear el pedido", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		return
	}
	notify.Enviar(notify.Evento{Tipo: notify.PedidoCreado, PedidoID: pedido.PK_ID_PEDIDO, Cliente: clienteDe(&pedido), DomicilioID: domicilioDe(&pedido)})
	httpx.Send(&c.Controller, http.StatusCreated, "Pedido creado exitosamente", pedido)
}

// @Title AssignDomicilio
// @Summary Asignar un domicilio a un pedido
// @Description Asigna un domicilio existente a un pedido y marca `delivery`=true. Responde con el pedido completo actualizado. Un Cliente solo puede hacerlo sobre su propio pedido (uno ajeno responde 404, como si no existiera) y con un domicilio que no pertenezca a otro pedido (404); el personal puede con cualquiera. Avisa por push a los trabajadores (best-effort, en segundo plano).
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param domicilio_id query int true "ID del domicilio (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.PedidoDoc} "Domicilio asignado al pedido"
// @Failure 400 {object} models.ApiResponse "pedido_id o domicilio_id inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido o domicilio no encontrado (para un Cliente, también si son de otro cliente)"
// @Failure 409 {object} models.ApiResponse "Conflicto con datos existentes"
// @Failure 500 {object} models.ApiResponse "Error al asignar domicilio"
// @Security BearerAuth
// @Router /pedidos/asignar-domicilio [post]
func (c *PedidoController) AssignDomicilio() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	domicilioID, err := httpx.PositiveInt64Param(&c.Controller, "domicilio_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.assign_domicilio.bad_request", err, map[string]interface{}{"domicilio_id": c.GetString("domicilio_id")})
		c.fail(http.StatusBadRequest, "El parámetro 'domicilio_id' es obligatorio y debe ser un entero positivo", err)
		return
	}
	o := orm.NewOrm()
	pedido := c.readPedido("assign_domicilio", o, claims)
	if pedido == nil {
		return
	}
	if !c.existsOrFail(o, "assign_domicilio", new(models.Domicilio), "ID", domicilioID, msgDomicilioNoEncontrado) {
		return
	}
	if !claims.IsStaff() && c.enOtroPedido(o, "assign_domicilio", "PK_ID_DOMICILIO", domicilioID, pedido.PK_ID_PEDIDO, msgDomicilioNoEncontrado) {
		return
	}
	pedido.PK_ID_DOMICILIO = &models.Domicilio{ID: domicilioID}
	pedido.DELIVERY = true
	pedido.UPDATED_AT = time.Now().UTC()
	if _, err := o.Update(pedido, "PK_ID_DOMICILIO", "DELIVERY", "UPDATED_AT"); err != nil {
		c.dbError("assign_domicilio.update_error", "Error al asignar domicilio", err, map[string]interface{}{"pedido_id": pedido.PK_ID_PEDIDO, "domicilio_id": domicilioID})
		return
	}
	notify.Enviar(notify.Evento{Tipo: notify.PedidoDomicilio, PedidoID: pedido.PK_ID_PEDIDO, DomicilioID: domicilioID})
	httpx.Send(&c.Controller, http.StatusOK, "Domicilio asignado correctamente", *pedido)
}

// fijarMontoCliente prepara la asignación de un pago por un Cliente dentro de tx:
// bloquea el pedido, exige que aún no tenga pago (409) y productos (409), y deja
// en el pago PENDIENTE el monto calculado por el servidor (409 si el pago no está
// PENDIENTE). Si responde un error deshace tx y devuelve false.
func (c *PedidoController) fijarMontoCliente(tx orm.TxOrmer, pedidoID, pagoID int64, ctxLog map[string]interface{}) bool {
	abortar := func(status int, msg string, err error) bool {
		_ = tx.Rollback()
		if status == http.StatusInternalServerError {
			c.dbError("assign_pago.monto_error", msg, err, ctxLog)
		} else {
			c.fail(status, msg, nil)
		}
		return false
	}
	var pagoActual int64
	if err := tx.Raw("SELECT COALESCE(pk_id_pago, 0) FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", pedidoID).QueryRow(&pagoActual); err != nil {
		return abortar(http.StatusInternalServerError, "Error al bloquear el pedido", err)
	}
	if pagoActual > 0 {
		return abortar(http.StatusConflict, "El pedido ya tiene un pago asignado", nil)
	}
	m, err := montopedido.Calcular(tx, pedidoID)
	if err != nil {
		return abortar(http.StatusInternalServerError, "Error al calcular el monto del pedido", err)
	}
	if m.Lineas == 0 {
		return abortar(http.StatusConflict, "El pedido no tiene productos: no se puede asignar un pago", nil)
	}
	res, err := tx.Raw("UPDATE pago SET monto = ?, updated_at = ? WHERE pk_id_pago = ? AND estado_pago = ?", m.Total, time.Now().UTC(), pagoID, models.EstadoPagoPendiente).Exec()
	if err != nil {
		return abortar(http.StatusInternalServerError, "Error al fijar el monto del pago", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return abortar(http.StatusConflict, "Un cliente solo puede asignar un pago en estado PENDIENTE", nil)
	}
	return true
}

// @Title AssignPago
// @Summary Asignar un pago a un pedido
// @Description Asigna un pago existente a un pedido. Por defecto (`cambiar_estado=true`) marca además el pedido como TERMINADO y el pago como PAGADO (ambos cambios en una sola transacción); con `cambiar_estado=false` solo vincula el pago. Responde con el pedido completo actualizado. Si el pedido pasa a TERMINADO, avisa por push al cliente (best-effort, en segundo plano). Un Cliente solo puede vincular un pago a su propio pedido (uno ajeno responde 404) y siempre con `cambiar_estado=false` (403 si es true o se omite, porque terminar el pedido y marcar el pago PAGADO es cosa del personal); además el pago no puede pertenecer a otro pedido (404) y el pedido no puede tener ya un pago (409, para no cambiar el monto tras aplicar descuentos). EL SERVIDOR MANDA EL MONTO: al asignar el pago de un Cliente, en la misma transacción y con el pedido bloqueado, el servidor recalcula `pago.monto = MAX(0, SUM(detalle_pedido.precio x cantidad) - descuentos aplicados)` y lo guarda, cualquiera que fuera el monto con el que se creó el pago; responde 409 si el pedido no tiene productos o si el pago no está PENDIENTE. Desde ese momento el pedido queda congelado para el Cliente (`/producto_pedido` responde 409). El personal conserva el monto del pago tal cual (ajustes de mostrador).
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param pago_id query int true "ID del pago (entero positivo)"
// @Param cambiar_estado query bool false "true (defecto): pedido TERMINADO y pago PAGADO; false: solo vincula el pago" default(true)
// @Success 200 {object} models.ApiResponse{data=models.PedidoDoc} "Pago asignado al pedido"
// @Failure 400 {object} models.ApiResponse "pedido_id, pago_id o cambiar_estado inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente pidió cambiar_estado=true (explícito o por defecto)"
// @Failure 404 {object} models.ApiResponse "Pedido o pago no encontrado (para un Cliente, también si son de otro cliente)"
// @Failure 409 {object} models.ApiResponse "Conflicto con datos existentes, o un Cliente intenta reemplazar el pago ya asignado, asignar un pago que no está PENDIENTE o pagar un pedido sin productos"
// @Failure 500 {object} models.ApiResponse "Error al asignar pago"
// @Security BearerAuth
// @Router /pedidos/asignar-pago [post]
func (c *PedidoController) AssignPago() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	pagoID, err := httpx.PositiveInt64Param(&c.Controller, "pago_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.assign_pago.bad_request", err, map[string]interface{}{"pago_id": c.GetString("pago_id")})
		c.fail(http.StatusBadRequest, "El parámetro 'pago_id' es obligatorio y debe ser un entero positivo", err)
		return
	}
	cambiarEstado := true
	if raw := strings.TrimSpace(c.GetString("cambiar_estado")); raw != "" {
		if cambiarEstado, err = strconv.ParseBool(raw); err != nil {
			c.fail(http.StatusBadRequest, "El parámetro 'cambiar_estado' debe ser true o false", err)
			return
		}
	}
	if !claims.IsStaff() && cambiarEstado {
		c.fail(http.StatusForbidden, "Un cliente no puede terminar el pedido ni marcar el pago como PAGADO: use cambiar_estado=false", nil)
		return
	}
	o := orm.NewOrm()
	pedido := c.readPedido("assign_pago", o, claims)
	if pedido == nil {
		return
	}
	if !c.existsOrFail(o, "assign_pago", new(models.Pago), "PK_ID_PAGO", pagoID, msgPagoNoEncontrado) {
		return
	}
	if !claims.IsStaff() {
		if c.enOtroPedido(o, "assign_pago", "PK_ID_PAGO", pagoID, pedido.PK_ID_PEDIDO, msgPagoNoEncontrado) {
			return
		}
	}

	ctxLog := map[string]interface{}{"pedido_id": pedido.PK_ID_PEDIDO, "pago_id": pagoID, "cambiar_estado": cambiarEstado}
	tx, err := o.Begin()
	if err != nil {
		c.dbError("assign_pago.tx_begin_error", "Error al asignar pago", err, ctxLog)
		return
	}
	if !claims.IsStaff() && !c.fijarMontoCliente(tx, pedido.PK_ID_PEDIDO, pagoID, ctxLog) {
		return
	}
	cols := []string{"PK_ID_PAGO", "UPDATED_AT"}
	estadoAnterior := pedido.ESTADO_PEDIDO
	pedido.PK_ID_PAGO = &models.Pago{PK_ID_PAGO: pagoID}
	pedido.UPDATED_AT = time.Now().UTC()
	if cambiarEstado {
		pedido.ESTADO_PEDIDO = models.EstadoPedidoTerminado
		cols = append(cols, "ESTADO_PEDIDO")
	}
	if _, err := tx.Update(pedido, cols...); err != nil {
		_ = tx.Rollback()
		c.dbError("assign_pago.update_error", "Error al asignar pago", err, ctxLog)
		return
	}
	if cambiarEstado {
		if _, err := tx.QueryTable(new(models.Pago)).Filter("PK_ID_PAGO", pagoID).Update(orm.Params{"ESTADO_PAGO": models.EstadoPagoPagado, "UPDATED_AT": pedido.UPDATED_AT}); err != nil {
			_ = tx.Rollback()
			c.dbError("assign_pago.update_pago_error", "Error al marcar el pago como PAGADO", err, ctxLog)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.dbError("assign_pago.tx_commit_error", "Error al asignar pago", err, ctxLog)
		return
	}
	notificarEstado(pedido, estadoAnterior)
	httpx.Send(&c.Controller, http.StatusOK, "Pago asignado correctamente", *pedido)
}

// @Title UpdateEstadoPedido
// @Summary Actualizar el estado de un pedido
// @Description Solo personal (trabajador o administrador): un Cliente recibe 403. Actualiza el estado de un pedido existente (sin cuerpo: `pedido_id` y `estado` van como query params). Estados válidos: INICIADO, EN_PREPARACION, LISTO, TERMINADO, CANCELADO (no distingue mayúsculas). Responde con el pedido completo actualizado. Si el estado cambia (salvo a INICIADO), el servidor avisa por push al cliente del pedido (best-effort, en segundo plano).
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param estado query string true "Nuevo estado del pedido" Enums(INICIADO,EN_PREPARACION,LISTO,TERMINADO,CANCELADO)
// @Success 200 {object} models.ApiResponse{data=models.PedidoDoc} "Estado actualizado"
// @Failure 400 {object} models.ApiResponse "pedido_id inválido o estado inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere un usuario trabajador (un Cliente no puede cambiar estados)"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al actualizar estado del pedido"
// @Security BearerAuth
// @Router /pedidos/actualizar-estado [put]
func (c *PedidoController) UpdateEstadoPedido() {
	claims, ok := authz.RequireStaff(&c.Controller)
	if !ok {
		return
	}
	estado := strings.ToUpper(strings.TrimSpace(c.GetString("estado")))
	switch estado {
	case models.EstadoPedidoIniciado, models.EstadoPedidoEnPreparacion, models.EstadoPedidoListo,
		models.EstadoPedidoTerminado, models.EstadoPedidoCancelado:
	default:
		logging.LogControllerError(c.Ctx, "pedidos.update_estado.bad_request", nil, map[string]interface{}{"estado": c.GetString("estado")})
		c.fail(http.StatusBadRequest, "Estado inválido (INICIADO, EN_PREPARACION, LISTO, TERMINADO o CANCELADO)", nil)
		return
	}
	o := orm.NewOrm()
	pedido := c.readPedido("update_estado", o, claims)
	if pedido == nil {
		return
	}
	estadoAnterior := pedido.ESTADO_PEDIDO
	pedido.ESTADO_PEDIDO = estado
	pedido.UPDATED_AT = time.Now().UTC()
	if _, err := o.Update(pedido, "ESTADO_PEDIDO", "UPDATED_AT"); err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.update_estado.update_error", err, map[string]interface{}{"pedido_id": pedido.PK_ID_PEDIDO, "estado": estado})
		c.fail(http.StatusInternalServerError, "Error al actualizar estado del pedido", err)
		return
	}
	notificarEstado(pedido, estadoAnterior)
	httpx.Send(&c.Controller, http.StatusOK, "Estado del pedido actualizado correctamente", *pedido)
}

// @Title GetPedidoDetails
// @Summary Obtener detalles completos de un pedido
// @Description Devuelve el pedido con su método de pago y sus productos. Un Cliente solo ve sus propios pedidos (uno ajeno responde 404, igual que si no existiera); el personal ve cualquiera. A diferencia de `GET /pedidos`, las relaciones van como número (0 cuando no existen), `fechaPedido` como DD-MM-YYYY, `horaPedido` como HH:MM:SS y `productos` es un string con un JSON (`[]` si no hay productos) cuyos elementos son {pk_id_producto, nombre, cantidad, precio, subtotal}.
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.PedidoDetails} "Detalles del pedido obtenidos exitosamente"
// @Failure 400 {object} models.ApiResponse "pedido_id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al obtener los detalles del pedido"
// @Security BearerAuth
// @Router /pedidos/detalles [get]
func (c *PedidoController) GetPedidoDetails() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "pedidos.details.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.fail(http.StatusBadRequest, msgPedidoIDInvalido, err)
		return
	}

	query := `
SELECT
    p.pk_id_pedido                                   AS pk_id_pedido,
    COALESCE(TO_CHAR(p.fecha, 'DD-MM-YYYY'), '')     AS fecha,
    COALESCE(TO_CHAR(p.hora, 'HH24:MI:SS'), '')     AS hora,
    COALESCE(p.delivery, false)                      AS delivery,
    COALESCE(p.estado_pedido::text, '')              AS estado_pedido,
    COALESCE(mp.tipo, '')                            AS metodo_pago,
    COALESCE((
        SELECT jsonb_agg(json_build_object(
            'pk_id_producto', d.pk_id_producto,
            'nombre', pr.nombre,
            'cantidad', d.cantidad,
            'precio', d.precio,
            'subtotal', d.cantidad * d.precio
        ))::text
        FROM detalle_pedido d
        JOIN producto pr ON pr.pk_id_producto = d.pk_id_producto
        WHERE d.pk_id_pedido = p.pk_id_pedido
    ), '[]')                                           AS productos,
    COALESCE(p.pk_id_pago, 0)                        AS pago_id,
    COALESCE(pa.pk_id_metodo_pago, 0)                AS metodo_pago_id,
    COALESCE(p.pk_id_domicilio, 0)                   AS domicilio_id,
    COALESCE(p.pk_documento_cliente, 0)             AS pk_documento_cliente
FROM pedido p
LEFT JOIN pago pa        ON p.pk_id_pago = pa.pk_id_pago
LEFT JOIN metodo_pago mp ON pa.pk_id_metodo_pago = mp.pk_id_metodo_pago
WHERE p.pk_id_pedido = ?;
    `

	var details models.PedidoDetails
	if err := orm.NewOrm().Raw(query, pedidoID).QueryRow(&details); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return
		}
		logging.LogControllerError(c.Ctx, "pedidos.details.db_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.fail(http.StatusInternalServerError, "Error al obtener los detalles del pedido", err)
		return
	}
	if !authz.EsDuenio(claims, details.DocumentoCliente) {
		c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Detalles del pedido obtenidos exitosamente", details)
}
