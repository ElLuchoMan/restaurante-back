package models

type HorarioTrabajadorCreateRequest struct {
	DocumentoTrabajador int64  `json:"documentoTrabajador" example:"10000000"`
	Dia                 string `json:"dia" example:"Lunes"`
	HoraInicio          string `json:"horaInicio" example:"08:00:00"`
	HoraFin             string `json:"horaFin" example:"12:00:00"`
}

type HorarioTrabajadorUpdateRequest struct {
	HoraInicio *string `json:"horaInicio,omitempty" example:"08:00:00"`
	HoraFin    *string `json:"horaFin,omitempty" example:"12:00:00"`
}

type ProductoCreateRequest struct {
	Nombre         string  `json:"nombre" example:"Bandeja Paisa"`
	Calorias       *int64  `json:"calorias,omitempty" example:"850"`
	Descripcion    *string `json:"descripcion,omitempty" example:"Descripción del producto"`
	Precio         int64   `json:"precio" example:"25000"`
	EstadoProducto string  `json:"estadoProducto" example:"DISPONIBLE"`
	Imagen         string  `json:"imagen,omitempty" example:"BASE64..."`
	Cantidad       int     `json:"cantidad" example:"10"`
	SubcategoriaId int64   `json:"subcategoriaId" example:"1"`
}

type ProductoUpdateRequest struct {
	Nombre         *string `json:"nombre,omitempty" example:"Bandeja Paisa"`
	Calorias       *int64  `json:"calorias,omitempty" example:"850"`
	Descripcion    *string `json:"descripcion,omitempty" example:"Nueva descripción"`
	Precio         *int64  `json:"precio,omitempty" example:"26000"`
	EstadoProducto *string `json:"estadoProducto,omitempty" example:"NO_DISPONIBLE"`
	Imagen         *string `json:"imagen,omitempty" example:"BASE64..."`
	Cantidad       *int    `json:"cantidad,omitempty" example:"8"`
	SubcategoriaId *int64  `json:"subcategoriaId,omitempty" example:"2"`
}

type ReservaCreateRequest struct {
	ContactoId    int64   `json:"contactoId" example:"1"`
	RestauranteId int64   `json:"restauranteId" example:"1"`
	FechaReserva  string  `json:"fechaReserva" example:"2025-01-31"`
	HoraReserva   string  `json:"horaReserva" example:"18:30:00"`
	Personas      int     `json:"personas" example:"4"`
	EstadoReserva *string `json:"estadoReserva,omitempty" example:"PENDIENTE"`
	Indicaciones  *string `json:"indicaciones,omitempty" example:"Mesa cerca a la ventana"`
	CreatedBy     *string `json:"createdBy,omitempty" example:"admin@example.com"`
}

type ReservaUpdateRequest struct {
	ContactoId    *int64  `json:"contactoId,omitempty" example:"1"`
	RestauranteId *int64  `json:"restauranteId,omitempty" example:"1"`
	FechaReserva  *string `json:"fechaReserva,omitempty" example:"2025-01-31"`
	HoraReserva   *string `json:"horaReserva,omitempty" example:"19:00:00"`
	Personas      *int    `json:"personas,omitempty" example:"5"`
	EstadoReserva *string `json:"estadoReserva,omitempty" example:"CONFIRMADA"`
	Indicaciones  *string `json:"indicaciones,omitempty" example:"Mesa al fondo"`
	UpdatedBy     *string `json:"updatedBy,omitempty" example:"operador@example.com"`
}

type PedidoCreateRequest struct {
	Delivery      *bool  `json:"delivery,omitempty" example:"false"`
	PKIDDomicilio *int64 `json:"pk_id_domicilio,omitempty" example:"0"`
	RestauranteId int64  `json:"restauranteId" example:"1"`
}

type ClienteCreateRequest struct {
	DocumentoCliente int64   `json:"documentoCliente" example:"1234567890"`
	Nombre           string  `json:"nombre" example:"Juan"`
	Apellido         string  `json:"apellido" example:"Pérez"`
	Correo           string  `json:"correo" example:"juan.perez@example.com"`
	Password         string  `json:"password" example:"MiPassSegura!"`
	Telefono         *string `json:"telefono,omitempty" example:"3001234567"`
	Direccion        *string `json:"direccion,omitempty" example:"Calle 123 #45-67"`
	Observaciones    *string `json:"observaciones,omitempty" example:"Cliente frecuente"`
}

