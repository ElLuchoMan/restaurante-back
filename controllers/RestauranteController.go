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

type RestauranteController struct {
	web.Controller
}

// Puntos de inyección para pruebas (permiten mockear el ORM en tests)
type restQuerySeter interface {
	All(interface{}, ...string) (int64, error)
}
type restOrmer interface {
	QueryTable(interface{}) restQuerySeter
	Read(interface{}, ...string) error
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}
type restQSAdapter struct{ qs orm.QuerySeter }

func (a restQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}

type restOrmAdapter struct{ o orm.Ormer }

func (a restOrmAdapter) QueryTable(i interface{}) restQuerySeter {
	return restQSAdapter{qs: a.o.QueryTable(i)}
}
func (a restOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a restOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a restOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a restOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var restOrmNew = func() restOrmer { return restOrmAdapter{o: orm.NewOrm()} }

// @Title GetAll
// @Summary Obtener todos los restaurantes
// @Description Devuelve todos los restaurantes registrados en la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.Restaurante} "Lista de restaurantes"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurantes [get]
func (c *RestauranteController) GetAll() {
	o := restOrmNew()
	var restaurantes []models.Restaurante

	_, err := o.QueryTable(new(models.Restaurante)).All(&restaurantes)
	if err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener restaurantes de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Restaurantes obtenidos exitosamente",
		Data:    restaurantes,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener restaurante por ID
// @Description Devuelve un restaurante específico por ID utilizando query parameters.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Restaurante"
// @Success 200 {object} models.ApiResponse{data=models.Restaurante} "Restaurante encontrado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes/search [get]
func (c *RestauranteController) GetById() {
	o := restOrmNew()
	id, err := c.GetInt("id")

	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "restaurantes.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: int64(id)}

	err = o.Read(&restaurante)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
		}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.getbyid.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Restaurante encontrado",
			Data:    restaurante,
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Restaurante encontrado",
		Data:    restaurante,
	}
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo restaurante
// @Description Crea un nuevo restaurante en la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   body  body   models.RestauranteCreateRequest true  "Datos del restaurante a crear"
// @Success 201 {object} models.ApiResponse{data=models.Restaurante} "Restaurante creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /restaurantes [post]
func (c *RestauranteController) Post() {
	o := restOrmNew()
	var restaurante models.Restaurante

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &restaurante); err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	_, err := o.Insert(&restaurante)
	if err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.post.insert_error", err, map[string]interface{}{"nombre": restaurante.NOMBRE_RESTAURANTE})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el restaurante",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Restaurante creado correctamente",
		Data:    restaurante,
	}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un restaurante
// @Description Actualiza los datos de un restaurante existente.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Restaurante"
// @Param   body  body   models.RestauranteUpdateRequest true  "Datos del restaurante a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse{data=models.Restaurante} "Restaurante actualizado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes [put]
func (c *RestauranteController) Put() {
	o := restOrmNew()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "restaurantes.put.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: int64(id)}

	if o.Read(&restaurante) == nil {
		var updatedRestaurante models.Restaurante
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedRestaurante); err != nil {
			logging.LogControllerError(c.Ctx, "restaurantes.put.bad_json", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}

		updatedRestaurante.PK_ID_RESTAURANTE = int64(id)

		_, err := o.Update(&updatedRestaurante)
		if err != nil {
			logging.LogControllerError(c.Ctx, "restaurantes.put.update_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el restaurante",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Restaurante actualizado",
			Data:    updatedRestaurante,
		}
		_ = c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
		}
		_ = c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un restaurante
// @Description Elimina un restaurante de la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Restaurante"
// @Success 204 {object} nil "Restaurante eliminado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes [delete]
func (c *RestauranteController) Delete() {
	o := restOrmNew()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "restaurantes.delete.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: int64(id)}

	if _, err := o.Delete(&restaurante); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Restaurante eliminado",
		}
		_ = c.ServeJSON()
	} else {
		logging.LogControllerError(c.Ctx, "restaurantes.delete.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
	}
}
