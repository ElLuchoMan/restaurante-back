package productopedido

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"restaurante/controllers/login"
	"restaurante/internal/authz"
	"restaurante/internal/httpx"
	"restaurante/internal/inventario"
	"restaurante/internal/montopedido"
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
	msgPedidoCongelado    = "El pedido ya tiene un pago asignado: sus productos no se pueden modificar"
	msgPedidoPagado       = "El pago del pedido ya está PAGADO: sus productos no se pueden modificar"
	msgPedidoDescuento    = "El pedido tiene un descuento aplicado: sus productos no se pueden modificar porque el monto del pago ya lo refleja"
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

// begin abre una transacción, bloquea la fila del pedido y, si el pedido ya
// tiene un pago asignado, valida que sus productos puedan cambiar (ver
// puedeModificar). Devuelve el id del pago (0 si no tiene) y false si ya
// respondió (404 si el pedido no existe o, para un Cliente, no es suyo; 409 si
// está congelado; 500 en otro error).
func (c *ProductoPedidoController) begin(op string, pedidoID int64, claims *login.Claims) (orm.TxOrmer, int64, bool) {
	ctx := map[string]interface{}{"pedido_id": pedidoID}
	tx, err := orm.NewOrm().Begin()
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido."+op+".tx_begin_error", err, ctx)
		c.fail(http.StatusInternalServerError, "No fue posible iniciar transacción", err)
		return nil, 0, false
	}
	var cliente, pagoID int64
	if err := tx.Raw("SELECT COALESCE(pk_documento_cliente, 0), COALESCE(pk_id_pago, 0) FROM pedido WHERE pk_id_pedido = ? FOR UPDATE", pedidoID).QueryRow(&cliente, &pagoID); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			_ = tx.Rollback()
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return nil, 0, false
		}
		c.failTx(tx, op+".lock_pedido_error", "No fue posible bloquear el pedido para actualización", err, ctx)
		return nil, 0, false
	}
	if !authz.EsDuenio(claims, cliente) {
		_ = tx.Rollback()
		c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
		return nil, 0, false
	}
	if pagoID > 0 && !c.puedeModificar(tx, op, claims, pagoID) {
		return nil, 0, false
	}
	return tx, pagoID, true
}

// puedeModificar decide si se pueden cambiar los productos de un pedido que ya
// tiene el pago pagoID (el pedido ya está bloqueado en tx). Un Cliente nunca
// (409): el monto del pago salió de esos productos y no puede cambiar tras
// asignarlo. El personal sí, salvo que el pago esté PAGADO (409) o el pedido
// tenga un descuento aplicado (409: se calculó sobre los productos anteriores);
// en los demás casos begin recalcula el monto del pago (ver recalcularPago).
// Si responde un error deshace tx y devuelve false.
func (c *ProductoPedidoController) puedeModificar(tx orm.TxOrmer, op string, claims *login.Claims, pagoID int64) bool {
	if !claims.IsStaff() {
		_ = tx.Rollback()
		c.fail(http.StatusConflict, msgPedidoCongelado, nil)
		return false
	}
	ctx := map[string]interface{}{"pago_id": pagoID}
	var estado string
	var descuentos int64
	if err := tx.Raw("SELECT estado_pago FROM pago WHERE pk_id_pago = ? FOR UPDATE", pagoID).QueryRow(&estado); err != nil {
		c.failTx(tx, op+".lock_pago_error", "No fue posible bloquear el pago del pedido", err, ctx)
		return false
	}
	if estado == models.EstadoPagoPagado {
		_ = tx.Rollback()
		c.fail(http.StatusConflict, msgPedidoPagado, nil)
		return false
	}
	if err := tx.Raw("SELECT COUNT(*) FROM pedido_descuento_aplicado d JOIN pedido p ON p.pk_id_pedido = d.pk_id_pedido WHERE p.pk_id_pago = ?", pagoID).QueryRow(&descuentos); err != nil {
		c.failTx(tx, op+".descuentos_error", "Error al validar los descuentos del pedido", err, ctx)
		return false
	}
	if descuentos > 0 {
		_ = tx.Rollback()
		c.fail(http.StatusConflict, msgPedidoDescuento, nil)
		return false
	}
	return true
}

