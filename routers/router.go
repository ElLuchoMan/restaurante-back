package routers

import (
	"restaurante/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// ===== Namespace PÚBLICO (primero en orden) =====
	// Sólo expone POST /restaurante/v1/clientes sin middleware de token
	public := beego.NewNamespace("/restaurante/v1",
		beego.NSNamespace("/clientes",
			// OJO: aquí NO hay NSBefore(ValidateToken)
			beego.NSRouter("/", &controllers.ClienteController{}, "options:Options;post:Post"),
		),
	)

	// ===== Namespace PROTEGIDO (después del público) =====
	protected := beego.NewNamespace("/restaurante/v1",
		// Login (público)
		beego.NSRouter("/login", &controllers.LoginController{}, "post:Login"),

		// Clientes protegidos (NO incluir post:Post aquí)
		beego.NSNamespace("/clientes",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ClienteController{}, "get:GetAll;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ClienteController{}, "get:GetById"),
		),

		// Restaurantes
		beego.NSNamespace("/restaurantes",
			beego.NSRouter("/", &controllers.RestauranteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.RestauranteController{}, "get:GetById"),
		),

		// Pedidos (protegido)
		beego.NSNamespace("/pedidos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/asignar-domicilio", &controllers.PedidoController{}, "post:AssignDomicilio"),
			beego.NSRouter("/asignar-pago", &controllers.PedidoController{}, "post:AssignPago"),
			beego.NSRouter("/actualizar-estado", &controllers.PedidoController{}, "put:UpdateEstadoPedido"),
			beego.NSRouter("/detalles", &controllers.PedidoController{}, "get:GetPedidoDetails"),
		),

		// Domicilios (protegido)
		beego.NSNamespace("/domicilios",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.DomicilioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.DomicilioController{}, "get:GetById"),
			beego.NSRouter("/asignar", &controllers.DomicilioController{}, "post:AsignarDomiciliario"),
		),

		// Trabajadores (protegido)
		beego.NSNamespace("/trabajadores",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.TrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.TrabajadorController{}, "get:GetById"),
		),

		// Horario de trabajador (protegido)
		beego.NSNamespace("/horario_trabajador",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.HorarioTrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		// Productos (público según tu definición actual)
		beego.NSNamespace("/productos",
			beego.NSRouter("/", &controllers.ProductoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ProductoController{}, "get:GetById"),
		),

		// Reservas (público según tu definición actual)
		beego.NSNamespace("/reservas",
			beego.NSRouter("/", &controllers.ReservaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.ReservaController{}, "get:GetById"),
			beego.NSRouter("/parameter", &controllers.ReservaController{}, "get:GetByParameter"),
		),

		// Métodos de pago (protegido)
		beego.NSNamespace("/metodos_pago",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.MetodoPagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.MetodoPagoController{}, "get:GetById"),
		),

		// Pagos (protegido)
		beego.NSNamespace("/pagos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PagoController{}, "get:GetById"),
		),

		// Pedido_Clientes (protegido)
		beego.NSNamespace("/pedido_clientes",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PedidoClienteController{}, "get:GetAll;post:Post"),
		),

		// Nóminas (protegido)
		beego.NSNamespace("/nominas",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		// Cambios de horario (público según tu definición actual)
		beego.NSNamespace("/cambios_horario",
			beego.NSRouter("/", &controllers.CambiosHorarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/actual", &controllers.CambiosHorarioController{}, "get:GetByCurrentDate"),
		),

		// Incidencias (protegido)
		beego.NSNamespace("/incidencias",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.IncidenciaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.IncidenciaController{}, "get:GetByDocumentAndDate"),
		),

		// Nómina trabajador (protegido)
		beego.NSNamespace("/nomina_trabajador",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.NominaTrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.NominaTrabajadorController{}, "get:GetByTrabajador"),
			beego.NSRouter("/mes", &controllers.NominaTrabajadorController{}, "get:GetNominasByMes"),
		),

		// Producto_pedido (protegido)
		beego.NSNamespace("/producto_pedido",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ProductoPedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),
	)

	// IMPORTANTE: agregar namespaces en orden: primero público, luego protegido
	beego.AddNamespace(public)
	beego.AddNamespace(protected)
}
