package routers

import (
	"restaurante/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	// Ruta para obtener todos los clientes
	beego.Router("/clientes", &controllers.ClienteController{}, "get:GetAll;post:Post")

	// Ruta para obtener, actualizar y eliminar un cliente por ID
	beego.Router("/clientes/:id", &controllers.ClienteController{}, "get:GetById;put:Put;delete:Delete")
}
