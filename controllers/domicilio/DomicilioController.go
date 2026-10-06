package domicilio

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"restaurante/internal/httpx"
	"restaurante/internal/notify"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DomicilioController struct {
	web.Controller
}

const (
	msgIDInvalido           = "El parámetro 'id' es obligatorio y debe ser un entero positivo"
	msgNoEncontrado         = "Domicilio no encontrado"
	msgTrabajadorNoExiste   = "Trabajador no encontrado"
	msgEstadoInvalido       = "Campo 'estado' inválido (PENDIENTE, EN_CAMINO o ENTREGADO)"
	msgFechaInvalida        = "Formato de fecha inválido (use YYYY-MM-DD)"
	msgDireccionTelefonoReq = "Los campos 'direccion' y 'telefono' son obligatorios y no pueden estar vacíos"
)

// normalizeEstado devuelve el estado en mayúsculas y si pertenece al enum.
func normalizeEstado(e string) (string, bool) {
	e = strings.ToUpper(strings.TrimSpace(e))
	switch e {
	case models.EstadoDomicilioPendiente, models.EstadoDomicilioEnCamino, models.EstadoDomicilioEntregado:
		return e, true
	}
	return e, false
}

func (c *DomicilioController) fail(status int, msg string, err error) {
	httpx.Fail(&c.Controller, status, msg, err)
}

// dbError registra el error y responde 409 si es un conflicto de PostgreSQL
// (unicidad o llave foránea) y 500 en cualquier otro caso.
func (c *DomicilioController) dbError(op, msg string, err error, ctx map[string]interface{}) {
	logging.LogControllerError(c.Ctx, "domicilios."+op, err, ctx)
	if httpx.IsPGConflict(err) {
		c.fail(http.StatusConflict, msg+": conflicto con datos existentes", err)
		return
	}
	c.fail(http.StatusInternalServerError, msg, err)
}

// readDomicilio lee el domicilio id; responde 404/500 y devuelve false si no se puede.
func (c *DomicilioController) readDomicilio(op string, o orm.Ormer, d *models.Domicilio) bool {
	if err := o.Read(d); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			c.fail(http.StatusNotFound, msgNoEncontrado, nil)
			return false
		}
		logging.LogControllerError(c.Ctx, "domicilios."+op+".read_error", err, map[string]interface{}{"id": d.ID})
		c.fail(http.StatusInternalServerError, "Error al consultar el domicilio", err)
		return false
	}
	return true
}

// trabajadorExists comprueba que el trabajador exista; responde 404/500 y
// devuelve false si no se puede continuar.
func (c *DomicilioController) trabajadorExists(op string, o orm.Ormer, documento int64) bool {
	n, err := o.QueryTable(new(models.Trabajador)).Filter("PK_DOCUMENTO_TRABAJADOR", documento).Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios."+op+".trabajador_error", err, map[string]interface{}{"trabajador": documento})
		c.fail(http.StatusInternalServerError, "Error al validar el trabajador", err)
		return false
	}
	if n == 0 {
		c.fail(http.StatusNotFound, msgTrabajadorNoExiste, nil)
		return false
	}
	return true
}

