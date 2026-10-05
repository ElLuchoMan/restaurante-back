package producto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"restaurante/database"
	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// ProductoController gestiona /productos.
type ProductoController struct {
	web.Controller
}

const (
	msgIDInvalido  = "El parámetro 'id' es inválido o está ausente"
	msgBadProducto = "Error al procesar los datos del producto. Si envía una imagen, debe ser Base64 válido o use multipart/form-data."
)

// readAll lee el archivo de imagen de un formulario multipart.
var readAll = io.ReadAll

// nullableUpdate son los campos de PUT que admiten null explícito (se limpian).
var nullableUpdate = []string{"calorias", "descripcion", "imagen", "subcategoriaId"}

// @Title GetAll
// @Summary Obtener productos con filtros
// @Description Devuelve los productos con filtros opcionales para imágenes y disponibilidad. Sin resultados, `data` es `[]`. `subcategoriaId` es el id (número) o null si el producto no tiene subcategoría.
// @Tags productos
// @Accept json
// @Produce json
// @Param   includeImage  query    bool   false  "Incluir imágenes Base64 en la respuesta (true/false, por defecto false)"
// @Param   onlyActive    query    bool   false  "Solo productos DISPONIBLE (true/false, por defecto false)"
// @Success 200 {object} models.ApiResponse{data=[]models.ProductoDoc} "Lista de productos (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "includeImage u onlyActive no son booleanos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos [get]
func (c *ProductoController) GetAll() {
	includeImage, err := c.GetBool("includeImage", false)
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'includeImage' debe ser true o false", err)
		return
	}
	onlyActive, err := c.GetBool("onlyActive", false)
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'onlyActive' debe ser true o false", err)
		return
	}

	qs := orm.NewOrm().QueryTable(new(models.Producto))
	if onlyActive {
		qs = qs.Filter("ESTADO_PRODUCTO", models.EstadoProductoDisponible)
	}
	var productos []models.Producto
	if _, err := qs.All(&productos); err != nil {
		logging.LogControllerError(c.Ctx, "productos.getall.db_error", err, map[string]interface{}{
			"onlyActive":   onlyActive,
			"includeImage": includeImage,
		})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener productos de la base de datos", err)
		return
	}
	if !includeImage {
		for i := range productos {
			productos[i].IMAGEN = ""
		}
	}
	httpx.Send(&c.Controller, http.StatusOK, "Productos obtenidos exitosamente", httpx.List(productos))
}

// load valida el query param `id` y lee el producto (400 / 404 / 500).
func (c *ProductoController) load(op string) (models.Producto, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "productos."+op+".bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return models.Producto{}, false
	}
	producto := models.Producto{PK_ID_PRODUCTO: id}
	if err := orm.NewOrm().Read(&producto); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Producto no encontrado", nil)
			return producto, false
		}
		logging.LogControllerError(c.Ctx, "productos."+op+".read_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al buscar el producto", err)
		return producto, false
	}
	return producto, true
}

// writeError traduce un fallo de escritura: unicidad -> 409, FK -> 400
// (subcategoría inexistente) y cualquier otro -> 500.
func (c *ProductoController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "productos."+op, err, nil)
	switch {
	case dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe un producto con esos datos", err)
	case dberr.IsForeignKey(err):
		httpx.Fail(&c.Controller, http.StatusBadRequest, "La subcategoría indicada no existe", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
	}
}