type ClienteUpdateRequest struct {
	Nombre        *string `json:"nombre,omitempty" example:"Juan"`
	Apellido      *string `json:"apellido,omitempty" example:"Pérez"`
	Correo        *string `json:"correo,omitempty" example:"juan.perez@example.com"`
	Password      *string `json:"password,omitempty" example:"NuevaPass!"`
	Telefono      *string `json:"telefono,omitempty" example:"3009876543"`
	Direccion     *string `json:"direccion,omitempty" example:"Carrera 10 #20-30"`
	Observaciones *string `json:"observaciones,omitempty" example:"Prefiere contacto por WhatsApp"`
}

type TrabajadorCreateRequest struct {
	DocumentoTrabajador int64   `json:"documentoTrabajador" example:"10000000"`
	Nombre              string  `json:"nombre" example:"María"`
	Apellido            string  `json:"apellido" example:"Gómez"`
	Rol                 string  `json:"rol" example:"Mesero"`
	FechaIngreso        string  `json:"fechaIngreso" example:"2025-01-31"`
	Sueldo              int64   `json:"sueldo" example:"2000000"`
	Password            string  `json:"password" example:"Secreta123"`
	Telefono            *string `json:"telefono,omitempty" example:"3012223344"`
	RestauranteId       *int64  `json:"restauranteId,omitempty" example:"1"`
	FechaNacimiento     *string `json:"fechaNacimiento,omitempty" example:"1990-05-20"`
}

type TrabajadorUpdateRequest struct {
	Nombre          *string `json:"nombre,omitempty" example:"María"`
	Apellido        *string `json:"apellido,omitempty" example:"Gómez"`
	Rol             *string `json:"rol,omitempty" example:"Cocinero"`
	Sueldo          *int64  `json:"sueldo,omitempty" example:"2200000"`
	Nuevo           *bool   `json:"nuevo,omitempty" example:"false"`
	Telefono        *string `json:"telefono,omitempty" example:"3012223344"`
	FechaIngreso    *string `json:"fechaIngreso,omitempty" example:"2025-02-01"`
	FechaRetiro     *string `json:"fechaRetiro,omitempty" example:"2025-12-31"`
	FechaNacimiento *string `json:"fechaNacimiento,omitempty" example:"1990-05-20"`
	Password        *string `json:"password,omitempty" example:"NuevaSecreta!"`
}

type MetodoPagoCreateRequest struct {
	Tipo    string  `json:"tipo" example:"NEQUI"`
	Detalle *string `json:"detalle,omitempty" example:"Cuenta 3001234567"`
}

type MetodoPagoUpdateRequest struct {
	Tipo    *string `json:"tipo,omitempty" example:"DAVIPLATA"`
	Detalle *string `json:"detalle,omitempty" example:"Cuenta 3007654321"`
}

type IncidenciaCreateRequest struct {
	DocumentoTrabajador int64  `json:"documentoTrabajador" example:"10000000"`
	FechaIncidencia     string `json:"fechaIncidencia" example:"2025-01-31"`
	Monto               int64  `json:"monto" example:"50000"`
	Resta               bool   `json:"resta" example:"true"`
	Motivo              string `json:"motivo" example:"Descuento por retraso"`
}

type IncidenciaUpdateRequest struct {
	DocumentoTrabajador *int64  `json:"documentoTrabajador,omitempty" example:"10000000"`
	FechaIncidencia     *string `json:"fechaIncidencia,omitempty" example:"2025-02-01"`
	Monto               *int64  `json:"monto,omitempty" example:"60000"`
	Resta               *bool   `json:"resta,omitempty" example:"false"`
	Motivo              *string `json:"motivo,omitempty" example:"Bonificación"`
}

type ProductoPedidoCreateRequest struct {
	PedidoId int64                     `json:"pedidoId" example:"1"`
	Detalles []ProductoPedidoItemInput `json:"detalles"`
}

