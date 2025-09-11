package productopedido

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"sort"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type productoPedidoOrmer interface {
	QueryTable(interface{}) orm.QuerySeter
	Insert(interface{}) (int64, error)
}

// productoPedidoNewOrm allows tests to stub orm.NewOrm.
var productoPedidoNewOrm = func() productoPedidoOrmer { return orm.NewOrm() }

// Hooks para pruebas (permite stubear Begin y el Ormer base en tests)
var productoPedidoBaseOrmNew = func() orm.Ormer { return orm.NewOrm() }
var productoPedidoBeginTx = func(o orm.Ormer) (orm.TxOrmer, error) { return o.Begin() }

// Hooks adicionales para facilitar pruebas del camino feliz en Update
var productoPedidoDeleteDetalles = func(tx orm.TxOrmer, pedidoID int64) error {
	_, err := tx.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).Delete()
	return err
}

func productoPedidoRequeryDetalleDefault(tx orm.TxOrmer, pedidoID int64, productoID int64, out *models.DetallePedido) error {
	return tx.QueryTable(new(models.DetallePedido)).
		Filter("PKIDPedido", pedidoID).
		Filter("PKIDProducto", productoID).
		One(out)
}

var productoPedidoRequeryDetalle = productoPedidoRequeryDetalleDefault

// productoPedidoComputeDeltas calcula deltas y necesidades de stock
func productoPedidoComputeDeltas(actuales []models.DetallePedido, nuevos map[int64]int) (map[int64]int, map[int64]int) {
	actualMap := make(map[int64]int)
	for _, a := range actuales {
		if a.PKIDProducto != nil {
			actualMap[a.PKIDProducto.PK_ID_PRODUCTO] += a.Cantidad
		}
	}
	deltas := make(map[int64]int)
	for pid, qty := range nuevos {
		prev := actualMap[pid]
		deltas[pid] = qty - prev
		delete(actualMap, pid)
	}
	for pid, prev := range actualMap {
		deltas[pid] = -prev
	}
	need := make(map[int64]int)
	for pid, d := range deltas {
		if d > 0 {
			need[pid] = d
		}
	}
	return deltas, need
}

type ProductoPedidoController struct {
	web.Controller
}

// Estructura para mapear las respuestas en camelCase
type ProductoPedidoResponse struct {
	PedidoID int64       `json:"pedidoId"`
	Detalles interface{} `json:"detalles"`
}

// detallePedidoInput represents the payload for DetallePedido without precio.
type detallePedidoInput struct {
	PKIDProducto int64 `json:"productoId"`
	Cantidad     int   `json:"cantidad"`
}

// @Title GetAll
// @Summary Obtener los productos de un pedido
// @Description Devuelve los productos consolidados en un pedido específico
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Success 200 {object} models.ApiResponse{data=productopedido.ProductoPedidoResponse} "Lista de productos del pedido"
// @Failure 404 {object} models.ApiResponse "No se encontraron productos asociados a este pedido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /producto_pedido [get]
func (c *ProductoPedidoController) GetAll() {
	pedidoID, err := c.GetInt64("pedido_id")
	if err != nil || pedidoID == 0 {
		logging.LogControllerError(c.Ctx, "producto_pedido.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'pedido_id' es obligatorio y debe ser válido"}
		_ = c.ServeJSON()
		return
	}

	o := productoPedidoNewOrm()
	var detalles []models.DetallePedido
	if _, err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).All(&detalles); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.getall.db_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener los productos del pedido", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if len(detalles) == 0 {
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "No se encontraron productos asociados a este pedido"}
		_ = c.ServeJSON()
		return
	}

	response := map[string]interface{}{"pedidoId": pedidoID, "detalles": detalles}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Productos del pedido obtenidos exitosamente", Data: response}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear un pedido con productos consolidados
