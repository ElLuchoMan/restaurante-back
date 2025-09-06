package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type CategoriaController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las categorías
// @Tags categorias
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.Categoria}
// @Failure 500 {object} models.ApiResponse
// @Router /categorias [get]
func (c *CategoriaController) GetAll() {
	o := orm.NewOrm()
	var categorias []models.Categoria
	if _, err := o.QueryTable(new(models.Categoria)).All(&categorias); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener categorías", Cause: err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Categorías obtenidas", Data: categorias}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener categoría por ID
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría"
// @Success 200 {object} models.ApiResponse{data=models.Categoria}
// @Failure 404 {object} models.ApiResponse
// @Router /categorias/search [get]
func (c *CategoriaController) GetById() {
	o := orm.NewOrm()
	id, _ := c.GetInt64("id")
	cat := models.Categoria{PK_ID_CATEGORIA: id}
	if err := o.Read(&cat); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Categoría no encontrada"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Categoría encontrada", Data: cat}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear categoría
// @Tags categorias
// @Accept json
// @Produce json
// @Param body body models.CategoriaCreateRequest true "Datos de categoría"
// @Success 201 {object} models.ApiResponse{data=models.Categoria}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /categorias [post]
func (c *CategoriaController) Post() {
	o := orm.NewOrm()
	var in struct{
		Nombre string `json:"nombre"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil || in.Nombre == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido o nombre requerido"}
		c.ServeJSON()
		return
	}
	cat := models.Categoria{NOMBRE: in.Nombre}
	if _, err := o.Insert(&cat); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear categoría", Cause: err.Error()}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Categoría creada", Data: cat}
	c.ServeJSON()
}

// @Title Put
// @Summary Actualizar categoría
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría"
// @Param body body models.CategoriaUpdateRequest true "Datos a actualizar"
// @Success 200 {object} models.ApiResponse{data=models.Categoria}
// @Failure 404 {object} models.ApiResponse
// @Router /categorias [put]
func (c *CategoriaController) Put() {
	o := orm.NewOrm()
	id, _ := c.GetInt64("id")
	cat := models.Categoria{PK_ID_CATEGORIA: id}
	if err := o.Read(&cat); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Categoría no encontrada"}
		c.ServeJSON()
		return
	}
	var in struct{ Nombre *string `json:"nombre"` }
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
		c.ServeJSON()
		return
	}
	if in.Nombre != nil { cat.NOMBRE = *in.Nombre }
	if _, err := o.Update(&cat, "NOMBRE"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar categoría", Cause: err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Categoría actualizada", Data: cat}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar categoría
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /categorias [delete]
func (c *CategoriaController) Delete() {
	o := orm.NewOrm()
	id, _ := c.GetInt64("id")
	if _, err := o.Delete(&models.Categoria{PK_ID_CATEGORIA: id}); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Categoría no encontrada"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Categoría eliminada"}
	c.ServeJSON()
}
