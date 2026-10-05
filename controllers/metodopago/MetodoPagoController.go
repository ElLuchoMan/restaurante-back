package metodopago

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type MetodoPagoController struct {
	web.Controller
}

const msgIDInvalido = "El parámetro 'id' es inválido o está ausente"

// @Title GetAll
// @Summary Obtener todos los métodos de pago
// @Description Devuelve todos los métodos de pago. Si no hay ninguno, `data` es una lista vacía `[]`.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.MetodoPago} "Lista de métodos de pago (puede ser vacía)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /metodos_pago [get]
func (c *MetodoPagoController) GetAll() {
	var metodos []models.MetodoPago
	if _, err := orm.NewOrm().QueryTable(new(models.MetodoPago)).All(&metodos); err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener métodos de pago de la base de datos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Métodos de pago obtenidos exitosamente", httpx.List(metodos))
}

// @Title GetById
// @Summary Obtener método de pago por ID
// @Description Devuelve un método de pago por ID (query param `id`).
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del método de pago (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /metodos_pago/search [get]
func (c *MetodoPagoController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	metodo := models.MetodoPago{PK_ID_METODO_PAGO: id}
	if err := orm.NewOrm().Read(&metodo); err != nil {
		c.readError("getbyid", id, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Método de pago encontrado", metodo)
}

// readError traduce un fallo de Read: ErrNoRows -> 404, cualquier otro -> 500.
func (c *MetodoPagoController) readError(op string, id int64, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Método de pago no encontrado", nil)
		return
	}
	logging.LogControllerError(c.Ctx, "metodos_pago."+op+".db_error", err, map[string]interface{}{"id": id})
	httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al consultar el método de pago", err)
}

// @Title Create
// @Summary Crear un nuevo método de pago
// @Description Crea un método de pago. `tipo` es obligatorio y no puede estar vacío; `detalle` es opcional (por defecto cadena vacía).
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   body  body   models.MetodoPagoCreateRequest true  "Datos del método de pago a crear"
// @Success 201 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago creado"
// @Failure 400 {object} models.ApiResponse "JSON inválido o `tipo` vacío"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad en base de datos"
// @Failure 500 {object} models.ApiResponse "Error al crear el método de pago"
// @Security BearerAuth
// @Router /metodos_pago [post]
func (c *MetodoPagoController) Post() {
	var in models.MetodoPagoCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.post.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Error en la solicitud", err)
		return
	}
	in.Tipo = strings.TrimSpace(in.Tipo)
	if in.Tipo == "" {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'tipo' es obligatorio", nil)
		return
	}
	metodo := models.MetodoPago{TIPO: in.Tipo}
	if in.Detalle != nil {
		metodo.DETALLE = *in.Detalle
	}
	if _, err := orm.NewOrm().Insert(&metodo); err != nil {
		c.writeError("post.insert_error", "Error al crear el método de pago", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Método de pago creado correctamente", metodo)
}

// writeError responde 409 si el error es un conflicto de PostgreSQL y 500 en otro caso.
func (c *MetodoPagoController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "metodos_pago."+op, err, nil)
	if httpx.IsPGConflict(err) {
		httpx.Fail(&c.Controller, http.StatusConflict, msg+": conflicto con datos existentes", err)
		return
	}
	httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
}

// @Title Update
// @Summary Actualizar un método de pago
// @Description Actualización parcial (merge): los campos ausentes del cuerpo se conservan. `tipo` y `detalle` no son anulables: enviar `null` en cualquiera responde 400. `tipo` no puede quedar vacío.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del método de pago (entero positivo)"
// @Param   body  body   models.MetodoPagoUpdateRequest true  "Campos a modificar (todos opcionales, ninguno anulable)"
// @Success 200 {object} models.ApiResponse{data=models.MetodoPago} "Método de pago actualizado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido, JSON inválido, null en campo no anulable o `tipo` vacío"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad en base de datos"
// @Failure 500 {object} models.ApiResponse "Error al actualizar el método de pago"
// @Security BearerAuth
// @Router /metodos_pago [put]
func (c *MetodoPagoController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	metodo := models.MetodoPago{PK_ID_METODO_PAGO: id}
	if err := o.Read(&metodo); err != nil {
		c.readError("put", id, err)
		return
	}
	var in models.MetodoPagoUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.put.bad_json", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Error en la solicitud", err)
		return
	}
	if in.Tipo != nil {
		tipo := strings.TrimSpace(*in.Tipo)
		if tipo == "" {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'tipo' no puede estar vacío", nil)
			return
		}
		metodo.TIPO = tipo
	}
	if in.Detalle != nil {
		metodo.DETALLE = *in.Detalle
	}
	if _, err := o.Update(&metodo, "TIPO", "DETALLE"); err != nil {
		c.writeError("put.update_error", "Error al actualizar el método de pago", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Método de pago actualizado", metodo)
}

// @Title Delete
// @Summary Eliminar un método de pago
// @Description Elimina un método de pago. Si está referenciado por pagos responde 409.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del método de pago (entero positivo)"
// @Success 200 {object} models.ApiResponse "Método de pago eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Failure 409 {object} models.ApiResponse "El método de pago está en uso por pagos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /metodos_pago [delete]
func (c *MetodoPagoController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "metodos_pago.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	n, err := orm.NewOrm().Delete(&models.MetodoPago{PK_ID_METODO_PAGO: id})
	if err != nil {
		c.writeError("delete.delete_error", "Error al eliminar el método de pago", err)
		return
	}
	if n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Método de pago no encontrado", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Método de pago eliminado", nil)
}
