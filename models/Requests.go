package models

type HorarioTrabajadorCreateRequest struct {
	DocumentoTrabajador int64  `json:"documentoTrabajador" validate:"required" example:"10000000"`
	Dia                 string `json:"dia" validate:"required" enums:"Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo" example:"Lunes"`
	HoraInicio          string `json:"horaInicio" validate:"required" example:"08:00:00"`
	HoraFin             string `json:"horaFin" validate:"required" example:"12:00:00"`
}

// HorarioTrabajadorUpdateRequest: PUT parcial; los campos ausentes se conservan
// y ninguno admite null (400).
type HorarioTrabajadorUpdateRequest struct {
	HoraInicio *string `json:"horaInicio,omitempty" example:"08:00:00"`
	HoraFin    *string `json:"horaFin,omitempty" example:"12:00:00"`
}

// ProductoCreateRequest es el cuerpo JSON de POST /productos. También se acepta
// multipart/form-data con los mismos campos (imagen como archivo). `imagen` es
// Base64 (se tolera el prefijo data:image/...;base64,). Sin `subcategoriaId` el
// producto queda sin subcategoría (null).
type ProductoCreateRequest struct {
	Nombre         string  `json:"nombre" validate:"required" example:"Bandeja Paisa"`
	Calorias       *int64  `json:"calorias,omitempty" extensions:"x-nullable" example:"850"`
	Descripcion    *string `json:"descripcion,omitempty" extensions:"x-nullable" example:"Descripción del producto"`
	Precio         int64   `json:"precio" validate:"required" minimum:"1" example:"25000"`
	EstadoProducto string  `json:"estadoProducto" validate:"required" enums:"DISPONIBLE,NO_DISPONIBLE" example:"DISPONIBLE"`
	Imagen         string  `json:"imagen,omitempty" example:"BASE64..."`
	Cantidad       int     `json:"cantidad" minimum:"0" example:"10"`
	SubcategoriaId *int64  `json:"subcategoriaId,omitempty" extensions:"x-nullable" example:"1"`
}

// ProductoUpdateRequest es el cuerpo de PUT /productos (merge): los campos
// ausentes se conservan. Admiten null explícito (se limpian): calorias,
// descripcion, imagen y subcategoriaId; null en el resto responde 400.
type ProductoUpdateRequest struct {
	Nombre         *string `json:"nombre,omitempty" example:"Bandeja Paisa"`
	Calorias       *int64  `json:"calorias,omitempty" extensions:"x-nullable" example:"850"`
	Descripcion    *string `json:"descripcion,omitempty" extensions:"x-nullable" example:"Nueva descripción"`
	Precio         *int64  `json:"precio,omitempty" minimum:"1" example:"26000"`
	EstadoProducto *string `json:"estadoProducto,omitempty" enums:"DISPONIBLE,NO_DISPONIBLE" example:"NO_DISPONIBLE"`
	Imagen         *string `json:"imagen,omitempty" extensions:"x-nullable" example:"BASE64..."`
	Cantidad       *int    `json:"cantidad,omitempty" minimum:"0" example:"8"`
	SubcategoriaId *int64  `json:"subcategoriaId,omitempty" extensions:"x-nullable" example:"2"`
}

// PedidoCreateRequest es el cuerpo de POST /pedidos. FECHA, HORA y ESTADO_PEDIDO
// (INICIADO) los fija el servidor; solo se respetan estos campos.
type PedidoCreateRequest struct {
	Delivery         *bool  `json:"delivery,omitempty" example:"false"`              // por defecto false; si es true exige pk_id_domicilio
	PKIDDomicilio    *int64 `json:"pk_id_domicilio,omitempty" example:"3"`           // id de un domicilio existente (entero > 0)
	RestauranteId    int64  `json:"restauranteId,omitempty" example:"1"`             // opcional; si se envía debe existir
	DocumentoCliente *int64 `json:"documentoCliente,omitempty" example:"1234567890"` // opcional; si se envía debe existir
}

// ClienteCreateRequest: registro público de clientes. La contraseña solo se
// recibe aquí (máx. 72 bytes, límite de bcrypt) y nunca se devuelve.
type ClienteCreateRequest struct {
	DocumentoCliente int64   `json:"documentoCliente" validate:"required" example:"1234567890"`
	Nombre           string  `json:"nombre" validate:"required" example:"Juan"`
	Apellido         string  `json:"apellido" validate:"required" example:"Pérez"`
	Correo           string  `json:"correo" validate:"required" format:"email" example:"juan.perez@example.com"`
	Password         string  `json:"password" validate:"required" maxLength:"72" example:"MiPassSegura!"`
	Telefono         string  `json:"telefono" validate:"required" example:"3001234567"`
	Direccion        *string `json:"direccion,omitempty" example:"Calle 123 #45-67"`
	Observaciones    *string `json:"observaciones,omitempty" example:"Cliente frecuente"`
}

