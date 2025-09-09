package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type subcatQuerySeter interface {
	All(interface{}, ...string) (int64, error)
	Filter(string, ...interface{}) subcatQuerySeter
}

type subcatOrmer interface {
	QueryTable(interface{}) subcatQuerySeter
	Insert(interface{}) (int64, error)
	Read(interface{}, ...string) error
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

type subQSAdapter struct{ qs orm.QuerySeter }

func (a subQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}
func (a subQSAdapter) Filter(field string, args ...interface{}) subcatQuerySeter {
	return subQSAdapter{qs: a.qs.Filter(field, args...)}
}

type subOrmAdapter struct{ o orm.Ormer }

func (a subOrmAdapter) QueryTable(i interface{}) subcatQuerySeter {
	return subQSAdapter{qs: a.o.QueryTable(i)}
}
func (a subOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a subOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a subOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a subOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var subcatOrmNew = func() subcatOrmer { return subOrmAdapter{o: orm.NewOrm()} }

type SubcategoriaController struct{ web.Controller }

// @Title GetAll
// @Summary Obtener todas las subcategorías
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param categoria_id query int false "Filtrar por categoría"
// @Success 200 {object} models.ApiResponse{data=[]models.Subcategoria}
// @Failure 500 {object} models.ApiResponse
// @Router /subcategorias [get]
func (c *SubcategoriaController) GetAll() {
	o := subcatOrmNew()
	qs := o.QueryTable(new(models.Subcategoria))
	if catID, err := c.GetInt64("categoria_id"); err == nil && catID > 0 {
		qs = qs.Filter("PK_ID_CATEGORIA", catID)
	}
	var subs []models.Subcategoria
	if _, err := qs.All(&subs); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.getall.db_error", err, map[string]interface{}{"categoria_id": c.GetString("categoria_id")})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener subcategorías", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Subcategorías obtenidas", Data: subs}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener subcategoría por ID
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría"
// @Success 200 {object} models.ApiResponse{data=models.Subcategoria}
// @Failure 404 {object} models.ApiResponse
// @Router /subcategorias/search [get]
func (c *SubcategoriaController) GetById() {
	o := subcatOrmNew()
	id, _ := c.GetInt64("id")
	s := models.Subcategoria{PK_ID_SUBCATEGORIA: id}
	if err := o.Read(&s); err != nil {
		if err != orm.ErrNoRows {
			logging.LogControllerError(c.Ctx, "subcategorias.getbyid.db_error", err, map[string]interface{}{"id": id})
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Subcategoría no encontrada"}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Subcategoría encontrada", Data: s}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear subcategoría
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param body body models.SubcategoriaCreateRequest true "Datos de subcategoría"
// @Success 201 {object} models.ApiResponse{data=models.Subcategoria}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /subcategorias [post]
func (c *SubcategoriaController) Post() {
	o := subcatOrmNew()
	var in struct {
		Nombre      string `json:"nombre"`
		CategoriaId int64  `json:"categoriaId"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil || in.Nombre == "" || in.CategoriaId == 0 {
		if err != nil {
			logging.LogControllerError(c.Ctx, "subcategorias.post.bad_json", err, nil)
		}
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido o campos requeridos faltantes"}
		_ = c.ServeJSON()
		return
	}
	s := models.Subcategoria{NOMBRE: in.Nombre, PK_ID_CATEGORIA: &models.Categoria{PK_ID_CATEGORIA: in.CategoriaId}}
	if _, err := o.Insert(&s); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.post.insert_error", err, map[string]interface{}{"nombre": in.Nombre, "categoriaId": in.CategoriaId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear subcategoría", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Subcategoría creada", Data: s}
	_ = c.ServeJSON()
}

// @Title Put
// @Summary Actualizar subcategoría
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría"
// @Param body body models.SubcategoriaUpdateRequest true "Datos a actualizar"
// @Success 200 {object} models.ApiResponse{data=models.Subcategoria}
// @Failure 404 {object} models.ApiResponse
// @Router /subcategorias [put]
func (c *SubcategoriaController) Put() {
	o := subcatOrmNew()
	id, _ := c.GetInt64("id")
	s := models.Subcategoria{PK_ID_SUBCATEGORIA: id}
	if err := o.Read(&s); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Subcategoría no encontrada"}
		_ = c.ServeJSON()
		return
	}
	var in struct {
		Nombre      *string `json:"nombre"`
		CategoriaId *int64  `json:"categoriaId"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	cols := []string{}
	if in.Nombre != nil {
		s.NOMBRE = *in.Nombre
		cols = append(cols, "NOMBRE")
	}
	if in.CategoriaId != nil {
		s.PK_ID_CATEGORIA = &models.Categoria{PK_ID_CATEGORIA: *in.CategoriaId}
		cols = append(cols, "PK_ID_CATEGORIA")
	}
	if _, err := o.Update(&s, cols...); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar subcategoría", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Subcategoría actualizada", Data: s}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar subcategoría
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /subcategorias [delete]
func (c *SubcategoriaController) Delete() {
	o := subcatOrmNew()
	id, _ := c.GetInt64("id")
	if _, err := o.Delete(&models.Subcategoria{PK_ID_SUBCATEGORIA: id}); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Subcategoría no encontrada"}
		_ = c.ServeJSON()
		return
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Subcategoría eliminada"}
	_ = c.ServeJSON()
}
