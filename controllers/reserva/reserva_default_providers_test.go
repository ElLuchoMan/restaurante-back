package reserva

import (
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// Referencias a las implementaciones por defecto, capturadas antes de que
// otros tests las sustituyan por dobles.
var (
	defaultQueryByDocumentoCliente  = queryReservasByDocumentoCliente
	defaultQueryByDocumentoContacto = queryReservasByDocumentoContacto
	defaultInsertReservaContacto    = insertReservaContacto
	defaultQueryContactoPorDoc      = queryReservaContactoByDocumento
	defaultQueryContactoPorCliente  = queryReservaContactoByCliente
	defaultReadCliente              = readCliente
)

// Ejecuta los proveedores ORM por defecto contra el driver SQL falso; solo se
// comprueba que recorren su camino (con y sin filtro de fecha), el resultado
// del driver falso no es relevante.
func TestReservaDefaultORMProviders(t *testing.T) {
	o := orm.NewOrm()
	var reservas []models.Reserva
	fecha := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	for _, useFecha := range []bool{false, true} {
		_, _ = defaultQueryByDocumentoCliente(o, 1, fecha, useFecha, &reservas)
		_, _ = defaultQueryByDocumentoContacto(o, 1, fecha, useFecha, &reservas)
	}

	_, _ = defaultInsertReservaContacto(o, &models.ReservaContacto{})
	_ = defaultQueryContactoPorDoc(o, 1, &models.ReservaContacto{})
	_ = defaultQueryContactoPorCliente(o, 1, &models.ReservaContacto{})
	_ = defaultReadCliente(o, &models.Cliente{})
}
