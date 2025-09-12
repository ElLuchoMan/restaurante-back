package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type DomicilioCreate struct {
	Direccion      string          `json:"direccion" example:"Calle 123 #45-67"`
	Telefono       string          `json:"telefono" example:"3001234567"`
	FechaDomicilio string          `json:"fechaDomicilio" example:"2025-01-31"`
	Observaciones  *string         `json:"observaciones,omitempty" example:"Dejar en portería"`
	CreatedBy      *string         `json:"createdBy,omitempty" example:"admin@example.com"`
	TrabajadorID   *int64          `json:"trabajadorAsignado,omitempty" example:"0"`
	Estado         EstadoDomicilio `json:"estadoDomicilio,omitempty" example:"PENDIENTE"`
}

type Domicilio struct {
	ID         int64           `orm:"column(pk_id_domicilio);pk;auto" json:"domicilioId"`
	Direccion  string          `orm:"column(direccion);type(text)" json:"direccion"`
	Telefono   string          `orm:"column(telefono);type(text)" json:"telefono"`
	Estado     EstadoDomicilio `orm:"column(estado_domicilio);type(estado_domicilio)" json:"estadoDomicilio"`
	Entregado  bool            `orm:"column(entregado);type(boolean);null" json:"entregado"`
	Fecha      time.Time       `orm:"column(fecha);type(date)" json:"fechaDomicilio"`
	Observ     *string         `orm:"column(observaciones);type(text);null" json:"observaciones,omitempty"`
	CreatedAt  time.Time       `orm:"column(created_at);type(timestamptz);auto_now_add;null" json:"createdAt"`
	UpdatedAt  time.Time       `orm:"column(updated_at);type(timestamptz);auto_now;null" json:"updatedAt"`
	CreatedBy  *string         `orm:"column(created_by);type(text);null" json:"createdBy,omitempty"`
	UpdatedBy  *string         `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
	Trabajador *Trabajador     `orm:"column(pk_documento_trabajador);rel(fk);null" json:"trabajadorAsignado,omitempty" swaggertype:"integer"`
}

func (d *Domicilio) TableName() string { return "domicilio" }

func init() {
	orm.RegisterModel(new(Domicilio))
}
