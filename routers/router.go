package routers

import (
	ch "restaurante/controllers/cambioshorario"
	cat "restaurante/controllers/categoria"
	cli "restaurante/controllers/cliente"
	cn "restaurante/controllers/controlnomina"
	dom "restaurante/controllers/domicilio"
	ht "restaurante/controllers/horariotrabajador"
	inc "restaurante/controllers/incidencia"
	loginc "restaurante/controllers/login"
	mp "restaurante/controllers/metodopago"
	nom "restaurante/controllers/nomina"
	nt "restaurante/controllers/nominatrabajador"
	pg "restaurante/controllers/pago"
	pd "restaurante/controllers/pedido"
	pph "restaurante/controllers/precioproductohist"
	prod "restaurante/controllers/producto"
	ppd "restaurante/controllers/productopedido"
	resv "restaurante/controllers/reserva"
	rc "restaurante/controllers/reservacontacto"
	rest "restaurante/controllers/restaurante"
	rdia "restaurante/controllers/restaurantedia"
	subc "restaurante/controllers/subcategoria"
	trab "restaurante/controllers/trabajador"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// ===== Namespace PÚBLICO (primero en orden) =====
	public := beego.NewNamespace("/restaurante/v1",
		beego.NSNamespace("/productos",
			beego.NSRouter("/", &prod.ProductoController{}, "get:GetAll"),
			beego.NSRouter("/search", &prod.ProductoController{}, "get:GetById"),
		),
		beego.NSNamespace("/restaurantes",
			beego.NSRouter("/", &rest.RestauranteController{}, "get:GetAll"),
			beego.NSRouter("/search", &rest.RestauranteController{}, "get:GetById"),
		),
		beego.NSNamespace("/restaurante_dia",
			beego.NSRouter("/", &rdia.RestauranteDiaController{}, "get:GetAll"),
			beego.NSRouter("/search", &rdia.RestauranteDiaController{}, "get:GetById"),
		),
		beego.NSNamespace("/subcategorias",
			beego.NSRouter("/", &subc.SubcategoriaController{}, "get:GetAll"),
			beego.NSRouter("/search", &subc.SubcategoriaController{}, "get:GetById"),
		),
		beego.NSNamespace("/trabajadores",
			beego.NSRouter("/", &trab.TrabajadorController{}, "get:GetAll"),
			beego.NSRouter("/search", &trab.TrabajadorController{}, "get:GetById"),
		),
		beego.NSNamespace("/reservas",
			beego.NSRouter("/", &resv.ReservaController{}, "get:GetAll"),
			beego.NSRouter("/search", &resv.ReservaController{}, "get:GetById"),
			beego.NSRouter("/parameter", &resv.ReservaController{}, "get:GetByParameter"),
		),
		beego.NSNamespace("/reserva_contacto",
			beego.NSRouter("/", &rc.ReservaContactoController{}, "get:GetAll"),
			beego.NSRouter("/search", &rc.ReservaContactoController{}, "get:GetById"),
		),
		// Categorías: GET públicos
		beego.NSNamespace("/categorias",
			beego.NSRouter("/", &cat.CategoriaController{}, "get:GetAll"),
			beego.NSRouter("/search", &cat.CategoriaController{}, "get:GetById"),
		),
	)

	// ===== Namespace PROTEGIDO (después del público) =====
	protected := beego.NewNamespace("/restaurante/v1",
		// Login (público)
		beego.NSRouter("/login", &loginc.LoginController{}, "post:Login"),

		// Producto-pedido (protegido)
		beego.NSNamespace("/producto_pedido",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ppd.ProductoPedidoController{}, "get:GetAll;post:Post;put:Update"),
		),

		// Productos (protegido: escritura)
		beego.NSNamespace("/productos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &prod.ProductoController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Restaurantes (protegido: escritura)
		beego.NSNamespace("/restaurantes",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &rest.RestauranteController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Subcategorías (protegido: escritura)
		beego.NSNamespace("/subcategorias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &subc.SubcategoriaController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Trabajadores (protegido: escritura)
		beego.NSNamespace("/trabajadores",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &trab.TrabajadorController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Reservas (protegido: escritura)
		beego.NSNamespace("/reservas",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &resv.ReservaController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Historial de precios de producto (protegido)
		beego.NSNamespace("/precio_producto_hist",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pph.PrecioProductoHistController{}, "get:GetAll"),
			beego.NSRouter("/search", &pph.PrecioProductoHistController{}, "get:GetById"),
		),

		// Pedidos (protegido)
		beego.NSNamespace("/pedidos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pd.PedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/asignar-domicilio", &pd.PedidoController{}, "post:AssignDomicilio"),
			beego.NSRouter("/asignar-pago", &pd.PedidoController{}, "post:AssignPago"),
			beego.NSRouter("/actualizar-estado", &pd.PedidoController{}, "put:UpdateEstadoPedido"),
			beego.NSRouter("/detalles", &pd.PedidoController{}, "get:GetPedidoDetails"),
		),

		// Cambios de horario (protegido: GET/POST/PUT/DELETE)
		beego.NSNamespace("/cambios_horario",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ch.CambiosHorarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/actual", &ch.CambiosHorarioController{}, "get:GetByCurrentDate"),
		),

		// Categorías (protegido: escritura)
		beego.NSNamespace("/categorias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cat.CategoriaController{}, "post:Post;put:Put;delete:Delete"),
		),

		// Clientes (todo privado)
		beego.NSNamespace("/clientes",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cli.ClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &cli.ClienteController{}, "get:GetById"),
		),

		// Control de nómina (todo privado)
		beego.NSNamespace("/control_nomina",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cn.ControlNominaController{}, "get:GetAll"),
			beego.NSRouter("/search", &cn.ControlNominaController{}, "get:GetById"),
		),

		// Domicilios (todo privado)
		beego.NSNamespace("/domicilios",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &dom.DomicilioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &dom.DomicilioController{}, "get:GetById"),
			beego.NSRouter("/asignar", &dom.DomicilioController{}, "post:AsignarDomiciliario"),
		),

		// Horario Trabajador (todo privado)
		beego.NSNamespace("/horario_trabajador",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ht.HorarioTrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		// Incidencias (todo privado)
		beego.NSNamespace("/incidencias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &inc.IncidenciaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &inc.IncidenciaController{}, "get:GetByDocumentAndDate"),
		),

		// Métodos de pago (todo privado)
		beego.NSNamespace("/metodos_pago",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &mp.MetodoPagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &mp.MetodoPagoController{}, "get:GetById"),
		),

		// Nómina (todo privado)
		beego.NSNamespace("/nominas",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &nom.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		// Nómina Trabajador (todo privado)
		beego.NSNamespace("/nomina_trabajador",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &nt.NominaTrabajadorController{}, "get:GetAll;post:Post"),
			beego.NSRouter("/search", &nt.NominaTrabajadorController{}, "get:GetByTrabajador"),
			beego.NSRouter("/mes", &nt.NominaTrabajadorController{}, "get:GetNominasByMes"),
		),

		// Pagos (todo privado)
		beego.NSNamespace("/pagos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pg.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &pg.PagoController{}, "get:GetById"),
		),
	)

	// IMPORTANTE: agregar namespaces en orden: primero público, luego protegido
	beego.AddNamespace(public)
	beego.AddNamespace(protected)
}
