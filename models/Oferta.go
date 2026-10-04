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

func (o *Oferta) BeforeInsert() {
	o.serializeDiasSemana()
}

func (o *Oferta) BeforeUpdate() {
	o.serializeDiasSemana()
}

func (o *Oferta) AfterLoad() {
	o.deserializeDiasSemana()
}

func (o *Oferta) serializeDiasSemana() {
	if len(o.DiasSemanaArray) == 0 {
		o.DiasSemana = ""
		return
	}
	jsonBytes, _ := json.Marshal(o.DiasSemanaArray)
	o.DiasSemana = string(jsonBytes)
}

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

func (o Oferta) MarshalJSON() ([]byte, error) {

	fechaInicioStr := FormatDateUTC(o.FechaInicio)
	fechaFinStr := FormatDateUTC(o.FechaFin)

	var horaInicioStr *string
	if o.HoraInicio != nil {
		h := *o.HoraInicio
		s := FormatTimeWithLMT(h)
		horaInicioStr = &s
	}

	var horaFinStr *string
	if o.HoraFin != nil {
		h := *o.HoraFin
		s := FormatTimeWithLMT(h)
		horaFinStr = &s
	}

	return json.Marshal(&struct {
		PkIdOferta      int64         `json:"ofertaId"`
		Titulo          string        `json:"titulo"`
		TipoDescuento   TipoDescuento `json:"tipoDescuento"`
		ValorDescuento  int64         `json:"valorDescuento"`
		FechaInicio     string        `json:"fechaInicio"`
		FechaFin        string        `json:"fechaFin"`
		DiasSemana      string        `json:"-"`
		DiasSemanaArray []string      `json:"diasSemana" swaggertype:"array,string"`
		HoraInicio      *string       `json:"horaInicio,omitempty"`
		HoraFin         *string       `json:"horaFin,omitempty"`
		Activo          bool          `json:"activo"`
		PkIdRestaurante *Restaurante  `json:"restauranteId" swaggertype:"integer"`
	}{
		PkIdOferta:      o.PkIdOferta,
		Titulo:          o.Titulo,
		TipoDescuento:   o.TipoDescuento,
		ValorDescuento:  o.ValorDescuento,
		FechaInicio:     fechaInicioStr,
		FechaFin:        fechaFinStr,
		DiasSemana:      o.DiasSemana,
		DiasSemanaArray: o.DiasSemanaArray,
		HoraInicio:      horaInicioStr,
		HoraFin:         horaFinStr,
		Activo:          o.Activo,
		PkIdRestaurante: o.PkIdRestaurante,
	})
}
