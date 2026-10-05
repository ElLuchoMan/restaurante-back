// Package restaurante expone la consulta (solo lectura) del restaurante.
//
// Decisión de producto: el restaurante es único y se administra directamente
// en la base de datos, por eso no existen endpoints POST/PUT/DELETE (antes
// había handlers sin ruta; se eliminaron como código muerto).
package restaurante

import (
	"errors"
	"net/http"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type RestauranteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los restaurantes
// @Description Devuelve los restaurantes registrados. `horaApertura` va como HH:MM:SS y `cambioHorarioId` (si existe) como objeto de cambio de horario. Sin resultados, `data` es `[]`. Endpoint público (no exige token).
// @Tags restaurantes
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.RestauranteDoc} "Lista de restaurantes (puede ser vacía)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurantes [get]
func (c *RestauranteController) GetAll() {
	var restaurantes []models.Restaurante
	if _, err := orm.NewOrm().QueryTable(new(models.Restaurante)).All(&restaurantes); err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener restaurantes de la base de datos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Restaurantes obtenidos exitosamente", httpx.List(restaurantes))
}

// @Title GetById
// @Summary Obtener restaurante por ID
// @Description Devuelve un restaurante por ID (query param `id`). Endpoint público (no exige token).
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del restaurante (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.RestauranteDoc} "Restaurante encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurantes/search [get]
func (c *RestauranteController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "restaurantes.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	restaurante := models.Restaurante{PK_ID_RESTAURANTE: id}
	if err := orm.NewOrm().Read(&restaurante); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Restaurante no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "restaurantes.getbyid.db_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al consultar el restaurante", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Restaurante encontrado", restaurante)
}
