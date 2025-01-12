package models

import "github.com/beego/beego/v2/client/orm"

type Incidencia struct {
	PK_ID_INCIDENCIA        int64  `orm:"column(PK_ID_INCIDENCIA);pk;auto" json:"PK_ID_INCIDENCIA"`
	FECHA                   string `orm:"column(FECHA);type(date)" json:"FECHA"`
	MONTO                   int64  `orm:"column(MONTO)" json:"MONTO"`
	RESTA                   bool   `orm:"column(RESTA);type(boolean)" json:"RESTA"`
	MOTIVO                  string `orm:"column(MOTIVO);type(text)" json:"MOTIVO"`
	PK_DOCUMENTO_TRABAJADOR *int64 `orm:"column(PK_DOCUMENTO_TRABAJADOR);null" json:"PK_DOCUMENTO_TRABAJADOR,omitempty"`
}

func (i *Incidencia) TableName() string {
	return "INCIDENCIA"
}

func init() {
	orm.RegisterModel(new(Incidencia))
}