// ClienteUpdateRequest: PUT parcial con merge. Los campos ausentes se conservan.
// `observaciones` es el único campo anulable (null lo limpia); null en cualquier
// otro campo responde 400. Una `password` nueva se guarda hasheada.
type ClienteUpdateRequest struct {
	Nombre        *string `json:"nombre,omitempty" example:"Juan"`
	Apellido      *string `json:"apellido,omitempty" example:"Pérez"`
	Correo        *string `json:"correo,omitempty" format:"email" example:"juan.perez@example.com"`
	Password      *string `json:"password,omitempty" maxLength:"72" example:"NuevaPass!"`
	Telefono      *string `json:"telefono,omitempty" example:"3009876543"`
	Direccion     *string `json:"direccion,omitempty" example:"Carrera 10 #20-30"`
	Observaciones *string `json:"observaciones,omitempty" x-nullable:"true" example:"Prefiere contacto por WhatsApp"`
}

// TrabajadorCreateRequest: solo un administrador puede crear trabajadores (y por
// tanto otros administradores). Fechas en YYYY-MM-DD. La contraseña solo se recibe
// aquí (máx. 72 bytes) y nunca se devuelve.
type TrabajadorCreateRequest struct {
	DocumentoTrabajador int64   `json:"documentoTrabajador" validate:"required" example:"10000000"`
	Nombre              string  `json:"nombre" validate:"required" example:"María"`
	Apellido            string  `json:"apellido" validate:"required" example:"Gómez"`
	Rol                 string  `json:"rol" validate:"required" enums:"Administrador,Mesero,Cocinero,Domiciliario,Oficios_varios" example:"Mesero"`
	FechaIngreso        string  `json:"fechaIngreso" validate:"required" example:"2025-01-31"`
	Sueldo              int64   `json:"sueldo" validate:"required" example:"2000000"`
	Password            string  `json:"password" validate:"required" maxLength:"72" example:"Secreta123"`
	Nuevo               *bool   `json:"nuevo,omitempty" example:"false"`
	Telefono            *string `json:"telefono,omitempty" example:"3012223344"`
	RestauranteId       *int64  `json:"restauranteId,omitempty" example:"1"`
	FechaNacimiento     *string `json:"fechaNacimiento,omitempty" example:"1990-05-20"`
}

// TrabajadorUpdateRequest: PUT parcial con merge (solo administrador). Los campos
// ausentes se conservan. Anulables (null los limpia): telefono, fechaNacimiento,
// fechaRetiro y restauranteId; null en cualquier otro campo responde 400.
// Fechas en YYYY-MM-DD.
type TrabajadorUpdateRequest struct {
	Nombre          *string `json:"nombre,omitempty" example:"María"`
	Apellido        *string `json:"apellido,omitempty" example:"Gómez"`
	Rol             *string `json:"rol,omitempty" enums:"Administrador,Mesero,Cocinero,Domiciliario,Oficios_varios" example:"Cocinero"`
	Sueldo          *int64  `json:"sueldo,omitempty" example:"2200000"`
	Nuevo           *bool   `json:"nuevo,omitempty" example:"false"`
	Telefono        *string `json:"telefono,omitempty" x-nullable:"true" example:"3012223344"`
	FechaIngreso    *string `json:"fechaIngreso,omitempty" example:"2025-02-01"`
	FechaRetiro     *string `json:"fechaRetiro,omitempty" x-nullable:"true" example:"2025-12-31"`
	FechaNacimiento *string `json:"fechaNacimiento,omitempty" x-nullable:"true" example:"1990-05-20"`
	RestauranteId   *int64  `json:"restauranteId,omitempty" x-nullable:"true" example:"1"`
	Password        *string `json:"password,omitempty" maxLength:"72" example:"NuevaSecreta!"`
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
	DocumentoTrabajador int64  `json:"documentoTrabajador" validate:"required" example:"10000000"`
	FechaIncidencia     string `json:"fechaIncidencia" validate:"required" example:"2025-01-31"`
	Monto               int64  `json:"monto" validate:"required" example:"50000"`
	Resta               bool   `json:"resta" validate:"required" example:"true"`
	Motivo              string `json:"motivo" validate:"required" example:"Descuento por retraso"`
}