// @Title GetAll
// @Summary Obtener todos los domicilios con posibilidad de filtrar
// @Description Devuelve los domicilios, con filtros opcionales combinables. Un filtro con formato inválido responde 400. Sin resultados responde 200 con `data` igual a `[]`. `fechaDomicilio` va como DD-MM-YYYY; `trabajadorAsignado` es el trabajador como objeto (solo `documentoTrabajador` es fiable) y se omite si no hay domiciliario. Con `trabajador` se devuelven solo los NO entregados que no tienen domiciliario o que tiene ese trabajador.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   direccion    query   string   false   "Filtrar por dirección (contiene, sin distinguir mayúsculas)"
// @Param   telefono     query   string   false   "Filtrar por teléfono (exacto)"
// @Param   fecha        query   string   false   "Filtrar por fecha (YYYY-MM-DD)"
// @Param   estado       query   string   false   "Filtrar por estado del domicilio" Enums(PENDIENTE,EN_CAMINO,ENTREGADO)
// @Param   updated_by   query   string   false   "Filtrar por usuario de la última actualización (contiene)"
// @Param   trabajador   query   int      false   "Documento del domiciliario solicitante (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.DomicilioDoc} "Lista de domicilios (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "Algún filtro tiene formato inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios [get]
func (c *DomicilioController) GetAll() {
	cond := orm.NewCondition()
	if v := c.GetString("direccion"); v != "" {
		cond = cond.And("Direccion__icontains", v)
	}
	if v := c.GetString("telefono"); v != "" {
		cond = cond.And("Telefono", v)
	}
	if v := c.GetString("updated_by"); v != "" {
		cond = cond.And("UpdatedBy__icontains", v)
	}
	if v := strings.TrimSpace(c.GetString("fecha")); v != "" {
		fecha, err := models.ParseDateToNoonUTC(v)
		if err != nil {
			c.fail(http.StatusBadRequest, "Parámetro 'fecha' inválido (use YYYY-MM-DD)", err)
			return
		}
		cond = cond.And("Fecha", fecha)
	}
	if v := c.GetString("estado"); v != "" {
		estado, ok := normalizeEstado(v)
		if !ok {
			c.fail(http.StatusBadRequest, "Parámetro 'estado' inválido (PENDIENTE, EN_CAMINO o ENTREGADO)", nil)
			return
		}
		cond = cond.And("Estado", estado)
	}
	if c.GetString("trabajador") != "" {
		trabajador, err := httpx.PositiveInt64Param(&c.Controller, "trabajador")
		if err != nil {
			logging.LogControllerError(c.Ctx, "domicilios.getall.bad_request", err, map[string]interface{}{"trabajador": c.GetString("trabajador")})
			c.fail(http.StatusBadRequest, "Parámetro 'trabajador' inválido", err)
			return
		}
		cond = cond.And("Entregado", false).AndCond(orm.NewCondition().
			Or("Trabajador__isnull", true).
			Or("Trabajador", trabajador))
	}

	var domicilios []models.Domicilio
	if _, err := orm.NewOrm().QueryTable(new(models.Domicilio)).SetCond(cond).OrderBy("ID").All(&domicilios); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getall.db_error", err, nil)
		c.fail(http.StatusInternalServerError, "Error al obtener domicilios de la base de datos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Domicilios obtenidos exitosamente", httpx.List(domicilios))
}

// domicilioDetalle es el `data` de GET /domicilios/search (documentado como
// models.DomicilioDetalleDoc).
type domicilioDetalle struct {
	Domicilio models.Domicilio            `json:"domicilio"`
	Cliente   *models.DomicilioClienteDoc `json:"cliente,omitempty"`
	Pedido    *models.DomicilioPedidoDoc  `json:"pedido,omitempty"`
}

type clienteRow struct {
	Documento int64  `orm:"column(documento)"`
	Nombre    string `orm:"column(nombre)"`
	Apellido  string `orm:"column(apellido)"`
}

type pedidoRow struct {
	PedidoID          int64           `orm:"column(pedido_id)"`
	PagoID            sql.NullInt64   `orm:"column(pago_id)"`
	PagoMonto         sql.NullFloat64 `orm:"column(pago_monto)"`
	SubtotalProductos sql.NullFloat64 `orm:"column(subtotal_productos)"`
	Productos         string          `orm:"column(productos)"`
}

