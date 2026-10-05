package subcategoria

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

// SubcategoriaController gestiona /subcategorias. En todas las respuestas
// `categoriaId` es el objeto categoría completo ({categoriaId, nombre}).
type SubcategoriaController struct{ web.Controller }

const msgIDInvalido = "El parámetro 'id' es inválido o está ausente"

// porID lee la subcategoría junto con su categoría (RelatedSel).
func porID(id int64) (models.Subcategoria, error) {
	var s models.Subcategoria
	err := orm.NewOrm().QueryTable(new(models.Subcategoria)).Filter("PK_ID_SUBCATEGORIA", id).RelatedSel("PK_ID_CATEGORIA").One(&s)
	return s, err
}

// load valida `id` y lee la subcategoría (400 / 404 / 500).
func (c *SubcategoriaController) load(op string) (models.Subcategoria, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias."+op+".bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return models.Subcategoria{}, false
	}
	s, err := porID(id)
	if err != nil {
		c.readError(op, id, err)
		return s, false
	}
	return s, true
}

func (c *SubcategoriaController) readError(op string, id int64, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Subcategoría no encontrada", nil)
		return
	}
	logging.LogControllerError(c.Ctx, "subcategorias."+op+".read_error", err, map[string]interface{}{"id": id})
	httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
}

// writeError: unicidad -> 409; FK en insert/update (categoría inexistente) ->
// 400; FK en delete (tiene productos o cupones) -> 409; otro -> 500.
func (c *SubcategoriaController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "subcategorias."+op, err, nil)
	switch {
	case dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe una subcategoría con esos datos", err)
	case dberr.IsForeignKey(err) && strings.HasPrefix(op, "delete"):
		httpx.Fail(&c.Controller, http.StatusConflict, "La subcategoría tiene productos asociados", err)
	case dberr.IsForeignKey(err):
		httpx.Fail(&c.Controller, http.StatusBadRequest, "La categoría indicada no existe", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
	}
}

// @Title GetAll
// @Summary Obtener todas las subcategorías
// @Description Devuelve las subcategorías, opcionalmente filtradas por categoría. Cada una trae `categoriaId` como objeto categoría. Sin resultados, `data` es `[]`.
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param categoria_id query int false "Filtrar por categoría (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.Subcategoria} "Lista de subcategorías (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "categoria_id inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /subcategorias [get]
func (c *SubcategoriaController) GetAll() {
	qs := orm.NewOrm().QueryTable(new(models.Subcategoria)).RelatedSel("PK_ID_CATEGORIA")
	if strings.TrimSpace(c.GetString("categoria_id")) != "" {
		catID, err := httpx.PositiveInt64Param(&c.Controller, "categoria_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'categoria_id' es inválido", err)
			return
		}
		qs = qs.Filter("PK_ID_CATEGORIA", catID)
	}
	var subs []models.Subcategoria
	if _, err := qs.All(&subs); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.getall.db_error", err, map[string]interface{}{"categoria_id": c.GetString("categoria_id")})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener subcategorías", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Subcategorías obtenidas", httpx.List(subs))
}