// @Title GetById
// @Summary Obtener producto por ID
// @Description Devuelve un producto por ID, incluyendo la imagen en Base64.
// @Tags productos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del producto (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ProductoDoc} "Producto encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos/search [get]
func (c *ProductoController) GetById() {
	producto, ok := c.load("getbyid")
	if !ok {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Producto encontrado", producto)
}

func isMultipart(c *ProductoController) bool {
	return strings.HasPrefix(strings.ToLower(c.Ctx.Input.Header("Content-Type")), "multipart/form-data")
}

// formFile lee el archivo `imagen` del formulario, si viene.
func (c *ProductoController) formFile() (string, bool, error) {
	file, _, err := c.GetFile("imagen")
	if err != nil || file == nil {
		return "", false, nil
	}
	defer func() { _ = file.Close() }()
	data, err := readAll(file)
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

// formInt interpreta el campo de formulario name; vacío -> (0, false, nil).
func (c *ProductoController) formInt(name string) (int64, bool, error) {
	raw := strings.TrimSpace(c.GetString(name))
	if raw == "" {
		return 0, false, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("el campo '%s' debe ser un entero", name)
	}
	return n, true, nil
}

// applyForm aplica los campos presentes del formulario multipart sobre p
// (merge: los campos vacíos o ausentes se conservan).
func (c *ProductoController) applyForm(p *models.Producto) error {
	if v := c.GetString("nombre"); v != "" {
		p.NOMBRE = v
	}
	if v := c.GetString("descripcion"); v != "" {
		p.DESCRIPCION = &v
	}
	if v := c.GetString("estadoProducto"); v != "" {
		p.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(v))
	}
	for _, f := range []struct {
		name string
		set  func(int64)
	}{
		{"precio", func(n int64) { p.PRECIO = n }},
		{"cantidad", func(n int64) { p.CANTIDAD = int(n) }},
		{"calorias", func(n int64) { p.CALORIAS = &n }},
		{"subcategoriaId", func(n int64) { p.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: n} }},
	} {
		n, ok, err := c.formInt(f.name)
		if err != nil {
			return err
		}
		if ok {
			f.set(n)
		}
	}
	img, ok, err := c.formFile()
	if err != nil {
		return fmt.Errorf("no se pudo leer la imagen: %w", err)
	}
	if ok {
		p.IMAGEN = img
	}
	return nil
}

