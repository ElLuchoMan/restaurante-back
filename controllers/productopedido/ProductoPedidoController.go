package productopedido

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ProductoPedidoController struct {
	web.Controller
}

// productoPedidoData es el `data` de GET/POST/PUT /producto_pedido
// (documentado como models.ProductoPedidoDoc).
type productoPedidoData struct {
	PedidoID int64                  `json:"pedidoId"`
	Detalles []models.DetallePedido `json:"detalles"`
}

const (
	msgPedidoNoEncontrado = "Pedido no encontrado"
	msgPedidoIDInvalido   = "El parámetro 'pedido_id' es obligatorio y debe ser un entero positivo"
	msgInventario         = "Inventario insuficiente para uno o más productos"
)

func (c *ProductoPedidoController) fail(status int, msg string, err error) {
	httpx.Fail(&c.Controller, status, msg, err)
}

// failTx registra el error, deshace la transacción y responde 409 si es un
// conflicto de PostgreSQL (p. ej. el producto ya está en el pedido) o 500.
func (c *ProductoPedidoController) failTx(tx orm.TxOrmer, op, msg string, err error, ctx map[string]interface{}) {
	_ = tx.Rollback()
	logging.LogControllerError(c.Ctx, "producto_pedido."+op, err, ctx)
	if httpx.IsPGConflict(err) {
		c.fail(http.StatusConflict, msg+": conflicto con datos existentes", err)
		return
	}
	c.fail(http.StatusInternalServerError, msg, err)
}

// begin abre una transacción y bloquea la fila del pedido. Devuelve false si
// ya respondió (404 si el pedido no existe, 500 en otro error).
func (c *ProductoPedidoController) begin(op string, pedidoID int64) (orm.TxOrmer, bool) {
	ctx := map[string]interface{}{"pedido_id": pedidoID}
	tx, err := orm.NewOrm().Begin()
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido."+op+".tx_begin_error", err, ctx)
		c.fail(http.StatusInternalServerError, "No fue posible iniciar transacción", err)
		return nil, false
	}
	var lock int
	if err := tx.Raw("SELECT 1 FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", pedidoID).QueryRow(&lock); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			_ = tx.Rollback()
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return nil, false
		}
		c.failTx(tx, op+".lock_pedido_error", "No fue posible bloquear el pedido para actualización", err, ctx)
		return nil, false
	}
	return tx, true
}

func sortedIDs[V any](m map[int64]V) []int64 {
	ids := make([]int64, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// checkStock bloquea los productos de need (id -> unidades que faltan por
// descontar) y verifica que existan (404) y que alcance el inventario (409,
// con el detalle en `data`). Devuelve false si ya respondió.
func (c *ProductoPedidoController) checkStock(tx orm.TxOrmer, op string, need map[int64]int) bool {
	if len(need) == 0 {
		return true
	}
	ids := sortedIDs(need)
	ph := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		ph[i], args[i] = "?", id
	}
	query := fmt.Sprintf("SELECT pk_id_producto, cantidad FROM producto WHERE pk_id_producto IN (%s) ORDER BY pk_id_producto FOR UPDATE", strings.Join(ph, ","))
	var rows []struct {
		PK       int64 `orm:"column(pk_id_producto)"`
		Cantidad int   `orm:"column(cantidad)"`
	}
	if _, err := tx.Raw(query, args...).QueryRows(&rows); err != nil {
		c.failTx(tx, op+".validar_inventario_error", "Error al validar inventario", err, nil)
		return false
	}
	avail := make(map[int64]int, len(rows))
	for _, r := range rows {
		avail[r.PK] = r.Cantidad
	}
	var missing []string
	insuf := []models.InventarioInsuficienteDoc{}
	for _, id := range ids {
		disp, ok := avail[id]
		switch {
		case !ok:
			missing = append(missing, fmt.Sprint(id))
		case disp < need[id]:
			insuf = append(insuf, models.InventarioInsuficienteDoc{ProductoId: id, Requerido: need[id], Disponible: disp})
		}
	}
	if len(missing) > 0 {
		_ = tx.Rollback()
		c.fail(http.StatusNotFound, "Producto no encontrado: "+strings.Join(missing, ", "), nil)
		return false
	}
	if len(insuf) > 0 {
		_ = tx.Rollback()
		httpx.Send(&c.Controller, http.StatusConflict, msgInventario, insuf)
		return false
	}
	return true
}

// applyDeltas ajusta el inventario: delta > 0 descuenta unidades y delta < 0
// las devuelve. Devuelve false si ya respondió.
func (c *ProductoPedidoController) applyDeltas(tx orm.TxOrmer, op string, deltas map[int64]int) bool {
	for _, pid := range sortedIDs(deltas) {
		if deltas[pid] == 0 {
			continue
		}
		if _, err := tx.Raw("UPDATE producto SET cantidad = cantidad - ? WHERE pk_id_producto = ?", deltas[pid], pid).Exec(); err != nil {
			c.failTx(tx, op+".stock_update_error", "Error al ajustar inventario", err, map[string]interface{}{"productoId": pid, "delta": deltas[pid]})
			return false
		}
	}
	return true
}

// insertDetalles inserta una línea por producto y la relee para devolver el
// precio unitario que fija la base de datos. Devuelve false si ya respondió.
func (c *ProductoPedidoController) insertDetalles(tx orm.TxOrmer, op string, pedidoID int64, items map[int64]int) ([]models.DetallePedido, bool) {
	detalles := []models.DetallePedido{}
	for _, pid := range sortedIDs(items) {
		ctx := map[string]interface{}{"pedido_id": pedidoID, "productoId": pid, "cantidad": items[pid]}
		detalle := models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedidoID}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: pid}, Cantidad: items[pid]}
		if _, err := tx.Insert(&detalle); err != nil {
			c.failTx(tx, op+".insert_detalle_error", "Error al guardar los productos del pedido", err, ctx)
			return nil, false
		}
		var actualizado models.DetallePedido
		if err := tx.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).Filter("PKIDProducto", pid).One(&actualizado); err != nil {
			c.failTx(tx, op+".requery_error", "Error al obtener el precio del producto", err, ctx)
			return nil, false
		}
		detalles = append(detalles, actualizado)
	}
	return detalles, true
}

