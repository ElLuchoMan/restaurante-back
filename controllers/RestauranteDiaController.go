package controllers

import (
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type RestauranteDiaController struct { web.Controller }

// @Title GetAll
// @Summary Listar días de servicio del restaurante
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param restaurante_id query int false "ID del restaurante"
// @Param dia query string false "Día del enum (Lunes..Domingo)"
// @Success 200 {object} models.ApiResponse{data=[]models.RestauranteDia}
// @Failure 500 {object} models.ApiResponse
// @Router /restaurante_dia [get]
func (c *RestauranteDiaController) GetAll() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.RestauranteDia))
	if rid, err := c.GetInt64("restaurante_id"); err == nil && rid > 0 { qs = qs.Filter("PK_ID_RESTAURANTE", rid) }
	if v := c.GetString("dia"); v != "" { qs = qs.Filter("DIA", v) }
	var list []models.RestauranteDia
	if _, err := qs.All(&list); err != nil { c.Ctx.Output.SetStatus(http.StatusInternalServerError); c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener días", Cause: err.Error()}; c.ServeJSON(); return }
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Días obtenidos", Data: list}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener registro por ID
// @Tags restaurante_dia
// @Accept json
// @Produce json
// @Param id query int true "ID del registro"
// @Success 200 {object} models.ApiResponse{data=models.RestauranteDia}
// @Failure 404 {object} models.ApiResponse
// @Router /restaurante_dia/search [get]
func (c *RestauranteDiaController) GetById() {
	o := orm.NewOrm()
	id, _ := c.GetInt64("id")
	row := models.RestauranteDia{PK_ID_RESTAURANTE_DIA: id}
	if err := o.Read(&row); err != nil { c.Ctx.Output.SetStatus(http.StatusOK); c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Registro no encontrado"}; c.ServeJSON(); return }
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Registro encontrado", Data: row}
	c.ServeJSON()
}