// @Title Post
// @Summary Crear un nuevo producto
// @Description Crea un producto y registra su precio inicial en el historial (ambas escrituras en una transacción). Acepta JSON (cuerpo `models.ProductoCreateRequest`, imagen en Base64, tolera prefijo `data:image/...;base64,`) o `multipart/form-data` con los mismos campos como texto y `imagen` como archivo. `nombre` obligatorio, `precio` > 0, `cantidad` >= 0, `estadoProducto` DISPONIBLE o NO_DISPONIBLE; `calorias`, `descripcion`, `imagen` y `subcategoriaId` son opcionales (sin subcategoría queda null). Respuesta 201 con el producto creado.
// @Tags productos
// @Accept json
// @Accept mpfd
// @Produce json
// @Param   body  body  models.ProductoCreateRequest  true  "Datos del producto (JSON). En multipart, los mismos campos como form-data"
// @Success 201 {object} models.ApiResponse{data=models.ProductoDoc} "Producto creado"
// @Failure 400 {object} models.ApiResponse "JSON/form inválido, imagen Base64 inválida, validación o subcategoría inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad"
// @Failure 500 {object} models.ApiResponse "Error al crear el producto"
// @Security BearerAuth
// @Router /productos [post]
func (c *ProductoController) Post() {
	var producto models.Producto
	if isMultipart(c) {
		if err := c.applyForm(&producto); err != nil {
			logging.LogControllerError(c.Ctx, "productos.post.bad_form", err, nil)
			httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadProducto, err)
			return
		}
	} else if err := json.Unmarshal(c.Ctx.Input.RequestBody, &producto); err != nil {
		logging.LogControllerError(c.Ctx, "productos.post.bad_json", err, map[string]interface{}{"contentType": c.Ctx.Input.Header("Content-Type")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadProducto, err)
		return
	}
	producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(string(producto.ESTADO_PRODUCTO)))

	if err := validateProducto(&producto); err != nil {
		logging.LogControllerError(c.Ctx, "productos.post.validation_error", err, map[string]interface{}{"nombre": producto.NOMBRE, "precio": producto.PRECIO})
		httpx.Fail(&c.Controller, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := orm.NewOrm().DoTx(func(_ context.Context, tx orm.TxOrmer) error {
		if _, err := tx.Insert(&producto); err != nil {
			return err
		}
		return registrarPrecio(tx, producto.PK_ID_PRODUCTO, producto.PRECIO)
	})
	if err != nil {
		c.writeError("post.insert_error", "Error al crear el producto", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Producto creado correctamente", producto)
}

// registrarPrecio guarda el precio vigente de hoy (hora de Bogotá) en el
// historial: actualiza el registro del día si ya existe o lo inserta.
func registrarPrecio(tx orm.TxOrmer, productoID, precio int64) error {
	ahora := time.Now()
	if database.BogotaZone != nil {
		ahora = ahora.In(database.BogotaZone)
	}
	hoy := time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 12, 0, 0, 0, time.UTC)
	res, err := tx.Raw("UPDATE precio_producto_hist SET precio = ? WHERE pk_id_producto = ? AND fecha_vigencia = ?", precio, productoID, hoy).Exec()
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return nil
	}
	_, err = tx.Insert(&models.PrecioProductoHist{
		PKIDProducto:  &models.Producto{PK_ID_PRODUCTO: productoID},
		Precio:        precio,
		FechaVigencia: hoy,
	})
	return err
}

// applyJSON aplica el cuerpo JSON (merge) sobre p. El error devuelto es de
// validación (HTTP 400).
func applyJSON(body []byte, p *models.Producto) error {
	var req models.ProductoUpdateRequest
	if err := httpx.DecodeMerge(body, &req, nullableUpdate...); err != nil {
		return err
	}
	if req.Nombre != nil {
		p.NOMBRE = *req.Nombre
	}
	if req.Precio != nil {
		p.PRECIO = *req.Precio
	}
	if req.EstadoProducto != nil {
		p.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(*req.EstadoProducto))
	}
	if req.Cantidad != nil {
		p.CANTIDAD = *req.Cantidad
	}
	if req.Calorias != nil {
		p.CALORIAS = req.Calorias
	} else if httpx.IsNull(body, "calorias") {
		p.CALORIAS = nil
	}
	if req.Descripcion != nil {
		p.DESCRIPCION = req.Descripcion
	} else if httpx.IsNull(body, "descripcion") {
		p.DESCRIPCION = nil
	}
	if req.Imagen != nil {
		img, err := models.DecodeImagenBase64(*req.Imagen)
		if err != nil {
			return fmt.Errorf("imagen Base64 inválida: %w", err)
		}
		p.IMAGEN = string(img)
	} else if httpx.IsNull(body, "imagen") {
		p.IMAGEN = ""
	}
	if req.SubcategoriaId != nil {
		if *req.SubcategoriaId <= 0 {
			return errors.New("el campo 'subcategoriaId' debe ser un entero positivo")
		}
		p.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: *req.SubcategoriaId}
	} else if httpx.IsNull(body, "subcategoriaId") {
		p.PK_ID_SUBCATEGORIA = nil
	}
	return nil
}

