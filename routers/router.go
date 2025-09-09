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
		// Productos: público para lectura (menú)
		beego.NSNamespace("/productos",
			beego.NSRouter("/", &controllers.ProductoController{}, "get:GetAll"),
			beego.NSRouter("/search", &controllers.ProductoController{}, "get:GetById"),
		),
		// Reservas: público GET/POST (crear reservas y consultar)
		beego.NSNamespace("/reservas",
			beego.NSRouter("/", &controllers.ReservaController{}, "get:GetAll;post:Post"),
			beego.NSRouter("/search", &controllers.ReservaController{}, "get:GetById"),
			beego.NSRouter("/parameter", &controllers.ReservaController{}, "get:GetByParameter"),
		),
		// Cambios de horario: público sólo consulta del día actual
		beego.NSNamespace("/cambios_horario",
			beego.NSRouter("/actual", &controllers.CambiosHorarioController{}, "get:GetByCurrentDate"),
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

		// Productos (protegido para escritura)
		beego.NSNamespace("/productos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ProductoController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Categorías (protegido)
		beego.NSNamespace("/categorias",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.CategoriaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.CategoriaController{}, "get:GetById"),
		),

		// Subcategorías (protegido)
		beego.NSNamespace("/subcategorias",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.SubcategoriaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.SubcategoriaController{}, "get:GetById"),
		),

		// Precio producto hist (protegido)
		beego.NSNamespace("/precio_producto_hist",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PrecioProductoHistController{}, "get:GetAll"),
			beego.NSRouter("/search", &controllers.PrecioProductoHistController{}, "get:GetById"),
		),

		// Control nómina (protegido)
		beego.NSNamespace("/control_nomina",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ControlNominaController{}, "get:GetAll"),
			beego.NSRouter("/search", &controllers.ControlNominaController{}, "get:GetById"),
		),

		// Restaurante día (protegido)
		beego.NSNamespace("/restaurante_dia",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.RestauranteDiaController{}, "get:GetAll"),
			beego.NSRouter("/search", &controllers.RestauranteDiaController{}, "get:GetById"),
		),

		// Reserva contacto (protegido)
		beego.NSNamespace("/reserva_contacto",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ReservaContactoController{}, "get:GetAll"),
			beego.NSRouter("/search", &controllers.ReservaContactoController{}, "get:GetById"),
		),

		// Reservas (protegido para escritura)
		beego.NSNamespace("/reservas",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.ReservaController{}, "put:Put;delete:Delete"),
		),

		// Cambios de horario (protegido para gestión)
		beego.NSNamespace("/cambios_horario",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.CambiosHorarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		// Pagos (protegido)
		beego.NSNamespace("/pagos",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &controllers.PagoController{}, "get:GetById"),
		),

		// Nóminas (protegido)
		beego.NSNamespace("/nominas",
			beego.NSBefore(controllers.ValidateToken),
			beego.NSRouter("/", &controllers.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
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
			beego.NSRouter("/", &controllers.ProductoPedidoController{}, "get:GetAll;post:Post;put:Update;delete:Delete"),
		),
	)

	// IMPORTANTE: agregar namespaces en orden: primero público, luego protegido
	beego.AddNamespace(public)
	beego.AddNamespace(protected)
}
