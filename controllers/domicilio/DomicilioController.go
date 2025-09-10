package domicilio

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// jsonMarshal is a variable to allow mocking json.Marshal in tests.
var jsonMarshal = json.Marshal

type DomicilioController struct {
	web.Controller
}

func isValidEstadoDomicilio(e string) bool {
	switch models.EstadoDomicilio(strings.ToUpper(e)) {
	case models.EstadoDomicilioPendiente, models.EstadoDomicilioEnCamino, models.EstadoDomicilioEntregado:
		return true
	}
	return false
}

// @Title GetAll
// @Summary Obtener todos los domicilios con posibilidad de filtrar
// @Description Devuelve todos los domicilios registrados en la base de datos, filtrando según criterios específicos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   direccion    query   string   false   "Filtrar por dirección"
// @Param   telefono     query   string   false   "Filtrar por teléfono"
// @Param   fecha        query   string   false   "Filtrar por fecha"
// @Param   estado       query   string   false   "Filtrar por estado del domicilio"
// @Param   updated_by   query   string   false   "Filtrar por usuario que realizó la última actualización"
// @Param   trabajador   query   int      false   "ID del domiciliario solicitante"
// @Success 200 {object} models.ApiResponse{data=[]models.Domicilio} "Lista de domicilios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios [get]
func (c *DomicilioController) GetAll() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.Domicilio))

	// Leer parámetros de la URL
	direccion := c.GetString("direccion")
	telefono := c.GetString("telefono")
	updatedBy := c.GetString("updated_by")
	fecha := c.GetString("fecha")
	estado := strings.ToUpper(c.GetString("estado"))
	trabajadorID, errTrab := c.GetInt("trabajador") // ID del domiciliario solicitante
	if c.GetString("trabajador") != "" && errTrab != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getall.bad_request", errTrab, map[string]interface{}{"trabajador": c.GetString("trabajador")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'trabajador' inválido", Cause: errTrab.Error()}
		_ = c.ServeJSON()
		return
	}

	// Aplicar filtros opcionales SOLO si se proporcionan (usar nombres de campos del struct)
	if direccion != "" {
		qs = qs.Filter("Direccion__icontains", direccion)
	}
	if telefono != "" {
		qs = qs.Filter("Telefono", telefono)
	}
	if updatedBy != "" {
		qs = qs.Filter("UpdatedBy__icontains", updatedBy)
	}
	if fecha != "" {
		if parsed, err := time.Parse("2006-01-02", fecha); err == nil {
			qs = qs.Filter("Fecha", parsed)
		}
	}
	if estado != "" {
		qs = qs.Filter("Estado", models.EstadoDomicilio(estado))
	}

	// Aplicar condición para que los domiciliarios solo vean pedidos que pueden tomar
	if trabajadorID != 0 {
		cond := orm.NewCondition().
			Or("Trabajador__isnull", true).
			Or("Trabajador__PK_DOCUMENTO_TRABAJADOR", trabajadorID)

		qs = qs.Filter("Entregado", false).SetCond(cond)
	}

	// Ejecutar consulta
	var domicilios []models.Domicilio
	count, err := qs.All(&domicilios)
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getall.db_error", err, map[string]interface{}{
			"direccion": direccion, "telefono": telefono, "updated_by": updatedBy, "fecha": fecha, "estado": estado, "trabajador": trabajadorID,
		})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener domicilios de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Si no hay domicilios, retornar mensaje informativo
	if count == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron domicilios que coincidan con los filtros proporcionados",
		}
		_ = c.ServeJSON()
		return
	}

	// Responder con los domicilios filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilios obtenidos exitosamente",
		Data:    domicilios,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener domicilio por ID (incluye cliente y pedido asociado si existen)
