package controllers

import (
	"net/http"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type restDiaOrmer interface {
	Raw(string, ...interface{}) orm.RawSeter
}

var restDiaOrmNew = func() restDiaOrmer { return orm.NewOrm() }

type RestauranteDiaController struct{ web.Controller }

// @Title GetAll
// @Summary Listar días de servicio del restaurante
// @Description Devuelve restauranteId, nombreRestaurante, horaApertura y dia.
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param restaurante_id query int false "ID del restaurante"
// @Param dia query string false "Día del enum (Lunes..Domingo)"
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /restaurante_dia [get]
func (c *RestauranteDiaController) GetAll() {
	o := restDiaOrmNew()
	query := `
SELECT rd.pk_id_restaurante    AS restaurante_id,
       r.nombre_restaurante    AS nombre_restaurante,
       TO_CHAR(r.hora_apertura, 'HH24:MI:SS') AS hora_apertura,
       rd.dia                  AS dia
FROM restaurante_dia rd
JOIN restaurante r ON r.pk_id_restaurante = rd.pk_id_restaurante
WHERE 1=1`
	args := []interface{}{}
	if rid, err := c.GetInt64("restaurante_id"); err == nil && rid > 0 {
		query += " AND rd.pk_id_restaurante = ?"
		args = append(args, rid)
	}
	if v := c.GetString("dia"); v != "" {
		query += " AND rd.dia = ?"
		args = append(args, v)
	}
	var rows []struct {
		RestID int64  `orm:"column(restaurante_id)"`
		Nombre string `orm:"column(nombre_restaurante)"`
		Hora   string `orm:"column(hora_apertura)"`
		Dia    string `orm:"column(dia)"`
	}
	if _, err := o.Raw(query, args...).QueryRows(&rows); err != nil {
		logging.LogControllerError(c.Ctx, "restaurante_dia.getall.db_error", err, map[string]interface{}{"restaurante_id": c.GetString("restaurante_id"), "dia": c.GetString("dia")})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener días", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	resp := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, map[string]interface{}{
			"restauranteId":     r.RestID,
			"nombreRestaurante": r.Nombre,
			"horaApertura":      r.Hora,
			"dia":               r.Dia,
		})
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Días obtenidos", Data: resp}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener registro por ID
// @Description Devuelve restauranteId, nombreRestaurante, horaApertura y dia.
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param id query int true "ID del registro"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /restaurante_dia/search [get]
func (c *RestauranteDiaController) GetById() {
	o := restDiaOrmNew()
	id, _ := c.GetInt64("id")
	query := `
SELECT rd.pk_id_restaurante    AS restaurante_id,
       r.nombre_restaurante    AS nombre_restaurante,
       TO_CHAR(r.hora_apertura, 'HH24:MI:SS') AS hora_apertura,
       rd.dia                  AS dia
FROM restaurante_dia rd
JOIN restaurante r ON r.pk_id_restaurante = rd.pk_id_restaurante
WHERE rd.pk_id_restaurante_dia = ?`
	var row struct {
		RestID int64  `orm:"column(restaurante_id)"`
		Nombre string `orm:"column(nombre_restaurante)"`
		Hora   string `orm:"column(hora_apertura)"`
		Dia    string `orm:"column(dia)"`
	}
	if err := o.Raw(query, id).QueryRow(&row); err != nil {
		if err != orm.ErrNoRows {
			logging.LogControllerError(c.Ctx, "restaurante_dia.getbyid.db_error", err, map[string]interface{}{"id": id})
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Registro no encontrado"}
		_ = c.ServeJSON()
		return
	}
	resp := map[string]interface{}{
		"restauranteId":     row.RestID,
		"nombreRestaurante": row.Nombre,
		"horaApertura":      row.Hora,
		"dia":               row.Dia,
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Registro encontrado", Data: resp}
	_ = c.ServeJSON()
}