// recalcularPago deja en el pago pagoID (PENDIENTE o NO_PAGO, sin descuentos) el
// monto calculado con los productos que quedaron en el pedido. Va en la misma
// transacción que el cambio de productos. Devuelve false si ya respondió.
func (c *ProductoPedidoController) recalcularPago(tx orm.TxOrmer, op string, pedidoID, pagoID int64) bool {
	ctx := map[string]interface{}{"pedido_id": pedidoID, "pago_id": pagoID}
	m, err := montopedido.Calcular(tx, pedidoID)
	if err != nil {
		c.failTx(tx, op+".recalcular_monto_error", "Error al calcular el monto del pedido", err, ctx)
		return false
	}
	if _, err := tx.Raw("UPDATE pago SET monto = ?, updated_at = ? WHERE pk_id_pago = ?", m.Total, time.Now().UTC(), pagoID).Exec(); err != nil {
		c.failTx(tx, op+".actualizar_pago_error", "Error al actualizar el monto del pago", err, ctx)
		return false
	}
	return true
}

// checkStock bloquea los productos de need (id -> unidades que faltan por
// descontar) y verifica que existan (404) y que alcance el inventario (409,
// con el detalle en `data`). Devuelve false si ya respondió.
func (c *ProductoPedidoController) checkStock(tx orm.TxOrmer, op string, need map[int64]int) bool {
	res, err := inventario.Bloquear(tx, need)
	if err != nil {
		c.failTx(tx, op+".validar_inventario_error", "Error al validar inventario", err, nil)
		return false
	}
	if len(res.NoExisten) > 0 {
		missing := make([]string, len(res.NoExisten))
		for i, id := range res.NoExisten {
			missing[i] = fmt.Sprint(id)
		}
		_ = tx.Rollback()
		c.fail(http.StatusNotFound, "Producto no encontrado: "+strings.Join(missing, ", "), nil)
		return false
	}
	if len(res.Insuficientes) > 0 {
		_ = tx.Rollback()
		httpx.Send(&c.Controller, http.StatusConflict, msgInventario, res.Insuficientes)
		return false
	}
	return true
}

// applyDeltas ajusta el inventario: delta > 0 descuenta unidades y delta < 0
// las devuelve. Devuelve false si ya respondió.
func (c *ProductoPedidoController) applyDeltas(tx orm.TxOrmer, op string, deltas map[int64]int) bool {
	if pid, err := inventario.Descontar(tx, deltas); err != nil {
		c.failTx(tx, op+".stock_update_error", "Error al ajustar inventario", err, map[string]interface{}{"productoId": pid, "delta": deltas[pid]})
		return false
	}
	return true
}

