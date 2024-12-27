package controllers

import (
	"encoding/base64"
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
// @Summary Obtener todos los productos
// @Description Devuelve todos los productos registrados en la base de datos sin la imagen (IMAGEN).
// @Tags productos
// @Accept json
// @Produce json
// @Success 200 {array} models.Producto "Lista de todos los productos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos [get]
func (c *ProductoController) GetAll() {
	o := orm.NewOrm()
	var productos []models.Producto

	_, err := o.QueryTable(new(models.Producto)).All(&productos)
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

	// Excluir la foto (IMAGEN) de cada producto
	for i := range productos {
		productos[i].IMAGEN = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos obtenidos exitosamente",
		Data:    productos,
	}
	c.ServeJSON()
}

// @Title GetAllActive
// @Summary Obtener todos los productos activos
// @Description Devuelve solo los productos que están activos (ACTIVO = TRUE).
// @Tags productos
// @Accept json
// @Produce json
// @Success 200 {array} models.Producto "Lista de productos activos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /productos/active [get]
func (c *ProductoController) GetAllActive() {
	o := orm.NewOrm()
	var productos []models.Producto

	_, err := o.QueryTable(new(models.Producto)).Filter("ACTIVO", true).All(&productos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener productos activos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la foto (IMAGEN) de cada producto
	for i := range productos {
		productos[i].IMAGEN = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos activos obtenidos exitosamente",
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

	producto := models.Producto{PK_ID_PRODUCTO: int64(id)}

	err = o.Read(&producto)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Producto no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Si todo está bien, devuelve el producto, incluyendo la imagen en Base64
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto encontrado",
		Data:    producto, // Incluye la imagen en Base64
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
// @Param   CALORIAS      formData  int     true   "Calorías del producto"
// @Param   DESCRIPCION   formData  string  false  "Descripción del producto"
// @Param   PRECIO        formData  int     true   "Precio del producto"
// @Param   PERSONALIZADO formData  bool    true   "Indica si el producto es personalizado"
// @Param   IMAGEN          formData  file    false  "Imagen del producto (opcional)"
// @Success 201 {object} models.Producto "Producto creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /productos [post]
func (c *ProductoController) Post() {
	o := orm.NewOrm()
	var producto models.Producto

	// Validar campos requeridos
	producto.NOMBRE = c.GetString("NOMBRE")
	if producto.NOMBRE == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'NOMBRE' es requerido.",
		}
		c.ServeJSON()
		return
	}

	calorias, _ := c.GetInt64("CALORIAS")
	producto.CALORIAS = &calorias
	producto.DESCRIPCION = c.GetString("DESCRIPCION")
	producto.PRECIO, _ = c.GetInt64("PRECIO")

	// Manejar la imagen opcional
	file, fileHeader, err := c.GetFile("IMAGEN")
	if err == nil {
		defer file.Close()

		// Validar el tamaño del archivo (máximo 1MB)
		if fileHeader.Size > 1024*1024 { // 1MB en bytes
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "La imagen no debe superar los 1MB.",
			}
			c.ServeJSON()
			return
		}

		// Leer y convertir a Base64
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error al leer la imagen.",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		producto.IMAGEN = base64.StdEncoding.EncodeToString(fileBytes)
	}

	// Insertar el nuevo producto en la base de datos
	_, err = o.Insert(&producto)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el producto.",
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
// @Param   CALORIAS      formData  int     true   "Calorías del producto"
// @Param   DESCRIPCION   formData  string  false  "Descripción del producto"
// @Param   PRECIO        formData  int     true   "Precio del producto"
// @Param   PERSONALIZADO formData  bool    true   "Indica si el producto es personalizado"
// @Param   IMAGEN          formData  file    false  "Imagen del producto (opcional)"
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
		producto.NOMBRE = c.GetString("NOMBRE")
		calorias, _ := c.GetInt64("CALORIAS")
		producto.CALORIAS = &calorias
		producto.DESCRIPCION = c.GetString("DESCRIPCION")
		producto.PRECIO, _ = c.GetInt64("PRECIO")

		// Manejar la imagen opcional
		file, fileHeader, err := c.GetFile("IMAGEN")
		if err == nil {
			defer file.Close()

			// Validar el tamaño del archivo (máximo 1MB)
			if fileHeader.Size > 1024*1024 { // 1MB en bytes
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "La imagen no debe superar los 1MB.",
				}
				c.ServeJSON()
				return
			}

			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Error al leer la imagen.",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			producto.IMAGEN = base64.StdEncoding.EncodeToString(fileBytes)
		}

		// Actualizar el producto en la base de datos
		_, err = o.Update(&producto)
		if err != nil {
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
// @Success 204 {object} nil "Producto desactivado"
// @Failure 404 {object} models.ApiResponse "Producto no encontrado"
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
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		c.ServeJSON()
		return
	}

	// Buscar el producto
	producto := models.Producto{PK_ID_PRODUCTO: int64(id)}
	if err := o.Read(&producto); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Producto no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Realizar el borrado lógico
	producto.ESTADO_PRODUCTO = "DISPONIBLE"
	if _, err := o.Update(&producto); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Producto no disponible",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desactivar el producto",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
