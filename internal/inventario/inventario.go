// Package inventario bloquea y descuenta el inventario de productos dentro de
// una transacción. Lo comparten /producto_pedido y /pedidos/checkout.
package inventario

import (
	"fmt"
	"sort"
	"strings"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// Resultado es el diagnóstico de Bloquear.
type Resultado struct {
	// NoExisten son los ids de productos que no existen (ascendente).
	NoExisten []int64
	// Insuficientes lista los productos cuyo inventario no alcanza (ascendente por id).
	Insuficientes []models.InventarioInsuficienteDoc
}

// SortedIDs devuelve las llaves de m en orden ascendente (orden estable de
// bloqueo: evita interbloqueos entre transacciones concurrentes).
func SortedIDs[V any](m map[int64]V) []int64 {
	ids := make([]int64, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Bloquear bloquea (FOR UPDATE, por id ascendente) los productos de need (id ->
// unidades requeridas) y verifica que existan y que alcance el inventario.
func Bloquear(tx orm.TxOrmer, need map[int64]int) (Resultado, error) {
	var res Resultado
	if len(need) == 0 {
		return res, nil
	}
	ids := SortedIDs(need)
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
		return res, err
	}
	avail := make(map[int64]int, len(rows))
	for _, r := range rows {
		avail[r.PK] = r.Cantidad
	}
	for _, id := range ids {
		disp, ok := avail[id]
		switch {
		case !ok:
			res.NoExisten = append(res.NoExisten, id)
		case disp < need[id]:
			res.Insuficientes = append(res.Insuficientes, models.InventarioInsuficienteDoc{ProductoId: id, Requerido: need[id], Disponible: disp})
		}
	}
	return res, nil
}

// Descontar ajusta el inventario: delta > 0 descuenta unidades y delta < 0 las
// devuelve (delta 0 no toca la fila). Ante un error devuelve el id del producto
// que falló.
func Descontar(tx orm.TxOrmer, deltas map[int64]int) (int64, error) {
	for _, pid := range SortedIDs(deltas) {
		if deltas[pid] == 0 {
			continue
		}
		if _, err := tx.Raw("UPDATE producto SET cantidad = cantidad - ? WHERE pk_id_producto = ?", deltas[pid], pid).Exec(); err != nil {
			return pid, err
		}
	}
	return 0, nil
}
