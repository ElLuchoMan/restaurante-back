package models

// Estructuras de respuesta SOLO para Swagger: reflejan exactamente lo que
// producen los MarshalJSON de Trabajador, HorarioTrabajador, Incidencia y
// CambiosHorario (nombres json, fechas DD-MM-YYYY, nullables y anidados).
// Los tests de este paquete verifican que coinciden con la salida real.

// RestauranteRef es la referencia mínima al restaurante dentro de un trabajador.
type RestauranteRef struct {
	RestauranteId int64 `json:"restauranteId" example:"1"`
}

// HorarioTrabajadorResponse es un horario semanal tal como lo devuelve el API.
type HorarioTrabajadorResponse struct {
	DocumentoTrabajador int64  `json:"documentoTrabajador" example:"10000000"`
	Dia                 string `json:"dia" enums:"Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo" example:"Lunes"`
	HoraInicio          string `json:"horaInicio" example:"08:00:00"`
	HoraFin             string `json:"horaFin" example:"16:00:00"`
}

// TrabajadorResponse es un trabajador tal como lo devuelve el API (sin password).
// `telefono`, `fechaNacimiento`, `fechaRetiro` y `restauranteId` se omiten cuando son nulos;
// `horarios` solo viene en GET /trabajadores y GET /trabajadores/search.
type TrabajadorResponse struct {
	DocumentoTrabajador int64                       `json:"documentoTrabajador" example:"10000000"`
	Nombre              string                      `json:"nombre" example:"María"`
	Apellido            string                      `json:"apellido" example:"Gómez"`
	Sueldo              int64                       `json:"sueldo" example:"2000000"`
	Telefono            *string                     `json:"telefono,omitempty" example:"3012223344"`
	FechaNacimiento     *string                     `json:"fechaNacimiento,omitempty" example:"20-05-1990"`
	Nuevo               bool                        `json:"nuevo" example:"false"`
	Rol                 string                      `json:"rol" enums:"Administrador,Mesero,Cocinero,Domiciliario,Oficios_varios" example:"Mesero"`
	FechaIngreso        string                      `json:"fechaIngreso" example:"31-01-2025"`
	FechaRetiro         *string                     `json:"fechaRetiro,omitempty" example:"31-12-2025"`
	Horarios            []HorarioTrabajadorResponse `json:"horarios,omitempty"`
	RestauranteId       *RestauranteRef             `json:"restauranteId,omitempty"`
}

// ClienteResumenResponse es el elemento de GET /clientes?fields=nombre_completo_telefono.
type ClienteResumenResponse struct {
	DocumentoCliente int64  `json:"documentoCliente" example:"1234567890"`
	NombreCompleto   string `json:"nombre_completo" example:"Juan Pérez"`
	Telefono         string `json:"telefono" example:"3001234567"`
}

// IncidenciaResponse es una incidencia tal como la devuelve el API.
type IncidenciaResponse struct {
	IncidenciaId        int64  `json:"incidenciaId" example:"1"`
	FechaIncidencia     string `json:"fechaIncidencia" example:"31-01-2025"`
	Monto               int64  `json:"monto" example:"50000"`
	Resta               bool   `json:"resta" example:"true"`
	Motivo              string `json:"motivo" example:"Descuento por retraso"`
	DocumentoTrabajador int64  `json:"documentoTrabajador" example:"10000000"`
}

// CambiosHorarioResponse es un cambio de horario tal como lo devuelve el API.
// `horaApertura` se omite cuando es nula.
type CambiosHorarioResponse struct {
	CambioHorarioId    int64   `json:"cambioHorarioId" example:"1"`
	FechaCambioHorario string  `json:"fechaCambioHorario" example:"31-01-2025"`
	HoraApertura       *string `json:"horaApertura,omitempty" example:"08:00:00"`
	HoraCierre         string  `json:"horaCierre" example:"18:00:00"`
	Abierto            bool    `json:"abierto" example:"true"`
}
