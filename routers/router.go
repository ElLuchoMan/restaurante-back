package routers

import (
	ch "restaurante/controllers/cambioshorario"
	cat "restaurante/controllers/categoria"
	cli "restaurante/controllers/cliente"
	cn "restaurante/controllers/controlnomina"
	cup "restaurante/controllers/cupon"
	desc "restaurante/controllers/descuento"
	dom "restaurante/controllers/domicilio"
	ht "restaurante/controllers/horariotrabajador"
	inc "restaurante/controllers/incidencia"
	loginc "restaurante/controllers/login"
	mp "restaurante/controllers/metodopago"
	nom "restaurante/controllers/nomina"
	nt "restaurante/controllers/nominatrabajador"
	ofer "restaurante/controllers/oferta"
	pg "restaurante/controllers/pago"
	pd "restaurante/controllers/pedido"
	pph "restaurante/controllers/precioproductohist"
	prod "restaurante/controllers/producto"
	ppd "restaurante/controllers/productopedido"
	push "restaurante/controllers/push"
	resv "restaurante/controllers/reserva"
	rc "restaurante/controllers/reservacontacto"
	rest "restaurante/controllers/restaurante"
	rdia "restaurante/controllers/restaurantedia"
	subc "restaurante/controllers/subcategoria"
	tel "restaurante/controllers/telemetria"
	trab "restaurante/controllers/trabajador"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	// Configurar CORS para permitir conexiones del frontend
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "x-correlation-id"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	// Configurar Swagger UI
	beego.Handler("/swagger/*", httpSwagger.WrapHandler)

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
			beego.NSRouter("/cliente", &resv.ReservaController{}, "get:GetByDocumentoCliente"),
			beego.NSRouter("/documento", &resv.ReservaController{}, "get:GetByDocumento"),
		),
		beego.NSNamespace("/reserva_contacto",
			beego.NSRouter("/", &rc.ReservaContactoController{}, "get:GetAll"),
			beego.NSRouter("/search", &rc.ReservaContactoController{}, "get:GetById"),
		),
		beego.NSNamespace("/categorias",
			beego.NSRouter("/", &cat.CategoriaController{}, "get:GetAll"),
			beego.NSRouter("/search", &cat.CategoriaController{}, "get:GetById"),
		),
	)

	protected := beego.NewNamespace("/restaurante/v1",
		beego.NSRouter("/login", &loginc.LoginController{}, "post:Login"),
		beego.NSRouter("/auth/refresh", &loginc.LoginController{}, "post:RefreshToken"),

		// Endpoints públicos (sin autenticación) - DEBEN IR ANTES de los NSBefore
		beego.NSRouter("/productos-populares", &tel.TelemetriaController{}, "get:GetProductosPopulares"),
		beego.NSRouter("/estados-pedidos", &tel.TelemetriaController{}, "get:GetEstadosPedidos"),
		beego.NSRouter("/productos-disponibles", &tel.TelemetriaController{}, "get:GetProductosDisponibles"),

		beego.NSNamespace("/producto_pedido",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ppd.ProductoPedidoController{}, "get:GetAll;post:Post;put:Update"),
		),

		beego.NSNamespace("/productos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &prod.ProductoController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/restaurantes",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &rest.RestauranteController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/subcategorias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &subc.SubcategoriaController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/trabajadores",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &trab.TrabajadorController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/reservas",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &resv.ReservaController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/precio_producto_hist",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pph.PrecioProductoHistController{}, "get:GetAll"),
			beego.NSRouter("/search", &pph.PrecioProductoHistController{}, "get:GetById"),
		),

		beego.NSNamespace("/pedidos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pd.PedidoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/asignar-domicilio", &pd.PedidoController{}, "post:AssignDomicilio"),
			beego.NSRouter("/asignar-pago", &pd.PedidoController{}, "post:AssignPago"),
			beego.NSRouter("/actualizar-estado", &pd.PedidoController{}, "put:UpdateEstadoPedido"),
			beego.NSRouter("/detalles", &pd.PedidoController{}, "get:GetPedidoDetails"),
		),

		beego.NSNamespace("/cambios_horario",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ch.CambiosHorarioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/actual", &ch.CambiosHorarioController{}, "get:GetByCurrentDate"),
		),

		beego.NSNamespace("/categorias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cat.CategoriaController{}, "post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/clientes",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cli.ClienteController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &cli.ClienteController{}, "get:GetById"),
		),

		beego.NSNamespace("/control_nomina",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cn.ControlNominaController{}, "get:GetAll"),
			beego.NSRouter("/search", &cn.ControlNominaController{}, "get:GetById"),
		),

		beego.NSNamespace("/domicilios",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &dom.DomicilioController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &dom.DomicilioController{}, "get:GetById"),
			beego.NSRouter("/asignar", &dom.DomicilioController{}, "post:AsignarDomiciliario"),
		),

		beego.NSNamespace("/horario_trabajador",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ht.HorarioTrabajadorController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/incidencias",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &inc.IncidenciaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &inc.IncidenciaController{}, "get:GetByDocumentAndDate"),
		),

		beego.NSNamespace("/metodos_pago",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &mp.MetodoPagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &mp.MetodoPagoController{}, "get:GetById"),
		),

		beego.NSNamespace("/nominas",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &nom.NominaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
		),

		beego.NSNamespace("/nomina_trabajador",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &nt.NominaTrabajadorController{}, "get:GetAll;post:Post"),
			beego.NSRouter("/search", &nt.NominaTrabajadorController{}, "get:GetByTrabajador"),
			beego.NSRouter("/mes", &nt.NominaTrabajadorController{}, "get:GetNominasByMes"),
		),

		beego.NSNamespace("/pagos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &pg.PagoController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &pg.PagoController{}, "get:GetById"),
		),

		beego.NSNamespace("/telemetria",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/dashboard", &tel.TelemetriaController{}, "get:GetDashboard"),
			beego.NSRouter("/sales", &tel.TelemetriaController{}, "get:GetSales"),
			beego.NSRouter("/products", &tel.TelemetriaController{}, "get:GetProducts"),
			beego.NSRouter("/users", &tel.TelemetriaController{}, "get:GetUsers"),
			beego.NSRouter("/time-analysis", &tel.TelemetriaController{}, "get:GetTimeAnalysis"),
			// Nuevos endpoints de métricas avanzadas
			beego.NSRouter("/rentabilidad", &tel.TelemetriaController{}, "get:GetRentabilidad"),
			beego.NSRouter("/segmentacion", &tel.TelemetriaController{}, "get:GetSegmentacion"),
			beego.NSRouter("/eficiencia", &tel.TelemetriaController{}, "get:GetEficiencia"),
			// Endpoints adicionales de análisis
			beego.NSRouter("/reservas-analisis", &tel.TelemetriaController{}, "get:GetReservasAnalisis"),
			beego.NSRouter("/pedidos-analisis", &tel.TelemetriaController{}, "get:GetPedidosAnalisis"),
		),
	)

	// Nuevas rutas API v1 para notificaciones, cupones y ofertas
	apiv1 := beego.NewNamespace("/restaurante/v1",
		// Ruta pública para ofertas activas (sin autenticación)
		beego.NSRouter("/ofertas/activas", &ofer.OfertaController{}, "get:ObtenerOfertasActivas"),

		// Rutas protegidas
		beego.NSNamespace("/push",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/dispositivos", &push.PushController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/dispositivos/search", &push.PushController{}, "get:GetById"),
			beego.NSRouter("/dispositivos/visto", &push.PushController{}, "patch:ActualizarUltimaVista"),
			beego.NSRouter("/dispositivos/topics", &push.PushController{}, "patch:ActualizarTopics"),
			beego.NSRouter("/envios", &push.PushController{}, "get:ListarEnvios;post:RegistrarEnvio"),
			beego.NSRouter("/enviar", &push.PushController{}, "post:EnviarNotificacion"),
		),

		beego.NSNamespace("/cupones",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &cup.CuponController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &cup.CuponController{}, "get:GetById"),
			beego.NSRouter("/validar", &cup.CuponController{}, "post:ValidarCupon"),
			beego.NSRouter("/redimir", &cup.CuponController{}, "post:RedimirCupon"),
			beego.NSRouter("/redenciones", &cup.CuponController{}, "get:ListarRedenciones"),
		),

		beego.NSNamespace("/ofertas",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/", &ofer.OfertaController{}, "get:GetAll;post:Post;put:Put;delete:Delete"),
			beego.NSRouter("/search", &ofer.OfertaController{}, "get:GetById"),
			beego.NSRouter("/productos", &ofer.OfertaController{}, "post:AsociarProducto;delete:DesasociarProducto"),
		),

		beego.NSNamespace("/descuentos",
			beego.NSBefore(loginc.ValidateToken),
			beego.NSRouter("/pedidos", &desc.DescuentoController{}, "get:GetAll;post:Post"),
		),
	)

	beego.AddNamespace(public)
	beego.AddNamespace(protected)
	beego.AddNamespace(apiv1)
}
