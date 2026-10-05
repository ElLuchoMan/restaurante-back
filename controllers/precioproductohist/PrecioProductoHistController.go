package precioproductohist

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// PrecioProductoHistController expone el historial de precios (solo lectura).
type PrecioProductoHistController struct{ web.Controller }

const selectHist = `
SELECT pph.pk_id_precio_hist, pph.pk_id_producto, pr.nombre, pr.estado_producto, pph.precio, pph.fecha_vigencia
FROM precio_producto_hist pph
JOIN producto pr ON pr.pk_id_producto = pph.pk_id_producto`

// histRow es una fila del historial junto con los datos del producto.
type histRow struct {
	ID       int64     `orm:"column(pk_id_precio_hist)"`
	Producto int64     `orm:"column(pk_id_producto)"`
	Nombre   string    `orm:"column(nombre)"`
	Estado   string    `orm:"column(estado_producto)"`
	Precio   int64     `orm:"column(precio)"`
	Fecha    time.Time `orm:"column(fecha_vigencia)"`
}

// histItem es la forma JSON de cada elemento (ver models.PrecioHistItemDoc).
type histItem struct {
	PrecioHistID   int64  `json:"precioHistId"`
	ProductoID     int64  `json:"productoId"`
	Nombre         string `json:"nombre"`
	EstadoProducto string `json:"estadoProducto"`
	Precio         int64  `json:"precio"`
	FechaVigencia  string `json:"fechaVigencia"`
}

func (r histRow) item() histItem {
	return histItem{
		PrecioHistID:   r.ID,
		ProductoID:     r.Producto,
		Nombre:         r.Nombre,
		EstadoProducto: r.Estado,
		Precio:         r.Precio,
		FechaVigencia:  models.FormatDateUTC(r.Fecha),
	}
}

// @Title GetAll
// @Summary Listar historial de precios
// @Description Lista el historial ordenado por fechaVigencia ascendente. Filtros opcionales: `producto_id` (entero positivo) y `fecha` (YYYY-MM-DD). Cada elemento trae `precioHistId`, `productoId`, `nombre`, `estadoProducto`, `precio` y `fechaVigencia` (DD-MM-YYYY). Sin resultados, `data` es `[]`.
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param producto_id query int false "ID del producto (entero positivo)"
// @Param fecha query string false "Fecha de vigencia (YYYY-MM-DD)"
// @Success 200 {object} models.ApiResponse{data=[]models.PrecioHistItemDoc} "Historial de precios (puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "producto_id o fecha inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /precio_producto_hist [get]
func (c *PrecioProductoHistController) GetAll() {
	query := selectHist + "\nWHERE 1=1"
	args := []interface{}{}

	if strings.TrimSpace(c.GetString("producto_id")) != "" {
		pid, err := httpx.PositiveInt64Param(&c.Controller, "producto_id")
		if err != nil {
			logging.LogControllerError(c.Ctx, "precio_hist.getall.bad_producto", err, map[string]interface{}{"producto_id": c.GetString("producto_id")})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'producto_id' es inválido", err)
			return
		}
		query += " AND pph.pk_id_producto = ?"
		args = append(args, pid)
	}
	if f := strings.TrimSpace(c.GetString("fecha")); f != "" {
		d, err := models.ParseDateToNoonUTC(f)
		if err != nil {
			logging.LogControllerError(c.Ctx, "precio_hist.getall.bad_fecha", err, map[string]interface{}{"fecha": f})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'fecha' debe tener formato YYYY-MM-DD", err)
			return
		}
		query += " AND pph.fecha_vigencia = ?"
		args = append(args, d)
	}
	query += " ORDER BY pph.fecha_vigencia ASC, pph.pk_id_precio_hist ASC"

	var rows []histRow
	if _, err := orm.NewOrm().Raw(query, args...).QueryRows(&rows); err != nil {
		logging.LogControllerError(c.Ctx, "precio_hist.getall.db_error", err, map[string]interface{}{"producto_id": c.GetString("producto_id"), "fecha": c.GetString("fecha")})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener historial de precios", err)
		return
	}

	items := make([]histItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, r.item())
	}
	httpx.Send(&c.Controller, http.StatusOK, "Historial de precios", items)
}

// @Title GetById
// @Summary Obtener historial por ID
// @Description Devuelve `precioHistId`, `productoId`, `nombre`, `estadoProducto`, `precio` y `fechaVigencia` (DD-MM-YYYY) del registro indicado.
// @Tags precio_producto_hist
// @Accept json
// @Produce json
// @Param id query int true "ID del historial (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.PrecioHistItemDoc} "Historial encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Historial no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /precio_producto_hist/search [get]
func (c *PrecioProductoHistController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "precio_hist.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	var row histRow
	if err := orm.NewOrm().Raw(selectHist+"\nWHERE pph.pk_id_precio_hist = ?", id).QueryRow(&row); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Historial no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "precio_hist.getbyid.db_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener el historial de precios", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Historial encontrado", row.item())
}
