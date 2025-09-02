package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/models"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

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
// @Success 200 {array} models.Domicilio "Lista de domicilios"
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
	trabajadorID, _ := c.GetInt("trabajador") // ID del domiciliario solicitante

	// Aplicar filtros opcionales SOLO si se proporcionan
	if direccion != "" {
		qs = qs.Filter("DIRECCION__icontains", direccion)
	}
	if telefono != "" {
		qs = qs.Filter("TELEFONO", telefono)
	}
	if updatedBy != "" {
		qs = qs.Filter("UPDATED_BY__icontains", updatedBy)
	}
	if fecha != "" {
		qs = qs.Filter("FECHA", fecha)
	}
	if estado != "" {
		qs = qs.Filter("ESTADO_DOMICILIO", estado)
	}

	// Aplicar condición para que los domiciliarios solo vean pedidos que pueden tomar
	if trabajadorID != 0 {
		cond := orm.NewCondition().
			Or("PK_DOCUMENTO_TRABAJADOR__isnull", true).
			Or("PK_DOCUMENTO_TRABAJADOR", trabajadorID)

		qs = qs.Filter("ENTREGADO", false).SetCond(cond)
	}

	// Ejecutar consulta
	var domicilios []models.Domicilio
	count, err := qs.All(&domicilios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener domicilios de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Si no hay domicilios, retornar mensaje informativo
	if count == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron domicilios que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Responder con los domicilios filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilios obtenidos exitosamente",
		Data:    domicilios,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener domicilio por ID (incluye cliente y pedido asociado si existen)
// @Description Devuelve un domicilio por ID y, si está asociado a un pedido, incluye documento/nombre del cliente y resumen del pedido (monto/productos).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 200 {object} models.ApiResponse "Domicilio encontrado (con cliente/pedido si aplica)"
// @Failure 400 {object} models.ApiResponse "Parámetro inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
		c.ServeJSON()
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
	cliErr := o.Raw(qCliente, id).QueryRow(&cli)

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
       )),'[]'::jsonb)::text
          FROM detalle_pedido d
          JOIN producto pr ON pr.pk_id_producto = d.pk_id_producto
         WHERE d.pk_id_pedido = p.pk_id_pedido) AS productos
FROM pedido p
LEFT JOIN pago pa ON pa.pk_id_pago = p.pk_id_pago
WHERE p.pk_id_domicilio = ?
ORDER BY p.pk_id_pedido DESC LIMIT 1;`
	pedErr := o.Raw(qPedido, id).QueryRow(&ped)

	resp := map[string]interface{}{"domicilio": domicilio}
	if cliErr == nil {
		resp["cliente"] = map[string]interface{}{
			"documento": cli.Documento,
			"nombre":    cli.Nombre,
			"apellido":  cli.Apellido,
		}
	}
	if pedErr == nil {
		var productos []map[string]interface{}
		if ped.Productos != "" {
			_ = json.Unmarshal([]byte(ped.Productos), &productos)
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
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un nuevo domicilio en la base de datos. El campo 'entregado' es generado automáticamente y no debe enviarse en la solicitud.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.DomicilioCreate true  "Datos del domicilio a crear (sólo campos permitidos)"
// @Success 201 {object} models.Domicilio "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	var input models.DomicilioCreate
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if input.Direccion == "" || input.Telefono == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Los campos 'direccion' y 'telefono' son obligatorios",
		}
		c.ServeJSON()
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.FechaDomicilio)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Formato de fecha inválido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{
		Direccion: input.Direccion,
		Telefono:  input.Telefono,
		Fecha:     parsedDate,
	}
	if input.Observaciones != nil {
		domicilio.Observ = input.Observaciones
	}
	if input.CreatedBy != nil {
		domicilio.CreatedBy = input.CreatedBy
	}
	if input.Estado != "" {
		if !isValidEstadoDomicilio(string(input.Estado)) {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Campo 'estadoDomicilio' inválido"}
			c.ServeJSON()
			return
		}
		domicilio.Estado = models.EstadoDomicilio(strings.ToUpper(string(input.Estado)))
	}
	if input.TrabajadorID != nil {
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
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualiza los datos de un domicilio existente. El campo 'entregado' es calculado automáticamente.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Domicilio"
// @Param   body  body   models.Domicilio true  "Datos del domicilio a actualizar"
// @Success 200 {object} models.Domicilio "Domicilio actualizado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	o := orm.NewOrm()
	id, err := strconv.ParseInt(c.GetString("id"), 10, 64)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	domicilio := models.Domicilio{ID: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		c.ServeJSON()
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if direccion, ok := input["direccion"].(string); ok {
		domicilio.Direccion = direccion
	}
	if telefono, ok := input["telefono"].(string); ok {
		domicilio.Telefono = telefono
	}
	if updatedBy, ok := input["updatedBy"].(string); ok {
		domicilio.UpdatedBy = &updatedBy
	}
	domicilio.UpdatedAt = time.Now().UTC()

	if _, err := o.Update(&domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
	c.ServeJSON()
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
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
	}
	c.ServeJSON()
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
	domicilio := models.Domicilio{ID: domicilioID}
	if err := o.Read(&domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		c.ServeJSON()
		return
	}

	if domicilio.Trabajador != nil {
		c.Ctx.Output.SetStatus(http.StatusConflict)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusConflict,
			Message: "Este domicilio ya ha sido asignado",
		}
		c.ServeJSON()
		return
	}

	domicilio.Trabajador = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: trabajadorID}
	if _, err := o.Update(&domicilio, "Trabajador"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al asignar domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio asignado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}