// @Title GetById
// @Summary Obtener domicilio por ID (incluye cliente y pedido asociado si existen)
// @Description Devuelve un domicilio por ID y, si está asociado a un pedido, el cliente (`cliente`) y el resumen del último pedido (`pedido`: pago, subtotal, total y productos). `cliente` y `pedido` se omiten si no hay pedido asociado. `fechaDomicilio` va como DD-MM-YYYY.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del domicilio (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.DomicilioDetalleDoc} "Domicilio encontrado (con cliente/pedido si aplica)"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	resp := domicilioDetalle{Domicilio: models.Domicilio{ID: id}}
	if !c.readDomicilio("getbyid", o, &resp.Domicilio) {
		return
	}

	var cli clienteRow
	qCliente := `
SELECT p.pk_documento_cliente AS documento,
       c.nombre               AS nombre,
       c.apellido             AS apellido
FROM pedido p
JOIN cliente c ON c.pk_documento_cliente = p.pk_documento_cliente
WHERE p.pk_id_domicilio = ?
ORDER BY p.pk_id_pedido DESC LIMIT 1;`
	if err := o.Raw(qCliente, id).QueryRow(&cli); err == nil {
		resp.Cliente = &models.DomicilioClienteDoc{Documento: cli.Documento, Nombre: cli.Nombre, Apellido: cli.Apellido}
	} else if !errors.Is(err, orm.ErrNoRows) {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.cliente_query_error", err, map[string]interface{}{"id": id})
		c.fail(http.StatusInternalServerError, "Error al consultar el cliente del domicilio", err)
		return
	}

	var ped pedidoRow
	qPedido := `
SELECT p.pk_id_pedido AS pedido_id,
       pa.pk_id_pago  AS pago_id,
       pa.monto::numeric AS pago_monto,
       (SELECT COALESCE(SUM(d.cantidad * d.precio),0)
          FROM detalle_pedido d
         WHERE d.pk_id_pedido = p.pk_id_pedido) AS subtotal_productos,
       (SELECT COALESCE(jsonb_agg(json_build_object(
           'pk_id_producto', d.pk_id_producto,
           'nombre',        pr.nombre,
           'cantidad',      d.cantidad,
           'precio',        d.precio,
           'subtotal',      d.cantidad * d.precio
       )), '[]'::jsonb)::text
          FROM detalle_pedido d
          JOIN producto pr ON pr.pk_id_producto = d.pk_id_producto
         WHERE d.pk_id_pedido = p.pk_id_pedido) AS productos
FROM pedido p
LEFT JOIN pago pa ON pa.pk_id_pago = p.pk_id_pago
WHERE p.pk_id_domicilio = ?
ORDER BY p.pk_id_pedido DESC LIMIT 1;`
	if err := o.Raw(qPedido, id).QueryRow(&ped); err == nil {
		resumen, err := resumenPedido(ped)
		if err != nil {
			logging.LogControllerError(c.Ctx, "domicilios.getbyid.productos_error", err, map[string]interface{}{"id": id})
			c.fail(http.StatusInternalServerError, "Error al leer los productos del pedido", err)
			return
		}
		resp.Pedido = resumen
	} else if !errors.Is(err, orm.ErrNoRows) {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.pedido_query_error", err, map[string]interface{}{"id": id})
		c.fail(http.StatusInternalServerError, "Error al consultar el pedido del domicilio", err)
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Domicilio encontrado", resp)
}