func (c *ProductoPedidoController) commit(tx orm.TxOrmer, op string, pedidoID int64) bool {
	if err := tx.Commit(); err != nil {
		c.failTx(tx, op+".tx_commit_error", "No fue posible confirmar transacción", err, map[string]interface{}{"pedido_id": pedidoID})
		return false
	}
	return true
}

// parseItems valida las líneas y las consolida por producto (las repetidas se
// suman). Un productoId <= 0 o una cantidad negativa es un error; las líneas
// con cantidad 0 se conservan en el mapa solo si keepZero (PUT: quitan el
// producto) y se ignoran en otro caso (POST).
func parseItems(items []models.ProductoPedidoItemInput, keepZero bool) (map[int64]int, error) {
	out := make(map[int64]int, len(items))
	for i, it := range items {
		if it.ProductoId <= 0 {
			return nil, fmt.Errorf("detalles[%d]: productoId debe ser un entero positivo", i)
		}
		if it.Cantidad < 0 {
			return nil, fmt.Errorf("detalles[%d]: cantidad no puede ser negativa", i)
		}
		if it.Cantidad > 0 || keepZero {
			out[it.ProductoId] += it.Cantidad
		}
	}
	return out, nil
}

// @Title GetAll
// @Summary Obtener los productos de un pedido
// @Description Devuelve las líneas (detalles) de un pedido. Un pedido existente sin productos responde 200 con `detalles` igual a `[]`; un pedido inexistente responde 404. En cada línea `pedidoId` y `productoId` son objetos de la relación (solo el id es fiable) y `precio` es el precio unitario.
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos del pedido (`detalles` puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "pedido_id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /producto_pedido [get]
func (c *ProductoPedidoController) GetAll() {
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.fail(http.StatusBadRequest, msgPedidoIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	var detalles []models.DetallePedido
	if _, err := o.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).OrderBy("PK_ID_DETALLE").All(&detalles); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.getall.db_error", err, map[string]interface{}{"pedido_id": pedidoID})
		c.fail(http.StatusInternalServerError, "Error al obtener los productos del pedido", err)
		return
	}
	if len(detalles) == 0 {
		n, err := o.QueryTable(new(models.Pedido)).Filter("PK_ID_PEDIDO", pedidoID).Count()
		if err != nil {
			logging.LogControllerError(c.Ctx, "producto_pedido.getall.pedido_error", err, map[string]interface{}{"pedido_id": pedidoID})
			c.fail(http.StatusInternalServerError, "Error al consultar el pedido", err)
			return
		}
		if n == 0 {
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return
		}
	}
	httpx.Send(&c.Controller, http.StatusOK, "Productos del pedido obtenidos exitosamente", productoPedidoData{PedidoID: pedidoID, Detalles: httpx.List(detalles)})
}

