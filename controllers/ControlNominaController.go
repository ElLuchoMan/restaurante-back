package controllers

import (
	"net/http"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ctrlNomOrmer interface {
	Raw(string, ...interface{}) orm.RawSeter
	QueryTable(interface{}) orm.QuerySeter
}

var ctrlNomOrmNew = func() ctrlNomOrmer { return orm.NewOrm() }

type ControlNominaController struct{ web.Controller }

// @Title GetAll
// @Summary Listar control de nómina
// @Description Opcional: filtrar por fecha (YYYY-MM-DD)
// @Tags control_nomina
// @Accept json
// @Produce json
// @Param fecha query string false "Fecha (YYYY-MM-DD)"
// @Success 200 {object} models.ApiResponse{data=[]models.ControlNomina}
// @Failure 500 {object} models.ApiResponse
// @Router /control_nomina [get]
func (c *ControlNominaController) GetAll() {
	o := ctrlNomOrmNew()
	qs := o.QueryTable(new(models.ControlNomina))
	if f := c.GetString("fecha"); f != "" {
		if d, err := time.Parse("2006-01-02", f); err == nil {
			qs = qs.Filter("Fecha", d)
		}
	}
	var list []models.ControlNomina
	if _, err := qs.All(&list); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener control de nómina", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Control de nómina", Data: list}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener control de nómina por ID
// @Tags control_nomina
// @Accept json
// @Produce json
// @Param id query int true "ID del control"
// @Success 200 {object} models.ApiResponse{data=models.ControlNomina}
// @Failure 404 {object} models.ApiResponse
// @Router /control_nomina/search [get]
func (c *ControlNominaController) GetById() {
	o := ctrlNomOrmNew()
	id, _ := c.GetInt64("id")
	row := models.ControlNomina{PK_ID_CONTROL_NOMINA: id}
	if err := o.QueryTable(new(models.ControlNomina)).Filter("PK_ID_CONTROL_NOMINA", id).One(&row); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Registro no encontrado"}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Registro encontrado", Data: row}
	_ = c.ServeJSON()
}
