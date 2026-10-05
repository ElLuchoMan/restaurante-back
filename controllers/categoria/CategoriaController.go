package categoria

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// CategoriaController gestiona /categorias.
type CategoriaController struct {
	web.Controller
}

const msgIDInvalido = "El parámetro 'id' es inválido o está ausente"

// @Title GetAll
// @Summary Obtener todas las categorías
// @Description Devuelve todas las categorías. Sin resultados, `data` es una lista vacía `[]`.
// @Tags categorias
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.Categoria} "Lista de categorías (puede ser vacía)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /categorias [get]
func (c *CategoriaController) GetAll() {
	var categorias []models.Categoria
	if _, err := orm.NewOrm().QueryTable(new(models.Categoria)).All(&categorias); err != nil {
		logging.LogControllerError(c.Ctx, "categorias.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener categorías", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Categorías obtenidas", httpx.List(categorias))
}

// load valida el query param `id` y lee la categoría; responde el error y
// devuelve false si falla (400 id inválido, 404 inexistente, 500 BD).
func (c *CategoriaController) load(op string) (models.Categoria, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "categorias."+op+".bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return models.Categoria{}, false
	}
	cat := models.Categoria{PK_ID_CATEGORIA: id}
	if err := orm.NewOrm().Read(&cat); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Categoría no encontrada", nil)
			return cat, false
		}
		logging.LogControllerError(c.Ctx, "categorias."+op+".read_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
		return cat, false
	}
	return cat, true
}

// writeError traduce un fallo de escritura: unicidad -> 409, FK -> 409 (en uso)
// y cualquier otro -> 500.
func (c *CategoriaController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "categorias."+op, err, nil)
	switch {
	case dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe una categoría con ese nombre", err)
	case dberr.IsForeignKey(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "La categoría tiene subcategorías o cupones asociados", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
	}
}

// @Title GetById
// @Summary Obtener categoría por ID
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.Categoria} "Categoría encontrada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Categoría no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /categorias/search [get]
func (c *CategoriaController) GetById() {
	cat, ok := c.load("getbyid")
	if !ok {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Categoría encontrada", cat)
}

// @Title Post
// @Summary Crear categoría
// @Description `nombre` es obligatorio y no puede estar vacío.
// @Tags categorias
// @Accept json
// @Produce json
// @Param body body models.CategoriaCreateRequest true "Datos de categoría"
// @Success 201 {object} models.ApiResponse{data=models.Categoria} "Categoría creada"
// @Failure 400 {object} models.ApiResponse "JSON inválido o nombre vacío"
// @Failure 409 {object} models.ApiResponse "Ya existe una categoría con ese nombre"
// @Failure 500 {object} models.ApiResponse "Error al crear la categoría"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Security BearerAuth
// @Router /categorias [post]
func (c *CategoriaController) Post() {
	var in models.CategoriaCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "categorias.post.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
		return
	}
	in.Nombre = strings.TrimSpace(in.Nombre)
	if in.Nombre == "" {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'nombre' es obligatorio", nil)
		return
	}
	cat := models.Categoria{NOMBRE: in.Nombre}
	if _, err := orm.NewOrm().Insert(&cat); err != nil {
		c.writeError("post.insert_error", "Error al crear categoría", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Categoría creada", cat)
}

// @Title Put
// @Summary Actualizar categoría
// @Description Actualización parcial (merge): los campos ausentes se conservan. `nombre` no es anulable (null responde 400) ni puede quedar vacío. Un cuerpo sin cambios responde 200 con la categoría.
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría (entero positivo)"
// @Param body body models.CategoriaUpdateRequest true "Campos a modificar (opcionales, ninguno anulable)"
// @Success 200 {object} models.ApiResponse{data=models.Categoria} "Categoría actualizada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido, JSON inválido, null en campo no anulable o nombre vacío"
// @Failure 404 {object} models.ApiResponse "Categoría no encontrada"
// @Failure 409 {object} models.ApiResponse "Ya existe una categoría con ese nombre"
// @Failure 500 {object} models.ApiResponse "Error al actualizar la categoría"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Security BearerAuth
// @Router /categorias [put]
func (c *CategoriaController) Put() {
	cat, ok := c.load("put")
	if !ok {
		return
	}
	var in models.CategoriaUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "categorias.put.bad_json", err, map[string]interface{}{"id": cat.PK_ID_CATEGORIA})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
		return
	}
	if in.Nombre != nil {
		nombre := strings.TrimSpace(*in.Nombre)
		if nombre == "" {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'nombre' no puede estar vacío", nil)
			return
		}
		cat.NOMBRE = nombre
		if _, err := orm.NewOrm().Update(&cat, "NOMBRE"); err != nil {
			c.writeError("put.update_error", "Error al actualizar categoría", err)
			return
		}
	}
	httpx.Send(&c.Controller, http.StatusOK, "Categoría actualizada", cat)
}

// @Title Delete
// @Summary Eliminar categoría
// @Description Elimina físicamente la categoría. Si tiene subcategorías o cupones asociados responde 409.
// @Tags categorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la categoría (entero positivo)"
// @Success 200 {object} models.ApiResponse "Categoría eliminada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Categoría no encontrada"
// @Failure 409 {object} models.ApiResponse "La categoría tiene elementos asociados"
// @Failure 500 {object} models.ApiResponse "Error al eliminar la categoría"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Security BearerAuth
// @Router /categorias [delete]
func (c *CategoriaController) Delete() {
	cat, ok := c.load("delete")
	if !ok {
		return
	}
	// Raw y no orm.Delete: el ORM borraría en cascada subcategorías y productos
	// por su cuenta; así la llave foránea de la BD protege los datos (409).
	if _, err := orm.NewOrm().Raw("DELETE FROM categoria WHERE pk_id_categoria = ?", cat.PK_ID_CATEGORIA).Exec(); err != nil {
		c.writeError("delete.delete_error", "Error al eliminar categoría", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Categoría eliminada", nil)
}
