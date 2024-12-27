package controllers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ProductoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener productos
// @Description Devuelve todos los productos registrados en la base de datos. Puedes incluir o excluir las imágenes con el parámetro `includeImage` y filtrar los productos activos con `onlyActive`.
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
		query = query.Filter("ESTADO_PRODUCTO", "DISPONIBLE")
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
			productos[i].IMAGEN = "" // Excluir imágenes
		} else if productos[i].IMAGEN != "" {
			productos[i].IMAGEN = "data:image/jpeg;base64," + productos[i].IMAGEN
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
		c.Ctx.Output.SetStatus(http.StatusNotFound)
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

// @Title Create
// @Summary Crear un nuevo producto
// @Description Crea un nuevo producto en la base de datos, incluyendo una imagen en formato Base64.
// @Tags productos
// @Accept multipart/form-data
// @Produce json
// @Param   NOMBRE        formData  string  true   "Nombre del producto"
// @Param   CALORIAS      formData  int     false   "Calorías del producto"
// @Param   DESCRIPCION   formData  string  false  "Descripción del producto"
// @Param   ESTADO_PRODUCTO formData  string    true   "Estado del producto"
// @Param   PRECIO        formData  int     true   "Precio del producto"
// @Param   IMAGEN        formData  file    false  "Imagen del producto (opcional)"
// @Param   CANTIDAD        formData  int     false   "Cantidad del producto"
// @Success 201 {object} models.Producto "Producto creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /productos [post]
func (c *ProductoController) Post() {
	o := orm.NewOrm()
	var producto models.Producto

	// Validar campos obligatorios
	producto.NOMBRE = c.GetString("NOMBRE")
	producto.ESTADO_PRODUCTO = c.GetString("ESTADO_PRODUCTO")
	calorias, _ := c.GetInt64("CALORIAS")
	producto.CALORIAS = &calorias
	producto.DESCRIPCION = c.GetString("DESCRIPCION")
	producto.PRECIO, _ = c.GetInt64("PRECIO")
	producto.CANTIDAD, _ = c.GetInt("CANTIDAD")

	if err := validateProducto(&producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Manejar la imagen opcional
	imagen, err := handleImageUpload(&c.Controller)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}
	producto.IMAGEN = imagen

	// Insertar en la base de datos
	_, err = o.Insert(&producto)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el producto",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Producto creado correctamente",
		Data:    producto,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un producto
// @Description Actualiza los datos de un producto existente, incluyendo una imagen en formato Base64.
// @Tags productos
// @Accept multipart/form-data
// @Produce json
// @Param   id            query    int     true   "ID del Producto"
// @Param   NOMBRE        formData  string  true   "Nombre del producto"
// @Param   CALORIAS      formData  int     false   "Calorías del producto"
// @Param   DESCRIPCION   formData  string  false  "Descripción del producto"
// @Param   ESTADO_PRODUCTO formData  string    true   "Estado del producto"
// @Param   PRECIO        formData  int     true   "Precio del producto"
// @Param   IMAGEN        formData  file    false  "Imagen del producto (opcional)"
// @Param   CANTIDAD        formData  int     false   "Cantidad del producto"
// @Success 200 {object} models.Producto "Producto actualizado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
// @Router /productos [put]
func (c *ProductoController) Put() {
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

	producto := models.Producto{PK_ID_PRODUCTO: int64(id)}

	if o.Read(&producto) == nil {
		// Copiar los valores actuales para comparación
		original := producto

		// Actualizar los campos
		producto.NOMBRE = c.GetString("NOMBRE")
		calorias, _ := c.GetInt64("CALORIAS")
		producto.CALORIAS = &calorias
		producto.DESCRIPCION = c.GetString("DESCRIPCION")
		producto.PRECIO, _ = c.GetInt64("PRECIO")
		producto.ESTADO_PRODUCTO = c.GetString("ESTADO_PRODUCTO")
		producto.CANTIDAD, _ = c.GetInt("CANTIDAD")

		// Validar datos
		if err := validateProducto(&producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
			c.ServeJSON()
			return
		}

		// Manejar imagen opcional
		imagen, err := handleImageUpload(&c.Controller)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
			c.ServeJSON()
			return
		}
		if imagen != "" {
			producto.IMAGEN = imagen
		}

		// Verificar si hubo cambios
		if producto == original {
			c.Ctx.Output.SetStatus(http.StatusNotModified)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotModified,
				Message: "No se realizaron cambios en el producto",
			}
			c.ServeJSON()
			return
		}

		// Actualizar en base de datos
		if _, err = o.Update(&producto); err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el producto.",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Producto actualizado",
			Data:    producto,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Producto no encontrado.",
		}
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
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}
	if producto.ESTADO_PRODUCTO == "NO DISPONIBLE" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El producto ya está desactivado.",
		}
		c.ServeJSON()
		return
	}
	// Cambiar el estado del producto a "NO DISPONIBLE" para el borrado lógico
	producto.ESTADO_PRODUCTO = "NO DISPONIBLE"
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

// Funciones auxiliares
func handleImageUpload(c *web.Controller) (string, error) {
	file, fileHeader, err := c.GetFile("IMAGEN")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", fmt.Errorf("error al obtener la imagen: %v", err)
	}
	defer file.Close()

	if fileHeader.Size > 1024*1024 {
		return "", fmt.Errorf("la imagen no debe superar los 1MB")
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error al leer la imagen: %v", err)
	}
	return base64.StdEncoding.EncodeToString(fileBytes), nil
}

func validateProducto(producto *models.Producto) error {
	if producto.NOMBRE == "" {
		return fmt.Errorf("el campo 'NOMBRE' es obligatorio")
	}
	if producto.PRECIO <= 0 {
		return fmt.Errorf("el campo 'PRECIO' debe ser un número mayor a 0")
	}
	if producto.CALORIAS != nil && *producto.CALORIAS < 0 {
		return fmt.Errorf("el campo 'CALORIAS' debe ser un número positivo")
	}
	if producto.ESTADO_PRODUCTO != "DISPONIBLE" && producto.ESTADO_PRODUCTO != "NO DISPONIBLE" {
		return fmt.Errorf("el campo 'ESTADO_PRODUCTO' debe ser 'DISPONIBLE' o 'NO DISPONIBLE'")
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
