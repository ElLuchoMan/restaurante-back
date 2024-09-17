package routers

import (
	"restaurante/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/restaurante/v1",
		// Rutas para clientes
		beego.NSNamespace("/clientes",
			beego.NSRouter("/", &controllers.ClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ClienteController{}, "get:GetById"),
		),
		// Rutas para restaurantes
		beego.NSNamespace("/restaurantes",
			beego.NSRouter("/", &controllers.RestauranteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.RestauranteController{}, "get:GetById"),
		),
		// Rutas para pedidos
		beego.NSNamespace("/pedidos",
			beego.NSRouter("/", &controllers.PedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PedidoController{}, "get:GetById"),
		),
		// Rutas para domicilios
		beego.NSNamespace("/domicilios",
			beego.NSRouter("/", &controllers.DomicilioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.DomicilioController{}, "get:GetById"),
		),
		// Rutas para trabajadores
		beego.NSNamespace("/trabajadores",
			beego.NSRouter("/", &controllers.TrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.TrabajadorController{}, "get:GetById"),
		),
	)

	beego.AddNamespace(ns)
}
