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
		// Rutas para platos
		beego.NSNamespace("/platos",
			beego.NSRouter("/", &controllers.PlatoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PlatoController{}, "get:GetById"),
		),
		// Rutas para reservas
		beego.NSNamespace("/reservas",
			beego.NSRouter("/", &controllers.ReservaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ReservaController{}, "get:GetById"),
		),
		// Rutas para métodos de pago
		beego.NSNamespace("/metodos_pago",
			beego.NSRouter("/", &controllers.MetodoPagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.MetodoPagoController{}, "get:GetById"),
		),
		// Rutas para pagos
		beego.NSNamespace("/pagos",
			beego.NSRouter("/", &controllers.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PagoController{}, "get:GetById"),
		),
		// Rutas para ingredientes
		beego.NSNamespace("/ingredientes",
			beego.NSRouter("/", &controllers.IngredienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.IngredienteController{}, "get:GetById"),
		),
		// Rutas para inventarios
		beego.NSNamespace("/inventarios",
			beego.NSRouter("/", &controllers.InventarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.InventarioController{}, "get:GetById"),
		),
		// Rutas para nóminas
		beego.NSNamespace("/nominas",
			beego.NSRouter("/", &controllers.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.NominaController{}, "get:GetById"),
		),
		// Rutas para plato_ingredientes
		beego.NSNamespace("/plato_ingredientes",
			beego.NSRouter("/", &controllers.PlatoIngredienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PlatoIngredienteController{}, "get:GetById"),
		),
		// Rutas para pedido_clientes
		beego.NSNamespace("/pedido_clientes",
			beego.NSRouter("/", &controllers.PedidoClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PedidoClienteController{}, "get:GetById"),
		),
	)

	beego.AddNamespace(ns)
}