// @Description Devuelve un domicilio por ID y, si está asociado a un pedido, incluye documento/nombre del cliente y resumen del pedido (monto/productos).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 200 {object} models.ApiResponse{data=models.Domicilio} "Domicilio encontrado (con cliente/pedido si aplica)"
// @Failure 400 {object} models.ApiResponse "Parámetro inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{ID: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// ---- Datos relacionados: cliente y pedido ----
	type clienteRow struct {
		Documento int64  `orm:"column(documento)"`
		Nombre    string `orm:"column(nombre)"`
		Apellido  string `orm:"column(apellido)"`
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
	if err := o.Raw(qCliente, id).QueryRow(&cli); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.cliente_query_error", err, map[string]interface{}{"id": id})
	}

	type pedidoRow struct {
		PedidoID          int64           `orm:"column(pedido_id)"`
		PagoID            sql.NullInt64   `orm:"column(pago_id)"`
		PagoMonto         sql.NullFloat64 `orm:"column(pago_monto)"`
		SubtotalProductos sql.NullFloat64 `orm:"column(subtotal_productos)"`
		Productos         string          `orm:"column(productos)"` // JSON
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
	if err := o.Raw(qPedido, id).QueryRow(&ped); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.getbyid.pedido_query_error", err, map[string]interface{}{"id": id})
	}

	resp := map[string]interface{}{"domicilio": domicilio}
	if cli.Documento != 0 {
		resp["cliente"] = map[string]interface{}{
			"documento": cli.Documento,
			"nombre":    cli.Nombre,
			"apellido":  cli.Apellido,
		}
	}
	if ped.PedidoID != 0 {
		var productos []map[string]interface{}
		if ped.Productos != "" {
			if err := json.Unmarshal([]byte(ped.Productos), &productos); err != nil {
				productos = nil
			}
		}
		total := 0.0
		if ped.PagoMonto.Valid {
			total = ped.PagoMonto.Float64
		} else if ped.SubtotalProductos.Valid {
			total = ped.SubtotalProductos.Float64
		}
		var pagoIDPtr *int64
		if ped.PagoID.Valid {
			v := ped.PagoID.Int64
			pagoIDPtr = &v
		}
		resp["pedido"] = map[string]interface{}{
			"pedidoId":          ped.PedidoID,
			"pagoId":            pagoIDPtr,
			"montoPago":         ped.PagoMonto.Float64,
			"subtotalProductos": ped.SubtotalProductos.Float64,
			"total":             total,
			"productos":         productos,
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio encontrado",
		Data:    resp,
	}
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un nuevo domicilio en la base de datos. El campo 'entregado' es generado automáticamente y no debe enviarse en la solicitud.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.DomicilioCreate true  "Datos del domicilio a crear (sólo campos permitidos)"
// @Success 201 {object} models.ApiResponse{data=models.Domicilio} "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	var input models.DomicilioCreate
	var raw map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &raw); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	// Normalizar trabajadorAsignado: 0, "", null => no asignado
	if v, ok := raw["trabajadorAsignado"]; ok {
		switch vv := v.(type) {
		case nil:
			delete(raw, "trabajadorAsignado")
		case string:
			if strings.TrimSpace(vv) == "" {
				delete(raw, "trabajadorAsignado")
			}
		case float64:
			if int64(vv) == 0 {
				delete(raw, "trabajadorAsignado")
			}
		}
	}
	// Decodificar al DTO ya sanitizado
	if bodySan, err := jsonMarshal(raw); err == nil {
		if err := json.Unmarshal(bodySan, &input); err != nil {
			logging.LogControllerError(c.Ctx, "domicilios.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar la solicitud", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	} else {
		logging.LogControllerError(c.Ctx, "domicilios.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Solicitud inválida", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if input.Direccion == "" || input.Telefono == "" {
		logging.LogControllerError(c.Ctx, "domicilios.post.validation_error", nil, map[string]interface{}{"missing": "direccion/telefono", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Los campos 'direccion' y 'telefono' son obligatorios"}
		_ = c.ServeJSON()
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.FechaDomicilio)
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.post.validation_error", err, map[string]interface{}{"fecha": input.FechaDomicilio, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{Direccion: input.Direccion, Telefono: input.Telefono, Fecha: parsedDate}
	if input.Observaciones != nil {
		domicilio.Observ = input.Observaciones
	}
	if input.CreatedBy != nil {
		domicilio.CreatedBy = input.CreatedBy
	}
	// compat: aceptar "estado" o "estadoDomicilio"
	est := string(input.Estado)
	if est == "" {
		if v, ok := raw["estado"].(string); ok {
			est = v
		}
	}
	if est != "" {
		if !isValidEstadoDomicilio(est) {
			logging.LogControllerError(c.Ctx, "domicilios.post.validation_error", nil, map[string]interface{}{"estado": est})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Campo 'estado' inválido"}
			_ = c.ServeJSON()
			return
		}
		domicilio.Estado = models.EstadoDomicilio(strings.ToUpper(est))
	}
	if input.TrabajadorID != nil && *input.TrabajadorID != 0 {
		domicilio.Trabajador = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: *input.TrabajadorID}
	}

	o := orm.NewOrm()
	cols := []string{"direccion", "fecha", "telefono"}
	vals := []interface{}{domicilio.Direccion, domicilio.Fecha, domicilio.Telefono}
	if domicilio.Observ != nil {
		cols = append(cols, "observaciones")
		vals = append(vals, *domicilio.Observ)
	}
	if domicilio.CreatedBy != nil {
		cols = append(cols, "created_by")
		vals = append(vals, *domicilio.CreatedBy)
	}
	if domicilio.Trabajador != nil {
		cols = append(cols, "pk_documento_trabajador")
		vals = append(vals, domicilio.Trabajador.PK_DOCUMENTO_TRABAJADOR)
	}
	if domicilio.Estado != "" {
		cols = append(cols, "estado_domicilio")
		vals = append(vals, domicilio.Estado)
	}

	ph := make([]string, len(vals))
	for i := range ph {
		ph[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO domicilio (%s) VALUES (%s) RETURNING pk_id_domicilio",
		strings.Join(cols, ","), strings.Join(ph, ","))

	if err := o.Raw(query, vals...).QueryRow(&domicilio.ID); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.post.insert_error", err, map[string]interface{}{"direccion": domicilio.Direccion, "telefono": domicilio.Telefono, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el domicilio",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	var ent bool
	var created, updated time.Time
	if err := o.Raw("SELECT entregado, created_at, updated_at FROM domicilio WHERE pk_id_domicilio = ?",
		domicilio.ID).QueryRow(&ent, &created, &updated); err == nil {
		domicilio.Entregado = ent
		domicilio.CreatedAt = created
		domicilio.UpdatedAt = updated
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Domicilio creado correctamente",
		Data:    domicilio,
	}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualiza los datos de un domicilio existente. El campo 'entregado' es calculado automáticamente.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Domicilio"
// @Param   body  body   models.DomicilioUpdateRequest true  "Datos del domicilio a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse{data=models.Domicilio} "Domicilio actualizado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	o := orm.NewOrm()
	id, err := strconv.ParseInt(c.GetString("id"), 10, 64)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "domicilios.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{ID: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.put.bad_json", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Construir lista de columnas a actualizar para evitar tocar columnas protegidas (p. ej., "Entregado")
	var colsToUpdate []string
	if direccion, ok := input["direccion"].(string); ok {
		domicilio.Direccion = direccion
		colsToUpdate = append(colsToUpdate, "Direccion")
	}
	if telefono, ok := input["telefono"].(string); ok {
		domicilio.Telefono = telefono
		colsToUpdate = append(colsToUpdate, "Telefono")
	}
	if updatedBy, ok := input["updatedBy"].(string); ok {
		domicilio.UpdatedBy = &updatedBy
		colsToUpdate = append(colsToUpdate, "UpdatedBy")
	}
	// Siempre actualizamos la marca de tiempo de modificación
	domicilio.UpdatedAt = time.Now().UTC()
	colsToUpdate = append(colsToUpdate, "UpdatedAt")

	if _, err := o.Update(&domicilio, colsToUpdate...); err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.put.update_error", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el domicilio", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	var ent bool
	var created, updated time.Time
	if err := o.Raw("SELECT entregado, created_at, updated_at FROM domicilio WHERE pk_id_domicilio = ?",
		domicilio.ID).QueryRow(&ent, &created, &updated); err == nil {
		domicilio.Entregado = ent
		domicilio.CreatedAt = created
		domicilio.UpdatedAt = updated
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio actualizado correctamente",
		Data:    domicilio,
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un domicilio
// @Description Elimina un domicilio de la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 204 {object} nil "Domicilio eliminado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [delete]
func (c *DomicilioController) Delete() {
	o := orm.NewOrm()
	id, err := strconv.ParseInt(c.GetString("id"), 10, 64)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "domicilios.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{ID: id}
	if _, err := o.Delete(&domicilio); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Domicilio eliminado",
		}
	} else {
		logging.LogControllerError(c.Ctx, "domicilios.delete.delete_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
	}
	_ = c.ServeJSON()
}

// @Title AsignarDomiciliario
// @Summary Asignar un domiciliario a un pedido de domicilio
// @Description Un domiciliario puede tomar un pedido si no ha sido asignado previamente
// @Tags domicilios
// @Accept json
// @Produce json
// @Param domicilio_id query int true "ID del domicilio"
// @Param trabajador_id query int true "ID del domiciliario que lo tomará"
// @Success 200 {object} models.ApiResponse "Domicilio asignado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado o ya asignado"
// @Failure 500 {object} models.ApiResponse "Error al asignar domicilio"
// @Security BearerAuth
// @Router /domicilios/asignar [post]
func (c *DomicilioController) AsignarDomiciliario() {
	domicilioID, _ := c.GetInt64("domicilio_id")
	trabajadorID, _ := c.GetInt64("trabajador_id")

	o := orm.NewOrm()
	res, err := o.Raw(
		"UPDATE domicilio SET estado_domicilio='EN_CAMINO', pk_documento_trabajador=? WHERE pk_id_domicilio=? AND pk_documento_trabajador IS NULL",
		trabajadorID, domicilioID,
	).Exec()
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.asignar.update_error", err, map[string]interface{}{"domicilio_id": domicilioID, "trabajador_id": trabajadorID})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al asignar domicilio",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logging.LogControllerError(c.Ctx, "domicilios.asignar.rows_affected_error", err, map[string]interface{}{"domicilio_id": domicilioID})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al asignar domicilio",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if affected != 1 {
		var exists int
		if err := o.Raw("SELECT COUNT(1) FROM domicilio WHERE pk_id_domicilio = ?", domicilioID).QueryRow(&exists); err != nil || exists == 0 {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Domicilio no encontrado",
			}
		} else {
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "Este domicilio ya ha sido asignado",
			}
		}
		_ = c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{
		ID:     domicilioID,
		Estado: models.EstadoDomicilioEnCamino,
		Trabajador: &models.Trabajador{
			PK_DOCUMENTO_TRABAJADOR: trabajadorID,
		},
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio asignado correctamente",
		Data:    domicilio,
	}
	_ = c.ServeJSON()
}
