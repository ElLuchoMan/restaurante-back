package controllers

import (
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type resContactoOrmer interface{ QueryTable(interface{}) orm.QuerySeter }
var resContactoOrmNew = func() resContactoOrmer { return orm.NewOrm() }

type ReservaContactoController struct { web.Controller }

// @Title GetAll
// @Summary Listar contactos de reserva
// @Tags reserva_contacto
// @Accept json
// @Produce json
// @Param documento_contacto query int false "Documento del contacto"
// @Param documento_cliente query int false "Documento del cliente"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaContacto}
// @Failure 500 {object} models.ApiResponse
// @Router /reserva_contacto [get]
func (c *ReservaContactoController) GetAll() {
	o := resContactoOrmNew()
	qs := o.QueryTable(new(models.ReservaContacto))
	if v, err := c.GetInt64("documento_contacto"); err == nil && v > 0 { qs = qs.Filter("DocumentoContacto", v) }
	if v, err := c.GetInt64("documento_cliente"); err == nil && v > 0 { qs = qs.Filter("PKDocumentoCliente", v) }
	var list []models.ReservaContacto
	if _, err := qs.All(&list); err != nil { c.Ctx.Output.SetStatus(http.StatusInternalServerError); c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener contactos", Cause: err.Error()}; c.ServeJSON(); return }
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Contactos obtenidos", Data: list}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener contacto por ID
// @Tags reserva_contacto
// @Accept json
// @Produce json
// @Param id query int true "ID del contacto"
// @Success 200 {object} models.ApiResponse{data=models.ReservaContacto}
// @Failure 404 {object} models.ApiResponse
// @Router /reserva_contacto/search [get]
func (c *ReservaContactoController) GetById() {
	o := resContactoOrmNew()
	id, _ := c.GetInt64("id")
	row := models.ReservaContacto{PKIDContacto: id}
	if err := o.QueryTable(new(models.ReservaContacto)).Filter("PKIDContacto", id).One(&row); err != nil { c.Ctx.Output.SetStatus(http.StatusOK); c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Contacto no encontrado"}; c.ServeJSON(); return }
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Contacto encontrado", Data: row}
	c.ServeJSON()
}
