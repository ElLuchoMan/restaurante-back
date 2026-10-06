package restaurantedia

import (
	"errors"
	"net/http"
	"strings"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type RestauranteDiaController struct{ web.Controller }

const selectDias = `
SELECT rd.pk_id_restaurante_dia AS restaurante_dia_id,
       rd.pk_id_restaurante     AS restaurante_id,
       r.nombre_restaurante     AS nombre_restaurante,
       TO_CHAR(r.hora_apertura, 'HH24:MI:SS') AS hora_apertura,
       rd.dia                   AS dia
FROM restaurante_dia rd
JOIN restaurante r ON r.pk_id_restaurante = rd.pk_id_restaurante`

var quitaAcentos = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u")

// canonicalDia devuelve el valor del enum que corresponde a v (sin distinguir
// mayúsculas ni acentos) o "" si no es un día válido.
func canonicalDia(v string) string {
	n := quitaAcentos.Replace(strings.ToLower(strings.TrimSpace(v)))
	for _, d := range models.DiasSemana {
		if quitaAcentos.Replace(strings.ToLower(d)) == n {
			return d
		}
	}
	return ""
}

// @Title GetAll
// @Summary Listar días de servicio del restaurante
// @Description Devuelve cada fila de restaurante_dia con su id (`restauranteDiaId`), el restaurante y su hora de apertura (HH:MM:SS). Sin resultados, `data` es `[]`. Endpoint público (no exige token).
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param restaurante_id query int false "Filtrar por ID del restaurante (entero positivo)"
// @Param dia query string false "Filtrar por día" Enums(Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo)
// @Success 200 {object} models.ApiResponse{data=[]models.RestauranteDiaView} "Lista de días (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "restaurante_id o dia inválidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurante_dia [get]
func (c *RestauranteDiaController) GetAll() {
	query := selectDias + "\nWHERE 1=1"
	args := []interface{}{}
	if raw := c.GetString("restaurante_id"); raw != "" {
		rid, err := httpx.PositiveInt64Param(&c.Controller, "restaurante_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'restaurante_id' es inválido", err)
			return
		}
		query += " AND rd.pk_id_restaurante = ?"
		args = append(args, rid)
	}
	if raw := c.GetString("dia"); raw != "" {
		dia := canonicalDia(raw)
		if dia == "" {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'dia' es inválido (Lunes, Martes, Miércoles, Jueves, Viernes, Sábado o Domingo)", nil)
			return
		}
		query += " AND rd.dia = ?"
		args = append(args, dia)
	}
	query += "\nORDER BY rd.pk_id_restaurante_dia"
	var rows []models.RestauranteDiaView
	if _, err := orm.NewOrm().Raw(query, args...).QueryRows(&rows); err != nil {
		logging.LogControllerError(c.Ctx, "restaurante_dia.getall.db_error", err, map[string]interface{}{"restaurante_id": c.GetString("restaurante_id"), "dia": c.GetString("dia")})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener días", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Días obtenidos", httpx.List(rows))
}

// @Title GetById
// @Summary Obtener registro de restaurante_dia por ID
// @Description Devuelve la fila de restaurante_dia cuyo `restauranteDiaId` coincide con `id`. Endpoint público (no exige token).
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param id query int true "ID del registro (restauranteDiaId, entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.RestauranteDiaView} "Registro encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Registro no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurante_dia/search [get]
func (c *RestauranteDiaController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	var row models.RestauranteDiaView
	if err := orm.NewOrm().Raw(selectDias+"\nWHERE rd.pk_id_restaurante_dia = ?", id).QueryRow(&row); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Registro no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "restaurante_dia.getbyid.db_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener el registro", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Registro encontrado", row)
}
