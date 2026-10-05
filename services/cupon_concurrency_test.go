package services

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bancoRedenciones emula, sobre el driver falso, el comportamiento de
// PostgreSQL relevante para la redención: SELECT ... FOR UPDATE bloquea la fila
// hasta el COMMIT/ROLLBACK de la transacción y los INSERT solo son visibles tras
// el COMMIT (READ COMMITTED). Con `sinBloqueo` ignora FOR UPDATE para demostrar
// que, sin el bloqueo, los topes SÍ se superan (la prueba detecta la carrera).
type bancoRedenciones struct {
	mu         sync.Mutex
	redenc     []redencionFila
	cuponLock  sync.Mutex
	pedidoLock map[int64]*sync.Mutex
	sinBloqueo bool
	// cliente de cada pedido (por id de pedido)
	clienteDe func(pedido int64) int64
}

type redencionFila struct{ cupon, cliente, pedido int64 }

func (b *bancoRedenciones) lockPedido(id int64) *sync.Mutex {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.pedidoLock == nil {
		b.pedidoLock = map[int64]*sync.Mutex{}
	}
	if b.pedidoLock[id] == nil {
		b.pedidoLock[id] = &sync.Mutex{}
	}
	return b.pedidoLock[id]
}

func (b *bancoRedenciones) total() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.redenc)
}

func (b *bancoRedenciones) contar(match func(redencionFila) bool) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	var n int64
	for _, r := range b.redenc {
		if match(r) {
			n++
		}
	}
	return n
}

func (b *bancoRedenciones) instalar(maxUsos, limiteCliente *int) {
	fk.reset()
	fk.rows["pedido"] = func(c *fkConn, q string, args []driver.NamedValue) ([]map[string]driver.Value, error) {
		id := args[0].Value.(int64)
		if !b.sinBloqueo && strings.Contains(q, "FOR UPDATE") {
			c.holdLock(b.lockPedido(id))
		}
		fila := fkPedido(b.clienteDe(id), models.EstadoPedidoIniciado, nil, nil)
		fila["pk_id_pedido"] = id
		return []map[string]driver.Value{fila}, nil
	}
	fk.rows["detalle_pedido"] = rowsFn(vals{"pk_id_detalle": int64(1), "pk_id_pedido": int64(9), "pk_id_producto": int64(2), "precio": int64(5000), "cantidad": int64(2)})
	fk.rows["cupon"] = func(c *fkConn, q string, _ []driver.NamedValue) ([]map[string]driver.Value, error) {
		if !b.sinBloqueo && strings.Contains(q, "FOR UPDATE") {
			c.holdLock(&b.cuponLock)
		}
		row := fkCuponRow()
		if maxUsos != nil {
			row["max_usos"] = int64(*maxUsos)
		}
		if limiteCliente != nil {
			row["limite_por_cliente"] = int64(*limiteCliente)
		}
		return []map[string]driver.Value{row}, nil
	}
	fk.counts["cupon_redencion"] = func(_ *fkConn, q string, args []driver.NamedValue) (int64, error) {
		// Ventana entre el conteo y el INSERT: sin bloqueo, todos cuentan antes de insertar.
		defer time.Sleep(15 * time.Millisecond)
		switch {
		case strings.Contains(q, `T0."pk_id_pedido" = $2`):
			pedido := args[1].Value.(int64)
			return b.contar(func(r redencionFila) bool { return r.pedido == pedido }), nil
		case strings.Contains(q, `T0."pk_documento_cliente" = $2`):
			cliente := args[1].Value.(int64)
			return b.contar(func(r redencionFila) bool { return r.cliente == cliente }), nil
		}
		return b.contar(func(redencionFila) bool { return true }), nil
	}
	fk.onInsert = func(c *fkConn, table string, args []driver.NamedValue) {
		if table != "cupon_redencion" {
			return
		}
		fila := redencionFila{cupon: args[0].Value.(int64), cliente: args[1].Value.(int64), pedido: args[2].Value.(int64)}
		c.onCommit(func() {
			b.mu.Lock()
			b.redenc = append(b.redenc, fila)
			b.mu.Unlock()
		})
	}
}

type resultado struct {
	err error
}

// redimirEnParalelo lanza una redención por cada pedido en `pedidos`, todas a la vez.
func redimirEnParalelo(t *testing.T, b *bancoRedenciones, pedidos []int64) (ok, conflictos int) {
	t.Helper()
	var wg sync.WaitGroup
	out := make([]resultado, len(pedidos))
	start := make(chan struct{})
	for i, p := range pedidos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			pid := p
			_, err := NewCuponServiceFromOrm(orm.NewOrm()).RedimirCupon(context.Background(), "VERANO",
				&models.RedimirCuponRequest{ClienteId: b.clienteDe(pid), PedidoId: &pid})
			out[i] = resultado{err: err}
		}()
	}
	close(start)
	wg.Wait()
	for _, r := range out {
		switch {
		case r.err == nil:
			ok++
		case errors.Is(r.err, ErrCuponConflicto):
			conflictos++
		default:
			t.Fatalf("error inesperado: %v", r.err)
		}
	}
	return ok, conflictos
}

func seq(from, n int64) []int64 {
	out := make([]int64, n)
	for i := range out {
		out[i] = from + int64(i)
	}
	return out
}

func intPtr(n int) *int { return &n }

func TestRedimirCupon_Concurrencia_MaxUsos(t *testing.T) {
	b := &bancoRedenciones{clienteDe: func(p int64) int64 { return 1000 + p }}
	b.instalar(intPtr(3), nil)
	ok, conflictos := redimirEnParalelo(t, b, seq(1, 12))
	assert.Equal(t, 3, ok, "exactamente maxUsos redenciones")
	assert.Equal(t, 9, conflictos)
	assert.Equal(t, 3, b.total())
}

func TestRedimirCupon_Concurrencia_LimitePorCliente(t *testing.T) {
	// Un mismo cliente con 8 pedidos intenta redimir a la vez con limitePorCliente = 2.
	b := &bancoRedenciones{clienteDe: func(int64) int64 { return 77 }}
	b.instalar(nil, intPtr(2))
	ok, conflictos := redimirEnParalelo(t, b, seq(1, 8))
	assert.Equal(t, 2, ok)
	assert.Equal(t, 6, conflictos)
	assert.Equal(t, 2, b.total())
}

func TestRedimirCupon_Concurrencia_MismoPedidoNoSeRedimeDosVeces(t *testing.T) {
	b := &bancoRedenciones{clienteDe: func(int64) int64 { return 77 }}
	b.instalar(nil, nil)
	ok, conflictos := redimirEnParalelo(t, b, []int64{5, 5, 5, 5})
	assert.Equal(t, 1, ok)
	assert.Equal(t, 3, conflictos, "el mismo cupón no puede redimirse dos veces en el mismo pedido (409)")
	assert.Equal(t, 1, b.total())
}

func TestRedimirCupon_SinBloqueoLaCarreraSeDetecta(t *testing.T) {
	// Control negativo: el banco ignora FOR UPDATE; todas las transacciones cuentan
	// 0 usos antes de insertar y se supera maxUsos. Demuestra que las pruebas de
	// arriba pasan por el bloqueo y no por casualidad.
	b := &bancoRedenciones{sinBloqueo: true, clienteDe: func(p int64) int64 { return 1000 + p }}
	b.instalar(intPtr(3), nil)
	ok, _ := redimirEnParalelo(t, b, seq(1, 12))
	require.Greater(t, ok, 3, "sin bloqueo se sobrepasa maxUsos")
}