// insertDetalles inserta una línea por producto y la relee para devolver el
// precio unitario que fija la base de datos. Devuelve false si ya respondió.
func (c *ProductoPedidoController) insertDetalles(tx orm.TxOrmer, op string, pedidoID int64, items map[int64]int) ([]models.DetallePedido, bool) {
	detalles := []models.DetallePedido{}
	for _, pid := range inventario.SortedIDs(items) {
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

// cerrar recalcula el monto del pago del pedido (si ya tiene uno) y confirma la
// transacción. Devuelve false si ya respondió.
func (c *ProductoPedidoController) cerrar(tx orm.TxOrmer, op string, pedidoID, pagoID int64) bool {
	if pagoID > 0 && !c.recalcularPago(tx, op, pedidoID, pagoID) {
		return false
	}
	return c.commit(tx, op, pedidoID)
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
// @Description Devuelve las líneas (detalles) de un pedido. Un Cliente solo puede leer las de su propio pedido (uno ajeno responde 404, igual que si no existiera); el personal lee cualquiera. Un pedido existente sin productos responde 200 con `detalles` igual a `[]`; un pedido inexistente responde 404. En cada línea `pedidoId` y `productoId` son objetos de la relación (solo el id es fiable) y `precio` es el precio unitario.
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos del pedido (`detalles` puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "pedido_id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado (para un Cliente, también si es de otro cliente)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /producto_pedido [get]
func (c *ProductoPedidoController) GetAll() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "producto_pedido.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.fail(http.StatusBadRequest, msgPedidoIDInvalido, err)
		return
	}
	o := orm.NewOrm()
	if !claims.IsStaff() {
		// Un Cliente solo lee los productos de su propio pedido; uno ajeno responde igual que uno inexistente.
		n, err := o.QueryTable(new(models.Pedido)).Filter("PK_ID_PEDIDO", pedidoID).Filter("PK_DOCUMENTO_CLIENTE", claims.Documento).Count()
		if err != nil {
			logging.LogControllerError(c.Ctx, "producto_pedido.getall.pedido_error", err, map[string]interface{}{"pedido_id": pedidoID})
			c.fail(http.StatusInternalServerError, "Error al consultar el pedido", err)
			return
		}
		if claims.Documento <= 0 || n == 0 {
			c.fail(http.StatusNotFound, msgPedidoNoEncontrado, nil)
			return
		}
	}
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
// @Description Un Cliente solo puede agregar productos a su propio pedido (uno ajeno responde 404, igual que si no existiera) y solo mientras el pedido NO tenga pago asignado: después responde 409 (el monto del pago salió de esos productos). El personal puede agregar a cualquiera; si el pedido tiene un pago PENDIENTE/NO_PAGO sin descuentos, el monto del pago se recalcula en la misma transacción (`MAX(0, SUM(precio x cantidad) - descuentos)`); si el pago está PAGADO o el pedido tiene un descuento aplicado, responde 409. Agrega las líneas indicadas a un pedido existente y descuenta el inventario, todo en una transacción. Las líneas repetidas se suman y las de `cantidad` 0 se ignoran (debe quedar al menos una con `cantidad` > 0); `productoId` <= 0 o `cantidad` negativa responden 400. Responde 404 si el pedido o algún producto no existe, 409 si el inventario no alcanza (`data` lista {productoId, requerido, disponible}) o si el producto ya está en el pedido (use PUT para modificarlo).
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param body body models.ProductoPedidoCreateRequest true "Pedido y sus productos"
// @Success 201 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos agregados exitosamente"
// @Failure 400 {object} models.ApiResponse "JSON inválido, pedidoId inválido, sin líneas válidas, productoId <= 0 o cantidad negativa"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido o producto no encontrado (para un Cliente, el pedido ajeno también)"
// @Failure 409 {object} models.ApiResponse{data=[]models.InventarioInsuficienteDoc} "Inventario insuficiente (con detalle en `data`), producto ya presente en el pedido, o pedido congelado (Cliente con pago asignado; personal con pago PAGADO o descuento aplicado)"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [post]
func (c *ProductoPedidoController) Post() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
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

	tx, pagoID, ok := c.begin("post", input.PedidoId, claims)
	if !ok || !c.checkStock(tx, "post", nuevos) || !c.applyDeltas(tx, "post", nuevos) {
		return
	}
	detalles, ok := c.insertDetalles(tx, "post", input.PedidoId, nuevos)
	if !ok || !c.cerrar(tx, "post", input.PedidoId, pagoID) {
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Pedido con productos agregado exitosamente", productoPedidoData{PedidoID: input.PedidoId, Detalles: detalles})
}

// @Title Update
// @Summary Reemplazar los productos de un pedido
// @Description Un Cliente solo puede modificar los productos de su propio pedido (uno ajeno responde 404, igual que si no existiera) y solo mientras el pedido NO tenga pago asignado: después responde 409 (el monto del pago salió de esos productos). El personal puede modificar los de cualquiera; si el pedido tiene un pago PENDIENTE/NO_PAGO sin descuentos, el monto del pago se recalcula en la misma transacción (`MAX(0, SUM(precio x cantidad) - descuentos)`); si el pago está PAGADO o el pedido tiene un descuento aplicado, responde 409. Reemplaza las líneas del pedido por la lista enviada (el cuerpo es un arreglo, no un objeto, así que no aplica el merge por campos): los productos que no aparezcan se quitan y una línea con `cantidad` 0 también quita el producto. El inventario se ajusta con la diferencia (descuenta o devuelve) en una transacción. Las líneas repetidas se suman; `productoId` <= 0 o `cantidad` negativa responden 400; la lista no puede estar vacía. Responde 404 si el pedido o algún producto a descontar no existe y 409 si el inventario no alcanza (`data` lista {productoId, requerido, disponible}).
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param body body models.ProductoPedidoUpdateRequest true "Lista completa de productos que debe quedar en el pedido"
// @Success 200 {object} models.ApiResponse{data=models.ProductoPedidoDoc} "Productos actualizados exitosamente"
// @Failure 400 {object} models.ApiResponse "pedido_id inválido, cuerpo que no es un arreglo, lista vacía, productoId <= 0 o cantidad negativa"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido o producto no encontrado (para un Cliente, el pedido ajeno también)"
// @Failure 409 {object} models.ApiResponse{data=[]models.InventarioInsuficienteDoc} "Inventario insuficiente (con detalle en `data`) o pedido congelado (Cliente con pago asignado; personal con pago PAGADO o descuento aplicado)"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [put]
func (c *ProductoPedidoController) Update() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
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

	tx, pagoID, ok := c.begin("update", pedidoID, claims)
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
	if !ok || !c.cerrar(tx, "update", pedidoID, pagoID) {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Productos del pedido actualizados exitosamente", productoPedidoData{PedidoID: pedidoID, Detalles: detalles})
}
