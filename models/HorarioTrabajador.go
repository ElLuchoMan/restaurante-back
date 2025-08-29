package models

import "github.com/beego/beego/v2/client/orm"

// HorarioTrabajador almacena los turnos asignados a un trabajador.
type HorarioTrabajador struct {
	PKIDHorarioTrabajador int64  `orm:"column(PK_ID_HORARIO_TRABAJADOR);pk;auto" json:"horarioTrabajadorId"`
	DocumentoTrabajador   int64  `orm:"column(PK_DOCUMENTO_TRABAJADOR)" json:"documentoTrabajador"`
	Dia                   string `orm:"column(DIA);size(20)" json:"dia"`
	HoraInicio            string `orm:"column(HORA_INICIO);type(time)" json:"horaInicio"`
	HoraFin               string `orm:"column(HORA_FIN);type(time)" json:"horaFin"`
}

// TableName especifica la tabla asociada al modelo.
func (h *HorarioTrabajador) TableName() string {
	return "HORARIO_TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(HorarioTrabajador))
}