// @Title Post
// @Summary Agregar productos a un pedido
// @Description Agrega las líneas indicadas a un pedido existente y descuenta el inventario, todo en una transacción. Las líneas repetidas se suman y las de `cantidad` 0 se ignoran (debe quedar al menos una con `cantidad` > 0); `productoId` <= 0 o `cantidad` negativa responden 400. Responde 404 si el pedido o algún producto no existe, 409 si el inventario no alcanza (`data` lista {productoId, requerido, disponible}) o si el producto ya está en el pedido (use PUT para modificarlo).
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param body body models.ProductoPedidoCreateRequest true "Pedido y sus productos"
// @Success 201 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos agregados exitosamente"
// @Failure 400 {object} models.ApiResponse "JSON inválido, pedidoId inválido, sin líneas válidas, productoId <= 0 o cantidad negativa"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido o producto no encontrado"
// @Failure 409 {object} models.ApiResponse{data=[]models.InventarioInsuficienteDoc} "Inventario insuficiente (con detalle en `data`) o producto ya presente en el pedido"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [post]
func (c *ProductoPedidoController) Post() {
	var input models.ProductoPedidoCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Datos inválidos", err)
		return
	}
	if input.PedidoId <= 0 || len(input.Detalles) == 0 {
		c.fail(http.StatusBadRequest, "El pedido (entero positivo) y los detalles de los productos son obligatorios", nil)
		return
	}
	nuevos, err := parseItems(input.Detalles, false)
	if err != nil {
		c.fail(http.StatusBadRequest, "Datos inválidos", err)
		return
	}
	if len(nuevos) == 0 {
		c.fail(http.StatusBadRequest, "Debe haber al menos un producto con cantidad mayor que 0", nil)
		return
	}

	tx, ok := c.begin("post", input.PedidoId)
	if !ok || !c.checkStock(tx, "post", nuevos) || !c.applyDeltas(tx, "post", nuevos) {
		return
	}
	detalles, ok := c.insertDetalles(tx, "post", input.PedidoId, nuevos)
	if !ok || !c.commit(tx, "post", input.PedidoId) {
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Pedido con productos agregado exitosamente", productoPedidoData{PedidoID: input.PedidoId, Detalles: detalles})
}

// @Title Update
// @Summary Reemplazar los productos de un pedido
// @Description Reemplaza las líneas del pedido por la lista enviada (el cuerpo es un arreglo, no un objeto, así que no aplica el merge por campos): los productos que no aparezcan se quitan y una línea con `cantidad` 0 también quita el producto. El inventario se ajusta con la diferencia (descuenta o devuelve) en una transacción. Las líneas repetidas se suman; `productoId` <= 0 o `cantidad` negativa responden 400; la lista no puede estar vacía. Responde 404 si el pedido o algún producto a descontar no existe y 409 si el inventario no alcanza (`data` lista {productoId, requerido, disponible}).
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param body body models.ProductoPedidoUpdateRequest true "Lista completa de productos que debe quedar en el pedido"
// @Success 200 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos actualizados exitosamente"
// @Failure 400 {object} models.ApiResponse "pedido_id inválido, cuerpo que no es un arreglo, lista vacía, productoId <= 0 o cantidad negativa"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido o producto no encontrado"
// @Failure 409 {object} models.ApiResponse{data=[]models.InventarioInsuficienteDoc} "Inventario insuficiente (con detalle en `data`)"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [put]
func (c *ProductoPedidoController) Update() {
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.fail(http.StatusBadRequest, msgPedidoIDInvalido, err)
		return
	}
	var items models.ProductoPedidoUpdateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &items); err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.update.bad_json", err, map[string]interface{}{"pedido_id": pedidoID, "body": string(c.Ctx.Input.RequestBody)})
		c.fail(http.StatusBadRequest, "Datos inválidos", err)
		return
	}
	if len(items) == 0 {
		c.fail(http.StatusBadRequest, "La lista de productos no puede estar vacía", nil)
		return
	}
	todos, err := parseItems(items, true)
	if err != nil {
		c.fail(http.StatusBadRequest, "Datos inválidos", err)
		return
	}
	nuevos := make(map[int64]int, len(todos))
	for pid, qty := range todos {
		if qty > 0 {
			nuevos[pid] = qty
		}
	}

	tx, ok := c.begin("update", pedidoID)
	if !ok {
		return
	}
	var actuales []models.DetallePedido
	if _, err := tx.QueryTable(new(models.DetallePedido)).Filter("PKIDPedido", pedidoID).All(&actuales); err != nil {
		c.failTx(tx, "update.query_actuales_error", "Error al buscar los detalles del pedido", err, map[string]interface{}{"pedido_id": pedidoID})
		return
	}
	deltas := make(map[int64]int, len(nuevos)+len(actuales))
	for pid, qty := range nuevos {
		deltas[pid] = qty
	}
	for _, a := range actuales {
		deltas[a.PKIDProducto.PK_ID_PRODUCTO] -= a.Cantidad
	}
	need := make(map[int64]int)
	for pid, d := range deltas {
		if d > 0 {
			need[pid] = d
		}
	}
	if !c.checkStock(tx, "update", need) || !c.applyDeltas(tx, "update", deltas) {
		return
	}
	if _, err := tx.Raw("DELETE FROM detalle_pedido WHERE pk_id_pedido = ?", pedidoID).Exec(); err != nil {
		c.failTx(tx, "update.delete_detalles_error", "Error al actualizar los productos del pedido", err, map[string]interface{}{"pedido_id": pedidoID})
		return
	}
	detalles, ok := c.insertDetalles(tx, "update", pedidoID, nuevos)
	if !ok || !c.commit(tx, "update", pedidoID) {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Productos del pedido actualizados exitosamente", productoPedidoData{PedidoID: pedidoID, Detalles: detalles})
}
