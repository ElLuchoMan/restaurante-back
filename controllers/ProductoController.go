package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"restaurante/models"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ProductoController struct {
	web.Controller
}

// puntos de inyección para tests
var (
	ormNewProducto    = orm.NewOrm
	queryProductosAll = func(o orm.Ormer, onlyActive bool, productos *[]models.Producto) (int64, error) {
		qs := o.QueryTable(new(models.Producto))
		if onlyActive {
			qs = qs.Filter("ESTADO_PRODUCTO", models.EstadoProductoDisponible)
		}
		return qs.All(productos)
	}
	readProductoFn     = func(o orm.Ormer, p *models.Producto) error { return o.Read(p) }
	insertProductoFn   = func(o orm.Ormer, p *models.Producto) (int64, error) { return o.Insert(p) }
	insertPrecioHistFn = func(o orm.Ormer, h *models.PrecioProductoHist) (int64, error) { return o.Insert(h) }
	updateProductoFn   = func(o orm.Ormer, p *models.Producto, cols ...string) (int64, error) { return o.Update(p, cols...) }
)

// @Title GetAll
// @Summary Obtener productos con filtros
// @Description Devuelve productos registrados con filtros opcionales para imágenes y disponibilidad.
// @Tags productos
// @Accept json
// @Produce json
// @Param   includeImage  query    bool   false  "Incluir imágenes Base64 en la respuesta (true o false, por defecto es false)"
// @Param   onlyActive    query    bool   false  "Filtrar solo productos disponibles (true o false, por defecto es false)"
// @Success 200 {object} models.ApiResponse{data=[]models.Producto} "Lista de productos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos [get]
func (c *ProductoController) GetAll() {
	o := ormNewProducto()
	var productos []models.Producto

	// Obtener valores de los parámetros
	includeImage, _ := c.GetBool("includeImage", false)
	onlyActive, _ := c.GetBool("onlyActive", false)

	// Ejecutar la consulta con filtros
	_, err := queryProductosAll(o, onlyActive, &productos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Manejar imágenes según el parámetro includeImage
	for i := range productos {
		if !includeImage {
			productos[i].IMAGEN = ""
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos obtenidos exitosamente",
		Data:    productos,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener producto por ID
// @Description Devuelve un producto específico por ID, incluyendo la imagen en formato Base64.
// @Tags productos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Producto"
// @Success 200 {object} models.ApiResponse{data=models.Producto} "Producto encontrado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Router /productos/search [get]
func (c *ProductoController) GetById() {
	o := ormNewProducto()
	id, err := c.GetInt("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	producto, err := getProductoByID(int64(id), o)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto encontrado",
		Data:    producto,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear un nuevo producto
// @Description Crea un nuevo producto. Puedes enviar JSON (imagen en Base64) o multipart/form-data con archivo.
// @Tags productos
// @Accept json
// @Accept mpfd
// @Produce json
// @Param   nombre         formData string  false "Nombre del producto"
// @Param   calorias       formData integer false "Calorías"
// @Param   descripcion    formData string  false "Descripción"
// @Param   precio         formData integer false "Precio"
// @Param   estadoProducto formData string  false "DISPONIBLE | NO_DISPONIBLE"
// @Param   cantidad       formData integer false "Cantidad"
// @Param   subcategoriaId formData integer false "ID de subcategoría"
// @Param   imagen         formData file    false "Archivo de imagen"
// @Success 201 {object} models.ApiResponse{data=models.Producto} "Producto creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /productos [post]
func (c *ProductoController) Post() {
	o := ormNewProducto()
	var producto models.Producto

	contentType := c.Ctx.Input.Header("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		// Leer campos desde form-data
		nombre := c.GetString("nombre")
		descripcion := c.GetString("descripcion")
		precioStr := c.GetString("precio")
		estado := strings.ToUpper(c.GetString("estadoProducto"))
		cantidadStr := c.GetString("cantidad")
		caloriasStr := c.GetString("calorias")
		subcatStr := c.GetString("subcategoriaId")

		producto.NOMBRE = nombre
		if descripcion != "" {
			producto.DESCRIPCION = &descripcion
		}
		if v, err := strconv.ParseInt(precioStr, 10, 64); err == nil {
			producto.PRECIO = v
		}
		producto.ESTADO_PRODUCTO = models.EstadoProducto(estado)
		if v, err := strconv.Atoi(cantidadStr); err == nil {
			producto.CANTIDAD = v
		}
		if v, err := strconv.ParseInt(caloriasStr, 10, 64); err == nil {
			producto.CALORIAS = &v
		}
		if v, err := strconv.ParseInt(subcatStr, 10, 64); err == nil {
			producto.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: v}
		}

		// Archivo de imagen (opcional)
		file, _, err := c.GetFile("imagen")
		if err == nil && file != nil {
			defer func() { _ = file.Close() }()
			if data, rerr := io.ReadAll(file); rerr == nil {
				producto.IMAGEN = string(data)
			}
		}
	} else {
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	}

	producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(string(producto.ESTADO_PRODUCTO)))

	if err := validateProducto(&producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if _, err := insertProductoFn(o, &producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el producto", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Alinear la secuencia del historial de precios con el máximo actual (previene duplicados de PK por desalineación de secuencia)
	_, _ = o.Raw("SELECT setval(pg_get_serial_sequence('precio_producto_hist','pk_id_precio_hist'), COALESCE((SELECT MAX(pk_id_precio_hist) FROM precio_producto_hist),0))").Exec()

	hist := models.PrecioProductoHist{PKIDProducto: &producto, Precio: producto.PRECIO, FechaVigencia: time.Now()}
	if _, err := insertPrecioHistFn(o, &hist); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al registrar historial de precios", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Producto creado correctamente", Data: producto}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un producto
// @Description Actualiza un producto. Puedes enviar JSON (imagen en Base64) o multipart/form-data con archivo.
// @Tags productos
// @Accept json
// @Accept mpfd
// @Produce json
// @Param   id             query   int     true  "ID del Producto"
// @Param   nombre         formData string  false "Nombre del producto"
// @Param   calorias       formData integer false "Calorías"
// @Param   descripcion    formData string  false "Descripción"
// @Param   precio         formData integer false "Precio"
// @Param   estadoProducto formData string  false "DISPONIBLE | NO_DISPONIBLE"
// @Param   cantidad       formData integer false "Cantidad"
// @Param   subcategoriaId formData integer false "ID de subcategoría"
// @Param   imagen         formData file    false "Archivo de imagen"
// @Success 200 {object} models.ApiResponse{data=models.Producto} "Producto actualizado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Router /productos [put]
func (c *ProductoController) Put() {
	o := ormNewProducto()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente.", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	producto := models.Producto{PK_ID_PRODUCTO: int64(id)}

	if readProductoFn(o, &producto) == nil {
		original := producto

		contentType := c.Ctx.Input.Header("Content-Type")
		if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
			// Sólo actualizar campos presentes
			if v := c.GetString("nombre"); v != "" {
				producto.NOMBRE = v
			}
			if v := c.GetString("descripcion"); v != "" {
				producto.DESCRIPCION = &v
			}
			if v := c.GetString("precio"); v != "" {
				if n, e := strconv.ParseInt(v, 10, 64); e == nil {
					producto.PRECIO = n
				}
			}
			if v := c.GetString("estadoProducto"); v != "" {
				producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(v))
			}
			if v := c.GetString("cantidad"); v != "" {
				if n, e := strconv.Atoi(v); e == nil {
					producto.CANTIDAD = n
				}
			}
			if v := c.GetString("calorias"); v != "" {
				if n, e := strconv.ParseInt(v, 10, 64); e == nil {
					producto.CALORIAS = &n
				}
			}
			if v := c.GetString("subcategoriaId"); v != "" {
				if n, e := strconv.ParseInt(v, 10, 64); e == nil {
					producto.PK_ID_SUBCATEGORIA = &models.Subcategoria{PK_ID_SUBCATEGORIA: n}
				}
			}

			// Imagen por archivo
			if file, _, ferr := c.GetFile("imagen"); ferr == nil && file != nil {
				defer func() { _ = file.Close() }()
				if data, rerr := io.ReadAll(file); rerr == nil {
					producto.IMAGEN = string(data)
				}
			}
		} else {
			var input models.Producto
			if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
				_ = c.ServeJSON()
				return
			}

			producto.NOMBRE = input.NOMBRE
			producto.CALORIAS = input.CALORIAS
			producto.DESCRIPCION = input.DESCRIPCION
			producto.PRECIO = input.PRECIO
			producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(string(input.ESTADO_PRODUCTO)))
			producto.CANTIDAD = input.CANTIDAD
			producto.PK_ID_SUBCATEGORIA = input.PK_ID_SUBCATEGORIA
			if len(input.IMAGEN) > 0 {
				producto.IMAGEN = input.IMAGEN
			}
		}

		if err := validateProducto(&producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
			_ = c.ServeJSON()
			return
		}

		if reflect.DeepEqual(producto, original) {
			c.Ctx.Output.SetStatus(http.StatusNotModified)
			c.Data["json"] = models.ApiResponse{Code: http.StatusNotModified, Message: "No se realizaron cambios en el producto"}
			_ = c.ServeJSON()
			return
		}

		if _, err = updateProductoFn(o, &producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el producto.", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}

		if producto.PRECIO != original.PRECIO {
			// Alinear la secuencia antes de insertar un nuevo historial
			_, _ = o.Raw("SELECT setval(pg_get_serial_sequence('precio_producto_hist','pk_id_precio_hist'), COALESCE((SELECT MAX(pk_id_precio_hist) FROM precio_producto_hist),0))").Exec()

			hist := models.PrecioProductoHist{PKIDProducto: &producto, Precio: producto.PRECIO, FechaVigencia: time.Now()}
			if _, err := insertPrecioHistFn(o, &hist); err != nil {
				c.Ctx.Output.SetStatus(http.StatusInternalServerError)
				c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al registrar historial de precios", Cause: err.Error()}
				_ = c.ServeJSON()
				return
			}
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Producto actualizado", Data: producto}
		_ = c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Producto no encontrado."}
		_ = c.ServeJSON()
	}

}

// @Title Delete
// @Summary Desactivar un producto
// @Description Desactiva un producto en la base de datos (borrado lógico).
// @Tags productos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Producto"
// @Success 200 {object} models.ApiResponse "Producto desactivado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos [delete]
func (c *ProductoController) Delete() {
	o := ormNewProducto()

	// Obtener el ID del query parameter
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente.",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Buscar el producto
	producto, err := getProductoByID(int64(id), o)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	if producto.ESTADO_PRODUCTO == models.EstadoProductoNoDisponible {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El producto ya está desactivado.",
		}
		_ = c.ServeJSON()
		return
	}
	// Cambiar el estado del producto a "NO_DISPONIBLE" para el borrado lógico
	producto.ESTADO_PRODUCTO = models.EstadoProductoNoDisponible
	if _, err := updateProductoFn(o, producto, "ESTADO_PRODUCTO"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desactivar el producto.",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto desactivado correctamente.",
	}
	_ = c.ServeJSON()
}

func validateProducto(producto *models.Producto) error {
	if producto.NOMBRE == "" {
		return fmt.Errorf("el campo 'nombre' es obligatorio")
	}
	if producto.PRECIO <= 0 {
		return fmt.Errorf("el campo 'precio' debe ser un número mayor a 0")
	}
	if producto.CALORIAS != nil && *producto.CALORIAS < 0 {
		return fmt.Errorf("el campo 'calorias' debe ser un número positivo")
	}
	if producto.ESTADO_PRODUCTO != models.EstadoProductoDisponible && producto.ESTADO_PRODUCTO != models.EstadoProductoNoDisponible {
		return fmt.Errorf("el campo 'estadoProducto' debe ser 'DISPONIBLE' o 'NO_DISPONIBLE'")
	}
	return nil
}

func getProductoByID(id int64, o orm.Ormer) (*models.Producto, error) {
	producto := &models.Producto{PK_ID_PRODUCTO: id}
	if err := readProductoFn(o, producto); err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("producto no encontrado")
		}
		return nil, fmt.Errorf("error al buscar el producto: %v", err)
	}
	return producto, nil
}
