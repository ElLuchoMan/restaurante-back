package models

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int     `orm:"pk" json:"PK_DOCUMENTO_CLIENTE"`
	NOMBRE               string  `json:"NOMBRE"`
	APELLIDO             string  `json:"APELLIDO"`
	DIRECCION            string  `json:"DIRECCION"`
	TELEFONO             string  `json:"TELEFONO"`
	OBSERVACIONES        *string `json:"OBSERVACIONES"`
	PASSWORD             string  `json:"PASSWORD"`
}
