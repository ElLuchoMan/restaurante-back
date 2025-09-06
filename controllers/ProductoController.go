package controllers

import (
	"encoding/json"
	"fmt"
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

// @Title GetAll
// @Summary Obtener productos con filtros
// @Description Devuelve productos registrados con filtros opcionales para imágenes y disponibilidad.
// @Tags productos
// @Accept json
// @Produce json
// @Param   includeImage  query    bool   false  "Incluir imágenes Base64 en la respuesta (true o false, por defecto es false)"
// @Param   onlyActive    query    bool   false  "Filtrar solo productos disponibles (true o false, por defecto es false)"
// @Success 200 {array} models.Producto "Lista de productos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos [get]
func (c *ProductoController) GetAll() {
	o := orm.NewOrm()
	var productos []models.Producto

	// Obtener valores de los parámetros
	includeImage, _ := c.GetBool("includeImage", false)
	onlyActive, _ := c.GetBool("onlyActive", false)

	// Construir la consulta con filtros
	query := o.QueryTable(new(models.Producto))
	if onlyActive {
		query = query.Filter("ESTADO_PRODUCTO", models.EstadoProductoDisponible)
	}

	// Ejecutar la consulta
	_, err := query.All(&productos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener producto por ID
// @Description Devuelve un producto específico por ID, incluyendo la imagen en formato Base64.
// @Tags productos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Producto"
// @Success 200 {object} models.Producto "Producto encontrado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Router /productos/search [get]
func (c *ProductoController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	producto, err := getProductoByID(int64(id), o)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto encontrado",
		Data:    producto,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear un nuevo producto
// @Description Crea un nuevo producto en la base de datos.
// @Tags productos
// @Accept json
// @Produce json
// @Param   producto  body     models.Producto true "Producto con imagen Base64"
// @Success 201 {object} models.Producto "Producto creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /productos [post]
func (c *ProductoController) Post() {
	o := orm.NewOrm()
	var producto models.Producto

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	producto.ESTADO_PRODUCTO = models.EstadoProducto(strings.ToUpper(string(producto.ESTADO_PRODUCTO)))

	if err := validateProducto(&producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
		c.ServeJSON()
		return
	}

	if _, err := o.Insert(&producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el producto", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	hist := models.PrecioProductoHist{PKIDProducto: &producto, Precio: producto.PRECIO, FechaVigencia: time.Now()}
	if _, err := o.Insert(&hist); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al registrar historial de precios", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Producto creado correctamente", Data: producto}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un producto
// @Description Actualiza los datos de un producto existente, incluyendo una imagen en formato Base64.
// @Tags productos
// @Accept json
// @Produce json
// @Param   id        query   int              true  "ID del Producto"
// @Param   producto  body    models.Producto true  "Datos del producto"
// @Success 200 {object} models.Producto "Producto actualizado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Router /productos [put]
func (c *ProductoController) Put() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente.", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	producto := models.Producto{PK_ID_PRODUCTO: int64(id)}

	if o.Read(&producto) == nil {
		original := producto

		var input models.Producto
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "JSON inválido", Cause: err.Error()}
			c.ServeJSON()
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

		if err := validateProducto(&producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
			c.ServeJSON()
			return
		}

		if reflect.DeepEqual(producto, original) {
			c.Ctx.Output.SetStatus(http.StatusNotModified)
			c.Data["json"] = models.ApiResponse{Code: http.StatusNotModified, Message: "No se realizaron cambios en el producto"}
			c.ServeJSON()
			return
		}

		if _, err = o.Update(&producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el producto.", Cause: err.Error()}
			c.ServeJSON()
			return
		}

		if producto.PRECIO != original.PRECIO {
			hist := models.PrecioProductoHist{PKIDProducto: &producto, Precio: producto.PRECIO, FechaVigencia: time.Now()}
			if _, err := o.Insert(&hist); err != nil {
				c.Ctx.Output.SetStatus(http.StatusInternalServerError)
				c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al registrar historial de precios", Cause: err.Error()}
				c.ServeJSON()
				return
			}
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Producto actualizado", Data: producto}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Producto no encontrado."}
		c.ServeJSON()
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
	o := orm.NewOrm()

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
		c.ServeJSON()
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
		c.ServeJSON()
		return
	}
	if producto.ESTADO_PRODUCTO == models.EstadoProductoNoDisponible {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El producto ya está desactivado.",
		}
		c.ServeJSON()
		return
	}
	// Cambiar el estado del producto a "NO_DISPONIBLE" para el borrado lógico
	producto.ESTADO_PRODUCTO = models.EstadoProductoNoDisponible
	if _, err := o.Update(producto, "ESTADO_PRODUCTO"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desactivar el producto.",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto desactivado correctamente.",
	}
	c.ServeJSON()
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
	if err := o.Read(producto); err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("producto no encontrado")
		}
		return nil, fmt.Errorf("error al buscar el producto: %v", err)
	}
	return producto, nil
}
