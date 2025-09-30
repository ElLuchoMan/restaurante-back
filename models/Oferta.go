package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Oferta struct {
	PkIdOferta      int64         `orm:"column(pk_id_oferta);pk;auto" json:"ofertaId"`
	Titulo          string        `orm:"column(titulo);type(text);unique" json:"titulo"`
	TipoDescuento   TipoDescuento `orm:"column(tipo_descuento);type(tipo_descuento)" json:"tipoDescuento"`
	ValorDescuento  int64         `orm:"column(valor_descuento);type(bigint)" json:"valorDescuento"`
	FechaInicio     time.Time     `orm:"column(fecha_inicio);type(date)" json:"fechaInicio"`
	FechaFin        time.Time     `orm:"column(fecha_fin);type(date)" json:"fechaFin"`
	DiasSemana      string        `orm:"column(dias_semana);type(text);null" json:"-"`
	DiasSemanaArray []string      `orm:"-" json:"diasSemana" swaggertype:"array,string"`
	HoraInicio      *time.Time    `orm:"column(hora_inicio);type(time);null" json:"horaInicio,omitempty"`
	HoraFin         *time.Time    `orm:"column(hora_fin);type(time);null" json:"horaFin,omitempty"`
	Activo          bool          `orm:"column(activo);type(boolean);default(true)" json:"activo"`
	PkIdRestaurante *Restaurante  `orm:"column(pk_id_restaurante);rel(fk)" json:"restauranteId" swaggertype:"integer"`
}

func (o *Oferta) TableName() string {
	return "oferta"
}

// BeforeInsert se ejecuta antes de insertar en la base de datos
func (o *Oferta) BeforeInsert() {
	o.serializeDiasSemana()
}

// BeforeUpdate se ejecuta antes de actualizar en la base de datos
func (o *Oferta) BeforeUpdate() {
	o.serializeDiasSemana()
}

// AfterLoad se ejecuta después de cargar desde la base de datos
func (o *Oferta) AfterLoad() {
	o.deserializeDiasSemana()
}

// serializeDiasSemana convierte el array a string para la base de datos
func (o *Oferta) serializeDiasSemana() {
	if len(o.DiasSemanaArray) == 0 {
		o.DiasSemana = ""
		return
	}
	jsonBytes, _ := json.Marshal(o.DiasSemanaArray)
	o.DiasSemana = string(jsonBytes)
}

// deserializeDiasSemana convierte el string de la base de datos a array
func (o *Oferta) deserializeDiasSemana() {
	if o.DiasSemana == "" {
		o.DiasSemanaArray = []string{}
		return
	}
	_ = json.Unmarshal([]byte(o.DiasSemana), &o.DiasSemanaArray)
}

func init() {
	orm.RegisterModel(new(Oferta))
}
