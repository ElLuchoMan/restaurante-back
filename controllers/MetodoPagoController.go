package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type metodoPagoQuerySeter interface {
	All(interface{}, ...string) (int64, error)
}

type metodoPagoOrmer interface {
	QueryTable(interface{}) metodoPagoQuerySeter
	Read(interface{}, ...string) error
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

type defaultQuerySeter struct {
	orm.QuerySeter
}

func (d *defaultQuerySeter) All(res interface{}, cols ...string) (int64, error) {
	return d.QuerySeter.All(res, cols...)
}

type defaultOrmer struct {
	orm.Ormer
}

func (d *defaultOrmer) QueryTable(v interface{}) metodoPagoQuerySeter {
	return &defaultQuerySeter{d.Ormer.QueryTable(v)}
}

var getOrm func() metodoPagoOrmer = func() metodoPagoOrmer { return &defaultOrmer{orm.NewOrm()} }

type MetodoPagoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los métodos de pago
// @Description Devuelve todos los métodos de pago registrados en la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.MetodoPago} "Lista de métodos de pago"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /metodos_pago [get]
func (c *MetodoPagoController) GetAll() {
	o := getOrm()
	var metodos []models.MetodoPago
	_, err := o.QueryTable(new(models.MetodoPago)).All(&metodos)
	if err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener métodos de pago de la base de datos", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Métodos de pago obtenidos exitosamente", Data: metodos}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener método de pago por ID
// @Description Devuelve un método de pago específico por ID utilizando query parameters.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Método de Pago"
// @Success 200 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago encontrado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Security BearerAuth
// @Router /metodos_pago/search [get]
func (c *MetodoPagoController) GetById() {
	o := getOrm()
	id, err := c.GetInt("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "metodos_pago.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	metodo := models.MetodoPago{PK_ID_METODO_PAGO: int64(id)}
	err = o.Read(&metodo)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Método de pago no encontrado", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Método de pago encontrado", Data: metodo}
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo método de pago
// @Description Crea un nuevo método de pago en la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   body  body   models.MetodoPagoCreateRequest true  "Datos del método de pago a crear"
// @Success 201 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /metodos_pago [post]
func (c *MetodoPagoController) Post() {
	o := getOrm()
	var metodo models.MetodoPago
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &metodo); err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error en la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	if _, err := o.Insert(&metodo); err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.post.insert_error", err, map[string]interface{}{"tipo": metodo.TIPO, "detalle": metodo.DETALLE})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el método de pago", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Método de pago creado correctamente", Data: metodo}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un método de pago
// @Description Actualiza los datos de un método de pago existente.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Método de Pago"
// @Param   body  body   models.MetodoPagoUpdateRequest true  "Datos del método de pago a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago actualizado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Security BearerAuth
// @Router /metodos_pago [put]
func (c *MetodoPagoController) Put() {
	o := getOrm()
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "metodos_pago.put.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	metodo := models.MetodoPago{PK_ID_METODO_PAGO: int64(id)}
	if o.Read(&metodo) == nil {
		var updatedMetodo models.MetodoPago
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedMetodo); err != nil {
			logging.LogControllerError(c.Ctx, "metodos_pago.put.bad_json", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error en la solicitud", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		updatedMetodo.PK_ID_METODO_PAGO = int64(id)
		if _, err := o.Update(&updatedMetodo); err != nil {
			logging.LogControllerError(c.Ctx, "metodos_pago.put.update_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el método de pago", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Método de pago actualizado", Data: updatedMetodo}
		_ = c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Método de pago no encontrado"}
		_ = c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un método de pago
// @Description Elimina un método de pago de la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Método de Pago"
// @Success 200 {object} models.ApiResponse "Método de pago eliminado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Security BearerAuth
// @Router /metodos_pago [delete]
func (c *MetodoPagoController) Delete() {
	o := getOrm()
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "metodos_pago.delete.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	metodo := models.MetodoPago{PK_ID_METODO_PAGO: int64(id)}
	if _, err := o.Delete(&metodo); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Método de pago eliminado"}
		_ = c.ServeJSON()
	} else {
		logging.LogControllerError(c.Ctx, "metodos_pago.delete.delete_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Método de pago no encontrado", Cause: err.Error()}
		_ = c.ServeJSON()
	}
}
