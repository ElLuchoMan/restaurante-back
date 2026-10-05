package models

import (
	"encoding/json"
	"strings"
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
	PkIdRestaurante *Restaurante  `orm:"column(pk_id_restaurante);rel(fk)" json:"restauranteId"`
}

func (o *Oferta) TableName() string {
	return "oferta"
}

// BeforeInsert / BeforeUpdate / AfterLoad NO los invoca el ORM de Beego: los
// controladores y servicios deben llamarlos explícitamente (antes de
// Insert/Update y después de Read/All).
func (o *Oferta) BeforeInsert() {
	o.serializeDiasSemana()
}

func (o *Oferta) BeforeUpdate() {
	o.serializeDiasSemana()
}

func (o *Oferta) AfterLoad() {
	o.deserializeDiasSemana()
}

// serializeDiasSemana escribe los días como literal de arreglo de PostgreSQL
// ({"Lunes","Martes"}), el formato de la columna dias_semana (dia_semana[]).
func (o *Oferta) serializeDiasSemana() {
	quoted := make([]string, 0, len(o.DiasSemanaArray))
	for _, d := range o.DiasSemanaArray {
		d = strings.ReplaceAll(d, `\`, `\\`)
		d = strings.ReplaceAll(d, `"`, `\"`)
		quoted = append(quoted, `"`+d+`"`)
	}
	o.DiasSemana = "{" + strings.Join(quoted, ",") + "}"
}

// deserializeDiasSemana acepta el literal de PostgreSQL ({Lunes,"Miércoles"}) y,
// por compatibilidad, un arreglo JSON (["Lunes"]). Siempre deja un slice no nulo.
func (o *Oferta) deserializeDiasSemana() {
	o.DiasSemanaArray = []string{}
	raw := strings.TrimSpace(o.DiasSemana)
	switch {
	case strings.HasPrefix(raw, "["):
		var dias []string
		if json.Unmarshal([]byte(raw), &dias) == nil && dias != nil {
			o.DiasSemanaArray = dias
		}
	case strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}"):
		o.DiasSemanaArray = parsePGTextArray(raw[1 : len(raw)-1])
	}
}

// parsePGTextArray interpreta el interior de un literal de arreglo de texto de
// PostgreSQL: elementos separados por coma, opcionalmente entre comillas dobles
// con escapes de barra invertida.
func parsePGTextArray(inner string) []string {
	out := []string{}
	var cur strings.Builder
	inQuotes, quoted, escaped := false, false, false
	flush := func() {
		v := cur.String()
		if quoted || strings.TrimSpace(v) != "" {
			if !quoted {
				v = strings.TrimSpace(v)
			}
			out = append(out, v)
		}
		cur.Reset()
		quoted = false
	}
	for _, r := range inner {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case inQuotes && r == '\\':
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
			quoted = true
		case r == ',' && !inQuotes:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
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

	dias := o.DiasSemanaArray
	if dias == nil {
		dias = []string{}
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
		PkIdRestaurante *Restaurante  `json:"restauranteId"`
	}{
		PkIdOferta:      o.PkIdOferta,
		Titulo:          o.Titulo,
		TipoDescuento:   o.TipoDescuento,
		ValorDescuento:  o.ValorDescuento,
		FechaInicio:     fechaInicioStr,
		FechaFin:        fechaFinStr,
		DiasSemana:      o.DiasSemana,
		DiasSemanaArray: dias,
		HoraInicio:      horaInicioStr,
		HoraFin:         horaFinStr,
		Activo:          o.Activo,
		PkIdRestaurante: o.PkIdRestaurante,
	})
}