// @Title Update
// @Summary Actualizar un producto
// @Description Actualización parcial (merge): los campos ausentes se conservan (incluida `subcategoriaId`). `calorias`, `descripcion`, `imagen` y `subcategoriaId` admiten null explícito (se limpian); null en cualquier otro campo responde 400. Un cuerpo sin cambios responde 200 con el producto actual (no 304). Si cambia `precio` se registra en el historial (misma transacción). Acepta JSON (`models.ProductoUpdateRequest`, imagen Base64) o `multipart/form-data` con los mismos campos como texto y `imagen` como archivo (solo se aplican los campos no vacíos).
// @Tags productos
// @Accept json
// @Accept mpfd
// @Produce json
// @Param   id    query   int     true  "ID del producto (entero positivo)"
// @Param   body  body    models.ProductoUpdateRequest  true  "Campos a modificar (JSON). En multipart, los mismos campos como form-data"
// @Success 200 {object} models.ApiResponse{data=models.ProductoDoc} "Producto actualizado (o sin cambios)"
// @Failure 400 {object} models.ApiResponse "id inválido, JSON/form inválido, null en campo no anulable, validación o subcategoría inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad"
// @Failure 500 {object} models.ApiResponse "Error al actualizar el producto"
// @Security BearerAuth
// @Router /productos [put]
func (c *ProductoController) Put() {
	producto, ok := c.load("put")
	if !ok {
		return
	}
	original := producto

	var err error
	if isMultipart(c) {
		err = c.applyForm(&producto)
	} else {
		err = applyJSON(c.Ctx.Input.RequestBody, &producto)
	}
	if err != nil {
		logging.LogControllerError(c.Ctx, "productos.put.bad_request", err, map[string]interface{}{"id": original.PK_ID_PRODUCTO})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadProducto, err)
		return
	}
	producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(string(producto.ESTADO_PRODUCTO)))

	if err := validateProducto(&producto); err != nil {
		logging.LogControllerError(c.Ctx, "productos.put.validation_error", err, map[string]interface{}{"id": original.PK_ID_PRODUCTO})
		httpx.Fail(&c.Controller, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if reflect.DeepEqual(producto, original) {
		httpx.Send(&c.Controller, http.StatusOK, "Sin cambios en el producto", producto)
		return
	}

	err = orm.NewOrm().DoTx(func(_ context.Context, tx orm.TxOrmer) error {
		if _, err := tx.Update(&producto); err != nil {
			return err
		}
		if producto.PRECIO != original.PRECIO {
			return registrarPrecio(tx, producto.PK_ID_PRODUCTO, producto.PRECIO)
		}
		return nil
	})
	if err != nil {
		c.writeError("put.update_error", "Error al actualizar el producto", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Producto actualizado", producto)
}

// @Title Delete
// @Summary Desactivar un producto
// @Description Borrado lógico: pone `estadoProducto` en NO_DISPONIBLE (no elimina la fila). Si ya estaba desactivado responde 400.
// @Tags productos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del producto (entero positivo)"
// @Success 200 {object} models.ApiResponse "Producto desactivado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o producto ya desactivado"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /productos [delete]
func (c *ProductoController) Delete() {
	producto, ok := c.load("delete")
	if !ok {
		return
	}
	if producto.ESTADO_PRODUCTO == models.EstadoProductoNoDisponible {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El producto ya está desactivado", nil)
		return
	}
	producto.ESTADO_PRODUCTO = models.EstadoProductoNoDisponible
	if _, err := orm.NewOrm().Update(&producto, "ESTADO_PRODUCTO"); err != nil {
		logging.LogControllerError(c.Ctx, "productos.delete.update_error", err, map[string]interface{}{"id": producto.PK_ID_PRODUCTO})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al desactivar el producto", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Producto desactivado correctamente", nil)
}

func validateProducto(producto *models.Producto) error {
	if strings.TrimSpace(producto.NOMBRE) == "" {
		return errors.New("el campo 'nombre' es obligatorio")
	}
	if producto.PRECIO <= 0 {
		return errors.New("el campo 'precio' debe ser un número mayor a 0")
	}
	if producto.CALORIAS != nil && *producto.CALORIAS < 0 {
		return errors.New("el campo 'calorias' debe ser un número positivo")
	}
	if producto.CANTIDAD < 0 {
		return errors.New("el campo 'cantidad' no puede ser negativo")
	}
	if producto.ESTADO_PRODUCTO != models.EstadoProductoDisponible && producto.ESTADO_PRODUCTO != models.EstadoProductoNoDisponible {
		return errors.New("el campo 'estadoProducto' debe ser 'DISPONIBLE' o 'NO_DISPONIBLE'")
	}
	return nil
}
