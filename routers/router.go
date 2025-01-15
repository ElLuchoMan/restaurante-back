package routers

import (
	"restaurante/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/restaurante/v1",
		// Ruta para login
		beego.NSRouter("/login", &controllers.LoginController{}, "post:Login"),

		// Rutas para clientes
		beego.NSNamespace("/clientes",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ClienteController{}, "get:GetById"),
		),
		// Rutas para restaurantes
		beego.NSNamespace("/restaurantes",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.RestauranteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.RestauranteController{}, "get:GetById"),
		),
		// Rutas para pedidos
		beego.NSNamespace("/pedidos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PedidoController{}, "get:GetById"),
			beego.NSRouter("/asignar-domicilio", &controllers.PedidoController{}, "post:AssignDomicilio"),
			beego.NSRouter("/asignar-pago", &controllers.PedidoController{}, "post:AssignPago"),
			beego.NSRouter("/actualizar-estado", &controllers.PedidoController{}, "put:UpdateEstadoPedido"),
		),

		// Rutas para domicilios
		beego.NSNamespace("/domicilios",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.DomicilioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.DomicilioController{}, "get:GetById"),
		),
		// Rutas para trabajadores
		beego.NSNamespace("/trabajadores",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.TrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.TrabajadorController{}, "get:GetById"),
		),
		// Rutas para platos
		beego.NSNamespace("/productos",
			beego.NSRouter("/", &controllers.ProductoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ProductoController{}, "get:GetById"),
		),
		// Rutas para reservas
		beego.NSNamespace("/reservas",
			beego.NSRouter("/", &controllers.ReservaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ReservaController{}, "get:GetById"),
		),
		// Rutas para métodos de pago
		beego.NSNamespace("/metodos_pago",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.MetodoPagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.MetodoPagoController{}, "get:GetById"),
		),
		// Rutas para pagos
		beego.NSNamespace("/pagos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PagoController{}, "get:GetById"),
		),
		// Rutas para pedido_clientes
		beego.NSNamespace("/pedido_clientes",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PedidoClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PedidoClienteController{}, "get:GetById"),
		),
		// Rutas para nominas
		beego.NSNamespace("/nominas",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),
		// Rutas para cambios_horario
		beego.NSNamespace("/cambios_horario",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.CambiosHorarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/actual", &controllers.CambiosHorarioController{}, "get:GetByCurrentDate"),
		),
		// Rutas para incidencias
		beego.NSNamespace("/incidencias",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.IncidenciaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.IncidenciaController{}, "get:GetByDocumentAndDate"),
		),
		// Rutas para nóminas de trabajadores
		beego.NSNamespace("/nomina_trabajador",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.NominaTrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.NominaTrabajadorController{}, "get:GetByTrabajador"),
			beego.NSRouter("/mes", &controllers.NominaTrabajadorController{}, "get:GetNominasByMes"),
		),
	)

	beego.AddNamespace(ns)
}
