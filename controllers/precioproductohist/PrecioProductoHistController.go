package precioproductohist

import (
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type pphOrmer interface {
	Raw(string, ...interface{}) orm.RawSeter
}

var pphOrmNew = func() pphOrmer { return orm.NewOrm() }

type PrecioProductoHistController struct{ web.Controller }

// @Title GetAll
// @Summary Listar historial de precios
// @Description Opcional: filtrar por producto y/o fecha_vigencia (YYYY-MM-DD). Devuelve nombre, estadoProducto, precio y fechaVigencia.
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param producto_id query int false "ID del producto"
// @Param fecha query string false "Fecha de vigencia (YYYY-MM-DD)"
// @Success 200 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /precio_producto_hist [get]
func (c *PrecioProductoHistController) GetAll() {
	o := pphOrmNew()

	query := `
SELECT pr.nombre, pr.estado_producto, pph.precio, pph.fecha_vigencia
FROM precio_producto_hist pph
JOIN producto pr ON pr.pk_id_producto = pph.pk_id_producto
WHERE 1=1`
	args := []interface{}{}

	if pid, err := c.GetInt64("producto_id"); err == nil && pid > 0 {
		query += " AND pph.pk_id_producto = ?"
		args = append(args, pid)
	}
	if f := c.GetString("fecha"); f != "" {
		if d, err := time.Parse("2006-01-02", f); err == nil {
			query += " AND pph.fecha_vigencia = ?"
			args = append(args, d)
		} else {
			logging.LogControllerError(c.Ctx, "precio_hist.getall.bad_fecha", err, map[string]interface{}{"fecha": f})
		}
	}
	query += " ORDER BY pph.fecha_vigencia ASC"

	var rows []struct {
		Nombre string    `orm:"column(nombre)"`
		Estado string    `orm:"column(estado_producto)"`
		Precio int64     `orm:"column(precio)"`
		Fecha  time.Time `orm:"column(fecha_vigencia)"`
	}
	if _, err := o.Raw(query, args...).QueryRows(&rows); err != nil {
		logging.LogControllerError(c.Ctx, "precio_hist.getall.db_error", err, map[string]interface{}{"producto_id": c.GetString("producto_id"), "fecha": c.GetString("fecha")})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener historial de precios", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	resp := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, map[string]interface{}{
			"nombre":         r.Nombre,
			"estadoProducto": r.Estado,
			"precio":         r.Precio,
			"fechaVigencia":  r.Fecha,
		})
	}

	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Historial de precios", Data: resp}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener historial por ID
// @Description Devuelve nombre, estadoProducto, precio y fechaVigencia para el registro indicado.
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param id query int true "ID del historial"
// @Success 200 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /precio_producto_hist/search [get]
func (c *PrecioProductoHistController) GetById() {
	o := pphOrmNew()
	id, _ := c.GetInt64("id")
	query := `
SELECT pr.nombre, pr.estado_producto, pph.precio, pph.fecha_vigencia
FROM precio_producto_hist pph
JOIN producto pr ON pr.pk_id_producto = pph.pk_id_producto
WHERE pph.pk_id_precio_hist = ?`
	var row struct {
		Nombre string    `orm:"column(nombre)"`
		Estado string    `orm:"column(estado_producto)"`
		Precio int64     `orm:"column(precio)"`
		Fecha  time.Time `orm:"column(fecha_vigencia)"`
	}
	if err := o.Raw(query, id).QueryRow(&row); err != nil {
		if err != orm.ErrNoRows {
			logging.LogControllerError(c.Ctx, "precio_hist.getbyid.db_error", err, map[string]interface{}{"id": id})
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Historial no encontrado"}
		_ = c.ServeJSON()
		return
	}
	resp := map[string]interface{}{
		"nombre":         row.Nombre,
		"estadoProducto": row.Estado,
		"precio":         row.Precio,
		"fechaVigencia":  row.Fecha,
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Historial encontrado", Data: resp}
	_ = c.ServeJSON()
}