// IncidenciaUpdateRequest: PUT parcial con merge; los campos ausentes se
// conservan y ninguno admite null (400). Fecha en YYYY-MM-DD.
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

// PagoCreateRequest es el cuerpo de POST /pagos. Todos los campos salvo
// updatedBy son obligatorios.
type PagoCreateRequest struct {
	EstadoPago   string `json:"estadoPago" enums:"PAGADO,PENDIENTE,NO_PAGO" example:"PAGADO"`
	FechaPago    string `json:"fechaPago" example:"2025-01-31"` // YYYY-MM-DD
	HoraPago     string `json:"horaPago" example:"14:30:00"`    // HH:MM o HH:MM:SS
	MetodoPagoId int64  `json:"metodoPagoId" example:"1"`       // debe existir
	Monto        int64  `json:"monto" example:"50000"`          // entero > 0
	UpdatedBy    string `json:"updatedBy,omitempty" example:"operador@example.com"`
}

// PagoUpdateRequest es el cuerpo de PUT /pagos (merge: los campos ausentes se
// conservan). Solo `updatedBy` es anulable; null en cualquier otro campo
// responde 400. `fecha` y `hora` son alias heredados de `fechaPago`/`horaPago`
// (si llegan ambos, gana `fechaPago`/`horaPago`).
type PagoUpdateRequest struct {
	FechaPago    *string `json:"fechaPago,omitempty" example:"2025-02-01"` // YYYY-MM-DD
	HoraPago     *string `json:"horaPago,omitempty" example:"15:00:00"`    // HH:MM o HH:MM:SS
	Fecha        *string `json:"fecha,omitempty" example:"2025-02-01"`     // alias de fechaPago
	Hora         *string `json:"hora,omitempty" example:"15:00:00"`        // alias de horaPago
	Monto        *int64  `json:"monto,omitempty" example:"60000"`          // entero > 0
	EstadoPago   *string `json:"estadoPago,omitempty" enums:"PAGADO,PENDIENTE,NO_PAGO" example:"PENDIENTE"`
	MetodoPagoId *int64  `json:"metodoPagoId,omitempty" example:"2"` // debe existir
	UpdatedBy    *string `json:"updatedBy,omitempty" extensions:"x-nullable" example:"operador@example.com"`
}

// DomicilioUpdateRequest es el cuerpo de PUT /domicilios (merge: los campos
// ausentes se conservan). Anulables: observaciones y updatedBy (null los
// limpia); null en direccion, telefono, estado o fechaDomicilio responde 400.
// `estadoDomicilio` es alias de `estado` (si llegan ambos, gana `estado`).
type DomicilioUpdateRequest struct {
	Direccion       *string `json:"direccion,omitempty" example:"Calle 45 #12-34"`
	Telefono        *string `json:"telefono,omitempty" example:"3001112233"`
	Estado          *string `json:"estado,omitempty" enums:"PENDIENTE,EN_CAMINO,ENTREGADO" example:"ENTREGADO"`
	EstadoDomicilio *string `json:"estadoDomicilio,omitempty" enums:"PENDIENTE,EN_CAMINO,ENTREGADO" example:"ENTREGADO"` // alias de estado
	Observaciones   *string `json:"observaciones,omitempty" extensions:"x-nullable" example:"Entregado en portería"`
	FechaDomicilio  *string `json:"fechaDomicilio,omitempty" example:"2025-01-31"` // YYYY-MM-DD
	UpdatedBy       *string `json:"updatedBy,omitempty" extensions:"x-nullable" example:"operador@example.com"`
}

type CambiosHorarioCreateRequest struct {
	FechaCambioHorario string  `json:"fechaCambioHorario" validate:"required" example:"2025-01-31"`
	Abierto            bool    `json:"abierto" validate:"required" example:"true"`
	HoraApertura       *string `json:"horaApertura,omitempty" example:"08:00:00"`
	HoraCierre         *string `json:"horaCierre,omitempty" example:"18:00:00"`
}

// CambiosHorarioUpdateRequest: PUT parcial con merge; los campos ausentes se
// conservan y ninguno admite null (400). Si abierto=false se fuerzan las horas
// 00:00:00 - 23:59:59.
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

// AuthResponse es el `data` de POST /login y POST /auth/refresh. `token` es un
// alias de `access_token`. `expires_in` son los segundos de vida del access token
// (7200 = 120 min, la duración real) serializados como string.
type AuthResponse struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    string `json:"expires_in" example:"7200"`
	Nombre       string `json:"nombre" example:"Juan Pérez"`
}