// @Title GetById
// @Summary Obtener subcategoría por ID
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.Subcategoria} "Subcategoría encontrada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Subcategoría no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /subcategorias/search [get]
func (c *SubcategoriaController) GetById() {
	s, ok := c.load("getbyid")
	if !ok {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Subcategoría encontrada", s)
}

// @Title Post
// @Summary Crear subcategoría
// @Description `nombre` y `categoriaId` (entero positivo de una categoría existente) son obligatorios.
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param body body models.SubcategoriaCreateRequest true "Datos de subcategoría"
// @Success 201 {object} models.ApiResponse{data=models.Subcategoria} "Subcategoría creada"
// @Failure 400 {object} models.ApiResponse "JSON inválido, campos faltantes o categoría inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad"
// @Failure 500 {object} models.ApiResponse "Error al crear la subcategoría"
// @Security BearerAuth
// @Router /subcategorias [post]
func (c *SubcategoriaController) Post() {
	var in models.SubcategoriaCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.post.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
		return
	}
	in.Nombre = strings.TrimSpace(in.Nombre)
	if in.Nombre == "" || in.CategoriaId <= 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Los campos 'nombre' y 'categoriaId' son obligatorios", nil)
		return
	}
	s := models.Subcategoria{NOMBRE: in.Nombre, PK_ID_CATEGORIA: &models.Categoria{PK_ID_CATEGORIA: in.CategoriaId}}
	if _, err := orm.NewOrm().Insert(&s); err != nil {
		c.writeError("post.insert_error", "Error al crear subcategoría", err)
		return
	}
	creada, err := porID(s.PK_ID_SUBCATEGORIA)
	if err != nil {
		c.readError("post", s.PK_ID_SUBCATEGORIA, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Subcategoría creada", creada)
}

// @Title Put
// @Summary Actualizar subcategoría
// @Description Actualización parcial (merge): los campos ausentes se conservan. Ningún campo admite null (400); `nombre` no puede quedar vacío y `categoriaId` debe ser una categoría existente. Un cuerpo sin cambios responde 200.
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría (entero positivo)"
// @Param body body models.SubcategoriaUpdateRequest true "Campos a modificar (opcionales, ninguno anulable)"
// @Success 200 {object} models.ApiResponse{data=models.Subcategoria} "Subcategoría actualizada"
// @Failure 400 {object} models.ApiResponse "id inválido, JSON inválido, null en campo no anulable, validación o categoría inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Subcategoría no encontrada"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad"
// @Failure 500 {object} models.ApiResponse "Error al actualizar la subcategoría"
// @Security BearerAuth
// @Router /subcategorias [put]
func (c *SubcategoriaController) Put() {
	s, ok := c.load("put")
	if !ok {
		return
	}
	var in models.SubcategoriaUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		logging.LogControllerError(c.Ctx, "subcategorias.put.bad_json", err, map[string]interface{}{"id": s.PK_ID_SUBCATEGORIA})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
		return
	}
	cols := []string{}
	if in.Nombre != nil {
		nombre := strings.TrimSpace(*in.Nombre)
		if nombre == "" {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'nombre' no puede estar vacío", nil)
			return
		}
		s.NOMBRE = nombre
		cols = append(cols, "NOMBRE")
	}
	if in.CategoriaId != nil {
		if *in.CategoriaId <= 0 {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo 'categoriaId' debe ser un entero positivo", nil)
			return
		}
		s.PK_ID_CATEGORIA = &models.Categoria{PK_ID_CATEGORIA: *in.CategoriaId}
		cols = append(cols, "PK_ID_CATEGORIA")
	}
	if len(cols) == 0 {
		httpx.Send(&c.Controller, http.StatusOK, "Subcategoría actualizada", s)
		return
	}
	if _, err := orm.NewOrm().Update(&s, cols...); err != nil {
		c.writeError("put.update_error", "Error al actualizar subcategoría", err)
		return
	}
	actualizada, err := porID(s.PK_ID_SUBCATEGORIA)
	if err != nil {
		c.readError("put", s.PK_ID_SUBCATEGORIA, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Subcategoría actualizada", actualizada)
}

// @Title Delete
// @Summary Eliminar subcategoría
// @Description Elimina físicamente la subcategoría. Si tiene productos asociados responde 409.
// @Tags subcategorias
// @Accept json
// @Produce json
// @Param id query int true "ID de la subcategoría (entero positivo)"
// @Success 200 {object} models.ApiResponse "Subcategoría eliminada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Subcategoría no encontrada"
// @Failure 409 {object} models.ApiResponse "La subcategoría tiene productos asociados"
// @Failure 500 {object} models.ApiResponse "Error al eliminar la subcategoría"
// @Security BearerAuth
// @Router /subcategorias [delete]
func (c *SubcategoriaController) Delete() {
	s, ok := c.load("delete")
	if !ok {
		return
	}
	// Raw y no orm.Delete: el ORM borraría en cascada los productos por su
	// cuenta; así la llave foránea de la BD protege los datos (409).
	if _, err := orm.NewOrm().Raw("DELETE FROM subcategoria WHERE pk_id_subcategoria = ?", s.PK_ID_SUBCATEGORIA).Exec(); err != nil {
		c.writeError("delete.delete_error", "Error al eliminar subcategoría", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Subcategoría eliminada", nil)
}