// resumenPedido arma el resumen de pedido de GET /domicilios/search. `total` es
// el monto del pago si existe y, si no, el subtotal de los productos.
func resumenPedido(ped pedidoRow) (*models.DomicilioPedidoDoc, error) {
	productos := []models.DetalleProductoDoc{}
	if err := json.Unmarshal([]byte(ped.Productos), &productos); err != nil {
		return nil, err
	}
	total := ped.SubtotalProductos.Float64
	if ped.PagoMonto.Valid {
		total = ped.PagoMonto.Float64
	}
	var pagoID *int64
	if ped.PagoID.Valid {
		pagoID = &ped.PagoID.Int64
	}
	return &models.DomicilioPedidoDoc{
		PedidoId:          ped.PedidoID,
		PagoId:            pagoID,
		MontoPago:         ped.PagoMonto.Float64,
		SubtotalProductos: ped.SubtotalProductos.Float64,
		Total:             total,
		Productos:         productos,
	}, nil
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un domicilio. `direccion`, `telefono` (no vacíos) y `fechaDomicilio` (YYYY-MM-DD) son obligatorios; `estadoDomicilio` (alias `estado`) es opcional y, si se omite, aplica el valor por defecto de la base de datos. `trabajadorAsignado` es el documento del trabajador (null o 0 = sin asignar; si se envía debe existir, 404 si no). `entregado` lo calcula la base de datos y no debe enviarse. Responde 201 con el domicilio creado (`fechaDomicilio` como DD-MM-YYYY).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.DomicilioCreate true  "Datos del domicilio a crear (sólo campos permitidos)"
// @Success 201 {object} models.ApiResponse{data=models.DomicilioDoc} "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "JSON inválido, campos obligatorios vacíos, fecha o estado inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "El trabajador indicado no existe"
// @Failure 409 {object} models.ApiResponse "Conflicto con datos existentes"
// @Failure 500 {object} models.ApiResponse "Error al crear el domicilio"
// @Security BearerAuth
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	var in struct {
		models.DomicilioCreate
		EstadoAlias string `json:"estado"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Error al procesar la solicitud", err)
		return
	}
	direccion, telefono := strings.TrimSpace(in.Direccion), strings.TrimSpace(in.Telefono)
	if direccion == "" || telefono == "" {
		c.fail(http.StatusBadRequest, msgDireccionTelefonoReq, nil)
		return
	}
	fecha, err := models.ParseDateToNoonUTC(in.FechaDomicilio)
	if err != nil {
		c.fail(http.StatusBadRequest, msgFechaInvalida, err)
		return
	}
	cols := []string{"direccion", "fecha", "telefono"}
	vals := []interface{}{direccion, fecha, telefono}
	if in.Observaciones != nil {
		cols, vals = append(cols, "observaciones"), append(vals, *in.Observaciones)
	}
	if in.CreatedBy != nil {
		cols, vals = append(cols, "created_by"), append(vals, *in.CreatedBy)
	}
	if est := firstNonEmpty(in.Estado, in.EstadoAlias); est != "" {
		estado, ok := normalizeEstado(est)
		if !ok {
			c.fail(http.StatusBadRequest, msgEstadoInvalido, nil)
			return
		}
		cols, vals = append(cols, "estado_domicilio"), append(vals, estado)
	}
	o := orm.NewOrm()
	if in.TrabajadorID != nil && *in.TrabajadorID != 0 {
		if *in.TrabajadorID < 0 {
			c.fail(http.StatusBadRequest, "El campo 'trabajadorAsignado' debe ser un documento positivo", nil)
			return
		}
		if !c.trabajadorExists("post", o, *in.TrabajadorID) {
			return
		}
		cols, vals = append(cols, "pk_documento_trabajador"), append(vals, *in.TrabajadorID)
	}

	ph := make([]string, len(vals))
	for i := range ph {
		ph[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO domicilio (%s) VALUES (%s) RETURNING pk_id_domicilio",
		strings.Join(cols, ","), strings.Join(ph, ","))
	var domicilio models.Domicilio
	if err := o.Raw(query, vals...).QueryRow(&domicilio.ID); err != nil {
		c.dbError("post.insert_error", "Error al crear el domicilio", err, map[string]interface{}{"direccion": direccion, "telefono": telefono})
		return
	}
	// Se relee la fila para devolver los valores que fija la base de datos
	// (entregado, estado por defecto, created_at, updated_at).
	if !c.readDomicilio("post", o, &domicilio) {
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Domicilio creado correctamente", domicilio)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualización parcial (merge): los campos ausentes del cuerpo se conservan; el cuerpo puede ser parcial (incluso `{}`, que solo refresca `updatedAt`). Campos: `direccion` y `telefono` (no vacíos), `estado` (alias `estadoDomicilio`: PENDIENTE, EN_CAMINO o ENTREGADO; permite marcar un domicilio como entregado), `observaciones`, `fechaDomicilio` (YYYY-MM-DD) y `updatedBy`. Anulables (null los limpia): `observaciones` y `updatedBy`; null en cualquier otro campo responde 400. `entregado` lo calcula la base de datos; la respuesta lo trae actualizado. Cuando el domicilio pasa a ENTREGADO, avisa por push al cliente del pedido (best-effort, en segundo plano).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del domicilio (entero positivo)"
// @Param   body  body   models.DomicilioUpdateRequest true  "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.DomicilioDoc} "Domicilio actualizado"
// @Failure 400 {object} models.ApiResponse "id inválido, JSON inválido, null en campo no anulable o valores inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Failure 409 {object} models.ApiResponse "Conflicto con datos existentes"
// @Failure 500 {object} models.ApiResponse "Error al actualizar el domicilio"
// @Security BearerAuth
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	domicilio := models.Domicilio{ID: id}
	if !c.readDomicilio("put", o, &domicilio) {
		return
	}
	estadoAnterior := domicilio.Estado
	var in models.DomicilioUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in, "observaciones", "updatedBy"); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.put.bad_json", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Error al procesar la solicitud", err)
		return
	}

	cols := []string{"UpdatedAt"}
	if in.Direccion != nil || in.Telefono != nil {
		if (in.Direccion != nil && strings.TrimSpace(*in.Direccion) == "") || (in.Telefono != nil && strings.TrimSpace(*in.Telefono) == "") {
			c.fail(http.StatusBadRequest, msgDireccionTelefonoReq, nil)
			return
		}
	}
	if in.Direccion != nil {
		domicilio.Direccion = strings.TrimSpace(*in.Direccion)
		cols = append(cols, "Direccion")
	}
	if in.Telefono != nil {
		domicilio.Telefono = strings.TrimSpace(*in.Telefono)
		cols = append(cols, "Telefono")
	}
	if est := firstNonNil(in.Estado, in.EstadoDomicilio); est != nil {
		estado, ok := normalizeEstado(*est)
		if !ok {
			c.fail(http.StatusBadRequest, msgEstadoInvalido, nil)
			return
		}
		domicilio.Estado = estado
		cols = append(cols, "Estado")
	}
	if in.FechaDomicilio != nil {
		fecha, err := models.ParseDateToNoonUTC(*in.FechaDomicilio)
		if err != nil {
			c.fail(http.StatusBadRequest, msgFechaInvalida, err)
			return
		}
		domicilio.Fecha = fecha
		cols = append(cols, "Fecha")
	}
	body := c.Ctx.Input.RequestBody
	if in.Observaciones != nil || httpx.IsNull(body, "observaciones") {
		domicilio.Observ = in.Observaciones
		cols = append(cols, "Observ")
	}
	if in.UpdatedBy != nil || httpx.IsNull(body, "updatedBy") {
		domicilio.UpdatedBy = in.UpdatedBy
		cols = append(cols, "UpdatedBy")
	}
	domicilio.UpdatedAt = time.Now().UTC()

	if _, err := o.Update(&domicilio, cols...); err != nil {
		c.dbError("put.update_error", "Error al actualizar el domicilio", err, map[string]interface{}{"id": id, "body": string(body)})
		return
	}
	// Se relee para devolver `entregado` y las marcas de tiempo calculados por la base de datos.
	if !c.readDomicilio("put", o, &domicilio) {
		return
	}
	if estadoAnterior != models.EstadoDomicilioEntregado && domicilio.Estado == models.EstadoDomicilioEntregado {
		notify.Enviar(notify.Evento{Tipo: notify.DomicilioEntregado, DomicilioID: id})
	}
	httpx.Send(&c.Controller, http.StatusOK, "Domicilio actualizado correctamente", domicilio)
}

// @Title Delete
// @Summary Eliminar un domicilio
// @Description Elimina un domicilio. Si está asociado a un pedido responde 409. Responde 200 con el mensaje de confirmación (sin `data`).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del domicilio (entero positivo)"
// @Success 200 {object} models.ApiResponse "Domicilio eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Failure 409 {object} models.ApiResponse "El domicilio está asociado a un pedido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios [delete]
func (c *DomicilioController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.fail(http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	n, err := orm.NewOrm().Delete(&models.Domicilio{ID: id})
	if err != nil {
		c.dbError("delete.delete_error", "Error al eliminar el domicilio", err, map[string]interface{}{"id": id})
		return
	}
	if n == 0 {
		c.fail(http.StatusNotFound, msgNoEncontrado, nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Domicilio eliminado", nil)
}

// @Title AsignarDomiciliario
// @Summary Asignar un domiciliario a un domicilio
// @Description Un domiciliario toma un domicilio que aún no tiene asignado: queda EN_CAMINO y con ese trabajador. Responde 404 si el domicilio o el trabajador no existen y 409 si el domicilio ya estaba asignado. `data` es el domicilio completo actualizado. Avisa por push al cliente del pedido (best-effort, en segundo plano).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param domicilio_id query int true "ID del domicilio (entero positivo)"
// @Param trabajador_id query int true "Documento del domiciliario que lo tomará (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.DomicilioDoc} "Domicilio asignado"
// @Failure 400 {object} models.ApiResponse "domicilio_id o trabajador_id inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio o trabajador no encontrado"
// @Failure 409 {object} models.ApiResponse "El domicilio ya ha sido asignado"
// @Failure 500 {object} models.ApiResponse "Error al asignar domicilio"
// @Security BearerAuth
// @Router /domicilios/asignar [post]
func (c *DomicilioController) AsignarDomiciliario() {
	domicilioID, err := httpx.PositiveInt64Param(&c.Controller, "domicilio_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.asignar.bad_request", err, map[string]interface{}{"domicilio_id": c.GetString("domicilio_id")})
		c.fail(http.StatusBadRequest, "El parámetro 'domicilio_id' es obligatorio y debe ser un entero positivo", err)
		return
	}
	trabajadorID, err := httpx.PositiveInt64Param(&c.Controller, "trabajador_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.asignar.bad_request", err, map[string]interface{}{"trabajador_id": c.GetString("trabajador_id")})
		c.fail(http.StatusBadRequest, "El parámetro 'trabajador_id' es obligatorio y debe ser un entero positivo", err)
		return
	}
	ctx := map[string]interface{}{"domicilio_id": domicilioID, "trabajador_id": trabajadorID}
	o := orm.NewOrm()
	if !c.trabajadorExists("asignar", o, trabajadorID) {
		return
	}
	res, err := o.Raw(
		"UPDATE domicilio SET estado_domicilio='EN_CAMINO', pk_documento_trabajador=? WHERE pk_id_domicilio=? AND pk_documento_trabajador IS NULL",
		trabajadorID, domicilioID,
	).Exec()
	if err != nil {
		c.dbError("asignar.update_error", "Error al asignar domicilio", err, ctx)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.dbError("asignar.rows_affected_error", "Error al asignar domicilio", err, ctx)
		return
	}
	domicilio := models.Domicilio{ID: domicilioID}
	if !c.readDomicilio("asignar", o, &domicilio) {
		return
	}
	if affected != 1 {
		c.fail(http.StatusConflict, "Este domicilio ya ha sido asignado", nil)
		return
	}
	notify.Enviar(notify.Evento{Tipo: notify.DomicilioAsignado, DomicilioID: domicilioID})
	httpx.Send(&c.Controller, http.StatusOK, "Domicilio asignado correctamente", domicilio)
}
