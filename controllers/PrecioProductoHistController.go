package controllers

import (
	"net/http"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PrecioProductoHistController struct { web.Controller }

// @Title GetAll
// @Summary Listar historial de precios
// @Description Opcional: filtrar por producto y/o fecha_vigencia (YYYY-MM-DD)
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param producto_id query int false "ID del producto"
// @Param fecha query string false "Fecha de vigencia (YYYY-MM-DD)"
// @Success 200 {object} models.ApiResponse{data=[]models.PrecioProductoHist}
// @Failure 500 {object} models.ApiResponse
// @Router /precio_producto_hist [get]
func (c *PrecioProductoHistController) GetAll() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.PrecioProductoHist))
	if pid, err := c.GetInt64("producto_id"); err == nil && pid > 0 {
		qs = qs.Filter("PKIDProducto", pid)
	}
	if f := c.GetString("fecha"); f != "" {
		if d, err := time.Parse("2006-01-02", f); err == nil { qs = qs.Filter("FechaVigencia", d) }
	}
	var list []models.PrecioProductoHist
	if _, err := qs.All(&list); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener historial de precios", Cause: err.Error()}
		c.ServeJSON(); return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Historial de precios", Data: list}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener historial por ID
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param id query int true "ID del historial"
// @Success 200 {object} models.ApiResponse{data=models.PrecioProductoHist}
// @Failure 404 {object} models.ApiResponse
// @Router /precio_producto_hist/search [get]
func (c *PrecioProductoHistController) GetById() {
	o := orm.NewOrm()
	id, _ := c.GetInt64("id")
	row := models.PrecioProductoHist{PK_ID_PRECIO_HIST: id}
	if err := o.Read(&row); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Historial no encontrado"}
		c.ServeJSON(); return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Historial encontrado", Data: row}
	c.ServeJSON()
}
