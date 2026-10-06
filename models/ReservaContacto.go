package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type ReservaContacto struct {
	PKIDContacto       int64    `orm:"column(pk_id_contacto);pk;auto" json:"contactoId"`
	NombreCompleto     string   `orm:"column(nombre_completo);type(text)" json:"nombreCompleto"`
	Telefono           *string  `orm:"column(telefono);type(text);null" json:"telefono,omitempty"`
	DocumentoContacto  *int64   `orm:"column(documento_contacto);null" json:"documentoContacto,omitempty"`
	PKDocumentoCliente *Cliente `orm:"column(pk_documento_cliente);rel(fk);null" json:"documentoCliente,omitempty"`
}

// ClienteRefResponse es la referencia al cliente registrado dentro de un
// contacto: solo su documento (nunca datos del cliente ni contraseña).
type ClienteRefResponse struct {
	DocumentoCliente int64 `json:"documentoCliente" example:"1015466495"`
}

// ReservaContactoResponse es EXACTAMENTE lo que el API devuelve por cada
// contacto de reserva (ReservaContacto.MarshalJSON). documentoContacto se
// envía para invitados y documentoCliente para clientes registrados.
type ReservaContactoResponse struct {
	ContactoID        int64               `json:"contactoId" example:"3"`
	NombreCompleto    string              `json:"nombreCompleto" example:"Ana Gómez"`
	Telefono          *string             `json:"telefono,omitempty" example:"3001234567"`
	DocumentoContacto *int64              `json:"documentoContacto,omitempty" example:"1015466494"`
	DocumentoCliente  *ClienteRefResponse `json:"documentoCliente,omitempty"`
}

func (r *ReservaContacto) TableName() string {
	return "reserva_contacto"
}

func init() {
	orm.RegisterModel(new(ReservaContacto))
}

// Response devuelve la representación pública del contacto.
func (r ReservaContacto) Response() ReservaContactoResponse {
	out := ReservaContactoResponse{
		ContactoID:        r.PKIDContacto,
		NombreCompleto:    r.NombreCompleto,
		Telefono:          r.Telefono,
		DocumentoContacto: r.DocumentoContacto,
	}
	if r.PKDocumentoCliente != nil {
		out.DocumentoCliente = &ClienteRefResponse{DocumentoCliente: r.PKDocumentoCliente.PK_DOCUMENTO_CLIENTE}
	}
	return out
}

func (r ReservaContacto) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Response())
}
