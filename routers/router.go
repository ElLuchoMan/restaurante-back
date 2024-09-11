package routers

import (
	"restaurante/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/restaurante/v1",
		beego.NSNamespace("/clientes",
			beego.NSRouter("/", &controllers.ClienteController{}, "get:GetAll;post:Post"),
			beego.NSRouter("/:id", &controllers.ClienteController{}, "get:GetById;put:Put;delete:Delete"),
		),
	)

	beego.AddNamespace(ns)
}
