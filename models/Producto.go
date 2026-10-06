package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	PK_ID_SUBCATEGORIA *Subcategoria  `orm:"column(pk_id_subcategoria);rel(fk);null" json:"subcategoriaId"`
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
	PKIDSubcategoria *int64         `json:"subcategoriaId"`
}

// DecodeImagenBase64 decodifica una imagen en Base64. Tolera el prefijo de
// data URI ("data:image/png;base64,"), espacios/saltos de línea y la falta de
// relleno ("=").
func DecodeImagenBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "data:") {
		i := strings.Index(s, ",")
		if i < 0 {
			return nil, fmt.Errorf("data URI de imagen inválido: falta la coma separadora")
		}
		s = s[i+1:]
	}
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, s)
	s = strings.TrimRight(s, "=")
	return base64.RawStdEncoding.DecodeString(s)
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
		PKIDSubcategoria: func() *int64 {
			if p.PK_ID_SUBCATEGORIA != nil {
				id := p.PK_ID_SUBCATEGORIA.PK_ID_SUBCATEGORIA
				return &id
			}
			return nil
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
	p.IMAGEN = ""
	if pj.IMAGEN != "" {
		img, err := DecodeImagenBase64(pj.IMAGEN)
		if err != nil {
			return err
		}
		p.IMAGEN = string(img)
	}
	p.CANTIDAD = pj.CANTIDAD
	p.PK_ID_SUBCATEGORIA = nil
	if pj.PKIDSubcategoria != nil && *pj.PKIDSubcategoria != 0 {
		p.PK_ID_SUBCATEGORIA = &Subcategoria{PK_ID_SUBCATEGORIA: *pj.PKIDSubcategoria}
	}
	return nil
}

func init() {
	if os.Getenv("SKIP_ORM_REGISTRATION") != "1" {
		orm.RegisterModel(new(Producto))
	}
}
