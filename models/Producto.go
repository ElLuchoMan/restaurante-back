package models

import (
	"encoding/json"
	"os"

	"github.com/beego/beego/v2/client/orm"
)

type Producto struct {
	PK_ID_PRODUCTO     int64          `orm:"column(pk_id_producto);pk;auto" json:"productoId"`
	NOMBRE             string         `orm:"column(nombre);type(text)" json:"nombre"`
	CALORIAS           *int64         `orm:"column(calorias);type(bigint)" json:"calorias"`
	DESCRIPCION        *string        `orm:"column(descripcion);type(text);null" json:"descripcion,omitempty"`
	PRECIO             int64          `orm:"column(precio);type(bigint)" json:"precio"`
	ESTADO_PRODUCTO    EstadoProducto `orm:"column(estado_producto);type(text)" json:"estadoProducto"`
	IMAGEN             []byte         `orm:"column(imagen);type(bytea);null" json:"imagen,omitempty"`
	CANTIDAD           int            `orm:"column(cantidad);type(integer)" json:"cantidad"`
	PK_ID_SUBCATEGORIA int64          `orm:"column(pk_id_subcategoria)" json:"subcategoriaId"`
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
	IMAGEN           []byte         `json:"imagen,omitempty"`
	CANTIDAD         int            `json:"cantidad"`
	PKIDSubcategoria int64          `json:"subcategoriaId"`
}

func (p Producto) MarshalJSON() ([]byte, error) {
	pj := productoJSON{
		PKIDProducto:     p.PK_ID_PRODUCTO,
		NOMBRE:           p.NOMBRE,
		CALORIAS:         p.CALORIAS,
		DESCRIPCION:      p.DESCRIPCION,
		PRECIO:           p.PRECIO,
		ESTADO_PRODUCTO:  p.ESTADO_PRODUCTO,
		IMAGEN:           p.IMAGEN,
		CANTIDAD:         p.CANTIDAD,
		PKIDSubcategoria: p.PK_ID_SUBCATEGORIA,
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
	p.IMAGEN = pj.IMAGEN
	p.CANTIDAD = pj.CANTIDAD
	p.PK_ID_SUBCATEGORIA = pj.PKIDSubcategoria
	return nil
}

func init() {
	if os.Getenv("ENABLE_PRODUCTO_MODEL") == "1" {
		orm.RegisterModel(new(Producto))
	}
}
