package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int64            `orm:"column(pk_id_reserva);pk;auto" json:"reservaId"`
	FECHA             time.Time        `orm:"column(fecha);type(date)" json:"fechaReserva"`
	HORA              time.Time        `orm:"column(hora);type(time);size(8)" json:"horaReserva"`
	PERSONAS          int              `orm:"column(personas)" json:"personas"`
	PK_ID_CONTACTO    *ReservaContacto `orm:"column(pk_id_contacto);rel(fk)" json:"contactoId"`
	PK_ID_RESTAURANTE *Restaurante     `orm:"column(pk_id_restaurante);rel(fk)" json:"restauranteId"`
	ESTADO_RESERVA    *EstadoReserva   `orm:"column(estado_reserva);type(estado_reserva);null" json:"estadoReserva,omitempty"`
	INDICACIONES      *string          `orm:"column(indicaciones);null" json:"indicaciones,omitempty"`
	CREATED_AT        time.Time        `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt" swaggertype:"string"`
	UPDATED_AT        time.Time        `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt" swaggertype:"string"`
	CREATED_BY        *string          `orm:"column(created_by);type(text);null" json:"createdBy,omitempty"`
	UPDATED_BY        *string          `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

// RestauranteReservaResponse es el restaurante embebido en una reserva
// (solo con RelatedSel; sin la relación cambioHorarioId).
type RestauranteReservaResponse struct {
	RestauranteID     int64  `json:"restauranteId" example:"1"`
	NombreRestaurante string `json:"nombreRestaurante" example:"Sazón Criolla"`
	HoraApertura      string `json:"horaApertura" example:"08:00:00" description:"HH:MM:SS"`
}

// ReservaResponse es EXACTAMENTE lo que el API devuelve por cada reserva
// (Reserva.MarshalJSON). Fechas de respuesta: fechaReserva DD-MM-YYYY,
// horaReserva HH:MM:SS, createdAt/updatedAt DD-MM-YYYY HH:MM:SS (hora de Bogotá).
type ReservaResponse struct {
	ReservaID     int64                       `json:"reservaId" example:"12"`
	FechaReserva  string                      `json:"fechaReserva" example:"31-01-2025" description:"DD-MM-YYYY"`
	HoraReserva   string                      `json:"horaReserva" example:"18:30:00" description:"HH:MM:SS"`
	Personas      int                         `json:"personas" example:"4"`
	ContactoID    *ReservaContactoResponse    `json:"contactoId"`
	RestauranteID *RestauranteReservaResponse `json:"restauranteId"`
	EstadoReserva *string                     `json:"estadoReserva,omitempty" enums:"PENDIENTE,CONFIRMADA,CANCELADA,CUMPLIDA" example:"PENDIENTE"`
	Indicaciones  *string                     `json:"indicaciones,omitempty" example:"Mesa cerca a la ventana"`
	CreatedAt     string                      `json:"createdAt" example:"31-01-2025 10:15:00" description:"DD-MM-YYYY HH:MM:SS"`
	UpdatedAt     string                      `json:"updatedAt" example:"31-01-2025 10:15:00" description:"DD-MM-YYYY HH:MM:SS"`
	CreatedBy     *string                     `json:"createdBy,omitempty" example:"admin@example.com"`
	UpdatedBy     *string                     `json:"updatedBy,omitempty" example:"operador@example.com"`
}

// ReservaCreateRequest es el cuerpo de POST /reservas. El contacto se resuelve
// con documentoContacto (invitado; requiere nombreCompleto si es nuevo) o con
// documentoCliente (cliente registrado). Si se envían ambos prevalece
// documentoContacto. contactoId NO se acepta.
type ReservaCreateRequest struct {
	DocumentoContacto *int64  `json:"documentoContacto,omitempty" example:"1015466494" description:"Documento del invitado. Obligatorio si no se envía documentoCliente"`
	DocumentoCliente  *int64  `json:"documentoCliente,omitempty" example:"1015466495" description:"Documento de un cliente registrado. Obligatorio si no se envía documentoContacto"`
	NombreCompleto    *string `json:"nombreCompleto,omitempty" example:"Ana Gómez" description:"Obligatorio al crear un contacto nuevo con documentoContacto; se ignora si el contacto ya existe"`
	Telefono          *string `json:"telefono,omitempty" example:"3001234567" description:"Solo al crear un contacto nuevo"`
	RestauranteId     *int64  `json:"restauranteId" binding:"required" example:"1"`
	FechaReserva      *string `json:"fechaReserva" binding:"required" example:"2025-01-31" description:"YYYY-MM-DD"`
	HoraReserva       *string `json:"horaReserva" binding:"required" example:"18:30:00" description:"HH:MM:SS"`
	Personas          *int    `json:"personas" binding:"required" minimum:"1" example:"4"`
	EstadoReserva     *string `json:"estadoReserva,omitempty" enums:"PENDIENTE,CONFIRMADA,CANCELADA,CUMPLIDA" default:"PENDIENTE" example:"PENDIENTE"`
	Indicaciones      *string `json:"indicaciones,omitempty" example:"Mesa cerca a la ventana"`
	CreatedBy         *string `json:"createdBy,omitempty" example:"admin@example.com"`
}

// ReservaUpdateRequest es el cuerpo (parcial) de PUT /reservas: los campos
// ausentes se conservan. Solo indicaciones y updatedBy admiten null (lo
// limpian); null en cualquier otro campo devuelve 400. contactoId NO se
// acepta: para cambiar el contacto envíe documentoContacto o documentoCliente.
type ReservaUpdateRequest struct {
	DocumentoContacto *int64  `json:"documentoContacto,omitempty" example:"1015466494"`
	DocumentoCliente  *int64  `json:"documentoCliente,omitempty" example:"1015466495"`
	NombreCompleto    *string `json:"nombreCompleto,omitempty" example:"Ana Gómez" description:"Solo al crear un contacto nuevo con documentoContacto"`
	Telefono          *string `json:"telefono,omitempty" example:"3001234567" description:"Solo al crear un contacto nuevo"`
	RestauranteId     *int64  `json:"restauranteId,omitempty" example:"1"`
	FechaReserva      *string `json:"fechaReserva,omitempty" example:"2025-01-31" description:"YYYY-MM-DD"`
	HoraReserva       *string `json:"horaReserva,omitempty" example:"19:00:00" description:"HH:MM:SS"`
	Personas          *int    `json:"personas,omitempty" minimum:"1" example:"5"`
	EstadoReserva     *string `json:"estadoReserva,omitempty" enums:"PENDIENTE,CONFIRMADA,CANCELADA,CUMPLIDA" example:"CONFIRMADA"`
	Indicaciones      *string `json:"indicaciones,omitempty" example:"Mesa al fondo" description:"Admite null (limpia el campo)"`
	UpdatedBy         *string `json:"updatedBy,omitempty" example:"operador@example.com" description:"Admite null (limpia el campo)"`
}

func (r *Reserva) TableName() string {
	return "reserva"
}

func init() {
	orm.RegisterModel(new(Reserva))
}

// Response devuelve la representación pública de la reserva.
func (t Reserva) Response() ReservaResponse {
	out := ReservaResponse{
		ReservaID:    t.PK_ID_RESERVA,
		FechaReserva: FormatDateUTC(t.FECHA),
		HoraReserva:  FormatTimeWithLMT(t.HORA),
		Personas:     t.PERSONAS,
		Indicaciones: t.INDICACIONES,
		CreatedAt:    FormatTimestampBogota(t.CREATED_AT),
		UpdatedAt:    FormatTimestampBogota(t.UPDATED_AT),
		CreatedBy:    t.CREATED_BY,
		UpdatedBy:    t.UPDATED_BY,
	}
	if t.ESTADO_RESERVA != nil {
		estado := string(*t.ESTADO_RESERVA)
		out.EstadoReserva = &estado
	}
	if t.PK_ID_CONTACTO != nil {
		contacto := t.PK_ID_CONTACTO.Response()
		out.ContactoID = &contacto
	}
	if t.PK_ID_RESTAURANTE != nil {
		out.RestauranteID = &RestauranteReservaResponse{
			RestauranteID:     t.PK_ID_RESTAURANTE.PK_ID_RESTAURANTE,
			NombreRestaurante: t.PK_ID_RESTAURANTE.NOMBRE_RESTAURANTE,
			HoraApertura:      FormatTimeWithLMT(t.PK_ID_RESTAURANTE.HORA_APERTURA),
		}
	}
	return out
}

func (t Reserva) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Response())
}