// @Description Crea un registro de productos consolidados en un pedido
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param body body models.ProductoPedidoCreateRequest true "Datos del pedido con productos"
// @Success 201 {object} models.ApiResponse{data=productopedido.ProductoPedidoResponse} "Pedido con productos agregado exitosamente"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [post]
func (c *ProductoPedidoController) Post() {
	var input struct {
		PedidoId int64                `json:"pedidoId"`
		Detalles []detallePedidoInput `json:"detalles"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Datos inválidos", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	if input.PedidoId == 0 || len(input.Detalles) == 0 {
		logging.LogControllerError(c.Ctx, "producto_pedido.post.validation_error", nil, map[string]interface{}{"pedidoId": input.PedidoId, "detalles_len": len(input.Detalles), "body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El pedido y los detalles de los productos son obligatorios"}
		_ = c.ServeJSON()
		return
	}

	nuevos := make(map[int64]int)
	for _, d := range input.Detalles {
		if d.PKIDProducto == 0 || d.Cantidad <= 0 {
			continue
		}
		nuevos[d.PKIDProducto] += d.Cantidad
	}

	o := productoPedidoBaseOrmNew()
	if len(nuevos) > 0 {
		ids := make([]int64, 0, len(nuevos))
		for pid := range nuevos {
			ids = append(ids, pid)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		ph := make([]string, len(ids))
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			ph[i] = "?"
			args[i] = id
		}
		query := fmt.Sprintf("SELECT pk_id_producto, cantidad FROM producto WHERE pk_id_producto IN (%s)", strings.Join(ph, ","))
		var rows []struct {
			PK       int64 `orm:"column(pk_id_producto)"`
			Cantidad int   `orm:"column(cantidad)"`
		}
		if _, err := o.Raw(query, args...).QueryRows(&rows); err != nil {
			// continuar, comportamiento esperado
		}
		avail := make(map[int64]int)
		for _, r := range rows {
			avail[r.PK] = r.Cantidad
		}
		var insuf []map[string]interface{}
		for pid, req := range nuevos {
			if avail[pid] < req {
				insuf = append(insuf, map[string]interface{}{"productoId": pid, "requerido": req, "disponible": avail[pid]})
			}
		}
		if len(insuf) > 0 {
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Inventario insuficiente para uno o más productos", Data: insuf}
			_ = c.ServeJSON()
			return
		}
	}

	tx, err := productoPedidoBeginTx(o)
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.post.tx_begin_error", err, map[string]interface{}{"pedidoId": input.PedidoId, "body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible iniciar transacción", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Bloquear la fila del pedido para evitar actualizaciones concurrentes sobre el mismo pedido
	var lockDummy int
	if err := tx.Raw("SELECT 1 FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", input.PedidoId).QueryRow(&lockDummy); err != nil {
		_ = tx.Rollback()
		logging.LogControllerError(c.Ctx, "producto_pedido.post.lock_pedido_error", err, map[string]interface{}{"pedidoId": input.PedidoId})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible bloquear el pedido para actualización", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	var detalles []models.DetallePedido
	// Aplicar actualizaciones en orden determinístico para minimizar deadlocks
	orderedIDs := make([]int64, 0, len(nuevos))
	for pid := range nuevos {
		orderedIDs = append(orderedIDs, pid)
	}
	sort.Slice(orderedIDs, func(i, j int) bool { return orderedIDs[i] < orderedIDs[j] })
	for _, prodID := range orderedIDs {
		qty := nuevos[prodID]
		res, err := tx.Raw("UPDATE producto SET cantidad = cantidad - ? WHERE pk_id_producto = ? AND cantidad >= ?", qty, prodID, qty).Exec()
		if err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.post.stock_update_error", err, map[string]interface{}{"productoId": prodID, "qty": qty, "body": string(c.Ctx.Input.RequestBody)})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al descontar inventario", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		affected, err := res.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.post.rows_affected_error", err, map[string]interface{}{"productoId": prodID})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al verificar actualización de inventario", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		if affected == 0 {
			_ = tx.Rollback()
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Inventario insuficiente para uno o más productos"}
			_ = c.ServeJSON()
			return
		}

		detalle := models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: input.PedidoId}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: prodID}, Cantidad: qty}
		if _, err := tx.Insert(&detalle); err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.post.insert_detalle_error", err, map[string]interface{}{"productoId": prodID, "qty": qty})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el pedido con productos", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		var actualizado models.DetallePedido
		if err := tx.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", *detalle.PKIDPedido).Filter("PKIDProducto", *detalle.PKIDProducto).One(&actualizado); err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.post.requery_error", err, map[string]interface{}{"productoId": prodID})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener el precio del producto", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		detalles = append(detalles, actualizado)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logging.LogControllerError(c.Ctx, "producto_pedido.post.tx_commit_error", err, map[string]interface{}{"pedidoId": input.PedidoId})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible confirmar transacción", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	response := map[string]interface{}{"pedidoId": input.PedidoId, "detalles": detalles}
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Pedido con productos agregado exitosamente", Data: response}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar productos en un pedido consolidado
// @Description Permite agregar o modificar productos en un pedido consolidado
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido a actualizar"
// @Param body body models.ProductoPedidoUpdateRequest true "Lista actualizada de productos"
// @Success 200 {object} models.ApiResponse{data=productopedido.ProductoPedidoResponse} "Productos actualizados exitosamente"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [put]
func (c *ProductoPedidoController) Update() {
	pedidoID, err := c.GetInt64("pedido_id")
	if err != nil || pedidoID == 0 {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'pedido_id' es obligatorio y debe ser válido"}
		_ = c.ServeJSON()
		return
	}
	var nuevosProductos []detallePedidoInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &nuevosProductos); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.bad_json", err, map[string]interface{}{"pedido_id": pedidoID, "body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Datos inválidos", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	if len(nuevosProductos) == 0 {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.validation_error", nil, map[string]interface{}{"pedido_id": pedidoID, "len": 0, "body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "La lista de productos no puede estar vacía"}
		_ = c.ServeJSON()
		return
	}

	nuevos := make(map[int64]int)
	for _, d := range nuevosProductos {
		if d.PKIDProducto == 0 || d.Cantidad < 0 {
			continue
		}
		nuevos[d.PKIDProducto] += d.Cantidad
	}
	o := productoPedidoBaseOrmNew()
	var actuales []models.DetallePedido
	if _, err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).All(&actuales); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.query_actuales_error", err, map[string]interface{}{"pedido_id": pedidoID, "body": string(c.Ctx.Input.RequestBody)})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al buscar los detalles del pedido", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	deltas, need := productoPedidoComputeDeltas(actuales, nuevos)
	if len(need) > 0 {
		ids := make([]int64, 0, len(need))
		for pid := range need {
			ids = append(ids, pid)
		}
		ph := make([]string, len(ids))
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			ph[i] = "?"
			args[i] = id
		}
		query := fmt.Sprintf("SELECT pk_id_producto, cantidad FROM producto WHERE pk_id_producto IN (%s)", strings.Join(ph, ","))
		var rows []struct {
			PK       int64 `orm:"column(pk_id_producto)"`
			Cantidad int   `orm:"column(cantidad)"`
		}
		if _, err := o.Raw(query, args...).QueryRows(&rows); err != nil {
			logging.LogControllerError(c.Ctx, "producto_pedido.update.validar_inventario_error", err, map[string]interface{}{"pedido_id": pedidoID})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al validar inventario", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		avail := make(map[int64]int)
		for _, r := range rows {
			avail[r.PK] = r.Cantidad
		}
		var insuf []map[string]interface{}
		for pid, req := range need {
			disp := avail[pid]
			if disp < req {
				insuf = append(insuf, map[string]interface{}{"productoId": pid, "requerido": req, "disponible": disp})
			}
		}
		if len(insuf) > 0 {
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Inventario insuficiente para uno o más productos", Data: insuf}
			_ = c.ServeJSON()
			return
		}
	}
	tx, err := productoPedidoBeginTx(o)
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.tx_begin_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible iniciar transacción", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Bloquear la fila del pedido para evitar actualizaciones concurrentes del mismo pedido
	var lockDummy2 int
	if err := tx.Raw("SELECT 1 FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", pedidoID).QueryRow(&lockDummy2); err != nil {
		_ = tx.Rollback()
		logging.LogControllerError(c.Ctx, "producto_pedido.update.lock_pedido_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible bloquear el pedido para actualización", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	// Iterar deltas en orden determinístico para minimizar deadlocks
	deltaIDs := make([]int64, 0, len(deltas))
	for pid := range deltas {
		deltaIDs = append(deltaIDs, pid)
	}
	sort.Slice(deltaIDs, func(i, j int) bool { return deltaIDs[i] < deltaIDs[j] })
	for _, pid := range deltaIDs {
		delta := deltas[pid]
		if delta > 0 {
			res, err := tx.Raw("UPDATE producto SET cantidad = cantidad - ? WHERE pk_id_producto = ? AND cantidad >= ?", delta, pid, delta).Exec()
			if err != nil {
				_ = tx.Rollback()
				logging.LogControllerError(c.Ctx, "producto_pedido.update.stock_update_error", err, map[string]interface{}{"productoId": pid, "delta": delta, "body": string(c.Ctx.Input.RequestBody)})
				c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al descontar inventario", Cause: err.Error()}
				_ = c.ServeJSON()
				return
			}
			affected, err := res.RowsAffected()
			if err != nil {
				_ = tx.Rollback()
				logging.LogControllerError(c.Ctx, "producto_pedido.update.rows_affected_error", err, map[string]interface{}{"productoId": pid})
				c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al verificar actualización de inventario", Cause: err.Error()}
				_ = c.ServeJSON()
				return
			}
			if affected == 0 {
				_ = tx.Rollback()
				c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Inventario insuficiente para uno o más productos"}
				_ = c.ServeJSON()
				return
			}
		} else if delta < 0 {
			inc := -delta
			if _, err := tx.Raw("UPDATE producto SET cantidad = cantidad + ? WHERE pk_id_producto = ?", inc, pid).Exec(); err != nil {
				_ = tx.Rollback()
				logging.LogControllerError(c.Ctx, "producto_pedido.update.stock_restore_error", err, map[string]interface{}{"productoId": pid, "inc": inc, "body": string(c.Ctx.Input.RequestBody)})
				c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al ajustar inventario", Cause: err.Error()}
				_ = c.ServeJSON()
				return
			}
		}
	}
	if err := productoPedidoDeleteDetalles(tx, pedidoID); err != nil {
		_ = tx.Rollback()
		logging.LogControllerError(c.Ctx, "producto_pedido.update.delete_detalles_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar los productos del pedido", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	var detalles []models.DetallePedido
	// Insertar detalles en orden determinístico
	newIDs := make([]int64, 0, len(nuevos))
	for pid := range nuevos {
		newIDs = append(newIDs, pid)
	}
	sort.Slice(newIDs, func(i, j int) bool { return newIDs[i] < newIDs[j] })
	for _, pid := range newIDs {
		qty := nuevos[pid]
		detalle := models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedidoID}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: pid}, Cantidad: qty}
		if _, err := tx.Insert(&detalle); err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.update.insert_detalle_error", err, map[string]interface{}{"productoId": pid, "qty": qty, "body": string(c.Ctx.Input.RequestBody)})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar los productos del pedido", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		var actualizado models.DetallePedido
		if err := productoPedidoRequeryDetalle(tx, detalle.PKIDPedido.PK_ID_PEDIDO, detalle.PKIDProducto.PK_ID_PRODUCTO, &actualizado); err != nil {
			_ = tx.Rollback()
			logging.LogControllerError(c.Ctx, "producto_pedido.update.requery_error", err, map[string]interface{}{"productoId": pid})
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener el precio del producto", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		detalles = append(detalles, actualizado)
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logging.LogControllerError(c.Ctx, "producto_pedido.update.tx_commit_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No fue posible confirmar transacción", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	response := map[string]interface{}{"pedidoId": pedidoID, "detalles": detalles}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Productos del pedido actualizados exitosamente", Data: response}
	_ = c.ServeJSON()
}
