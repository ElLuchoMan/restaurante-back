package models

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/beego/beego/v2/client/orm"
)

type Producto struct {
	PK_ID_PRODUCTO     int64          `orm:"column(pk_id_producto);pk;auto" json:"productoId"`
	NOMBRE             string         `orm:"column(nombre);type(text)" json:"nombre"`
	CALORIAS           *int64         `orm:"column(calorias);type(bigint);null" json:"calorias"`
	DESCRIPCION        *string        `orm:"column(descripcion);type(text);null" json:"descripcion,omitempty"`
	PRECIO             int64          `orm:"column(precio);type(bigint)" json:"precio"`
	ESTADO_PRODUCTO    EstadoProducto `orm:"column(estado_producto);type(estado_producto)" json:"estadoProducto"`
	IMAGEN             string         `orm:"column(imagen);type(bytea);null" json:"imagen"`
	CANTIDAD           int            `orm:"column(cantidad);type(integer)" json:"cantidad"`
	PK_ID_SUBCATEGORIA *Subcategoria  `orm:"column(pk_id_subcategoria);rel(fk)" json:"subcategoriaId"`
}

func (p *Producto) TableName() string {
	return "producto"
}

type productoJSON struct {
	PKIDProducto     int64          `json:"productoId"`
	NOMBRE           string         `json:"nombre"`
	CALORIAS         *int64         `json:"calorias"`
	DESCRIPCION      *string        `json:"descripcion,omitempty"`
	PRECIO           int64          `json:"precio"`
	ESTADO_PRODUCTO  EstadoProducto `json:"estadoProducto"`
	IMAGEN           string         `json:"imagen,omitempty"`
	CANTIDAD         int            `json:"cantidad"`
	PKIDSubcategoria int64          `json:"subcategoriaId"`
}

func (p Producto) MarshalJSON() ([]byte, error) {
	pj := productoJSON{
		PKIDProducto:    p.PK_ID_PRODUCTO,
		NOMBRE:          p.NOMBRE,
		CALORIAS:        p.CALORIAS,
		DESCRIPCION:     p.DESCRIPCION,
		PRECIO:          p.PRECIO,
		ESTADO_PRODUCTO: p.ESTADO_PRODUCTO,
		IMAGEN:          base64.StdEncoding.EncodeToString([]byte(p.IMAGEN)),
		CANTIDAD:        p.CANTIDAD,
		PKIDSubcategoria: func() int64 {
			if p.PK_ID_SUBCATEGORIA != nil {
				return p.PK_ID_SUBCATEGORIA.PK_ID_SUBCATEGORIA
			}
			return 0
		}(),
	}
	return json.Marshal(pj)
}

func (p *Producto) UnmarshalJSON(data []byte) error {
	var pj productoJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return err
	}
	p.PK_ID_PRODUCTO = pj.PKIDProducto
	p.NOMBRE = pj.NOMBRE
	p.CALORIAS = pj.CALORIAS
	p.DESCRIPCION = pj.DESCRIPCION
	p.PRECIO = pj.PRECIO
	p.ESTADO_PRODUCTO = pj.ESTADO_PRODUCTO
	if pj.IMAGEN != "" {
		if img, err := base64.StdEncoding.DecodeString(pj.IMAGEN); err == nil {
			p.IMAGEN = string(img)
		} else {
			return err
		}
	} else {
		p.IMAGEN = ""
	}
	p.CANTIDAD = pj.CANTIDAD
	if pj.PKIDSubcategoria != 0 {
		p.PK_ID_SUBCATEGORIA = &Subcategoria{PK_ID_SUBCATEGORIA: pj.PKIDSubcategoria}
	} else {
		p.PK_ID_SUBCATEGORIA = nil
	}
	return nil
}

func init() {
	if os.Getenv("SKIP_ORM_REGISTRATION") != "1" {
		orm.RegisterModel(new(Producto))
	}
}