type ProductoPedidoItemInput struct {
	ProductoId int64 `json:"productoId" example:"1"`
	Cantidad   int   `json:"cantidad" example:"2"`
}

type ProductoPedidoUpdateRequest []ProductoPedidoItemInput

type PagoCreateRequest struct {
	EstadoPago   string `json:"estadoPago" example:"PAGADO"`
	FechaPago    string `json:"fechaPago" example:"2025-01-31"`
	HoraPago     string `json:"horaPago" example:"14:30:00"`
	MetodoPagoId int64  `json:"metodoPagoId" example:"1"`
	Monto        int64  `json:"monto" example:"50000"`
	UpdatedBy    string `json:"updatedBy,omitempty" example:"operador@example.com"`
}

type PagoUpdateRequest struct {
	Fecha     *string `json:"fecha,omitempty" example:"2025-02-01"`
	Hora      *string `json:"hora,omitempty" example:"15:00:00"`
	Monto     *int64  `json:"monto,omitempty" example:"60000"`
	Estado    *string `json:"estadoPago,omitempty" example:"PENDIENTE"`
	MetodoId  *int64  `json:"metodoPagoId,omitempty" example:"2"`
	UpdatedBy *string `json:"updatedBy,omitempty" example:"operador@example.com"`
}

type DomicilioUpdateRequest struct {
	Direccion *string `json:"direccion,omitempty" example:"Calle 45 #12-34"`
	Telefono  *string `json:"telefono,omitempty" example:"3001112233"`
	UpdatedBy *string `json:"updatedBy,omitempty" example:"operador@example.com"`
}

type RestauranteCreateRequest struct {
	NombreRestaurante string  `json:"nombreRestaurante" example:"El Fogón de María"`
	HoraApertura      *string `json:"horaApertura,omitempty" example:"08:00:00"`
	CambioHorarioId   *int64  `json:"cambioHorarioId,omitempty" example:"1"`
}

type RestauranteUpdateRequest struct {
	NombreRestaurante *string `json:"nombreRestaurante,omitempty" example:"El Fogón Centro"`
	HoraApertura      *string `json:"horaApertura,omitempty" example:"09:00:00"`
	CambioHorarioId   *int64  `json:"cambioHorarioId,omitempty" example:"2"`
}

type NominaCreateRequest struct {
	Fecha        string  `json:"fecha" example:"2025-01-31"`
	EstadoNomina *string `json:"estadoNomina,omitempty" example:"NO_PAGO"`
}

type CambiosHorarioCreateRequest struct {
	FechaCambioHorario string  `json:"fechaCambioHorario" example:"2025-01-31"`
	Abierto            bool    `json:"abierto" example:"true"`
	HoraApertura       *string `json:"horaApertura,omitempty" example:"08:00:00"`
	HoraCierre         *string `json:"horaCierre,omitempty" example:"18:00:00"`
}

type CambiosHorarioUpdateRequest struct {
	FechaCambioHorario *string `json:"fechaCambioHorario,omitempty" example:"2025-02-01"`
	Abierto            *bool   `json:"abierto,omitempty"`
	HoraApertura       *string `json:"horaApertura,omitempty" example:"09:00:00"`
	HoraCierre         *string `json:"horaCierre,omitempty" example:"17:00:00"`
}

type CategoriaCreateRequest struct {
	Nombre string `json:"nombre" example:"Bebidas"`
}

type CategoriaUpdateRequest struct {
	Nombre *string `json:"nombre,omitempty" example:"Bebidas frías"`
}

type SubcategoriaCreateRequest struct {
	Nombre      string `json:"nombre" example:"Gaseosas"`
	CategoriaId int64  `json:"categoriaId" example:"1"`
}

type SubcategoriaUpdateRequest struct {
	Nombre      *string `json:"nombre,omitempty" example:"Gaseosas zero"`
	CategoriaId *int64  `json:"categoriaId,omitempty" example:"1"`
}

// AuthResponse representa la respuesta de autenticación con tokens
type AuthResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    string `json:"expires_in" example:"1800"`
	Nombre       string `json:"nombre" example:"Juan Pérez"`
}
