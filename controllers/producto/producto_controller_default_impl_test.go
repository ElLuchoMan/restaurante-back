package producto

import (
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

// spyOrmer intercepta QueryTable/Read/Insert/Update, pero delega el resto al Ormer embebido
type spyOrmer struct {
	orm.Ormer
	lastQS *spyQS
	// señales para asserts
	readCalled   bool
	insertCalled int
	updateCalled bool
	updateCols   []string
}

func (s *spyOrmer) QueryTable(table interface{}) orm.QuerySeter {
	realQS := s.Ormer.QueryTable(table)
	s.lastQS = &spyQS{QuerySeter: realQS}
	return s.lastQS
}

func (s *spyOrmer) Read(md interface{}, cols ...string) error {
	s.readCalled = true
	// simular lectura exitosa: si es Producto, completar un par de campos
	if p, ok := md.(*models.Producto); ok {
		p.NOMBRE = "Spy"
		p.PRECIO = 100
		p.ESTADO_PRODUCTO = models.EstadoProductoDisponible
	}
	return nil
}

func (s *spyOrmer) Insert(md interface{}) (int64, error) {
	s.insertCalled++
	// simular PK autoincremental
	switch v := md.(type) {
	case *models.Producto:
		v.PK_ID_PRODUCTO = 1
	case *models.PrecioProductoHist:
		// nada extra
	}
	return 1, nil
}

func (s *spyOrmer) Update(md interface{}, cols ...string) (int64, error) {
	s.updateCalled = true
	s.updateCols = append([]string{}, cols...)
	return 1, nil
}

// spyQS es un QuerySeter minimal que captura Filter y materializa en All sin tocar DB
type spyQS struct {
	orm.QuerySeter
	filterField  string
	filterValues []interface{}
}

func (q *spyQS) Filter(field string, values ...interface{}) orm.QuerySeter {
	q.filterField = field
	q.filterValues = append([]interface{}{}, values...)
	return q
}

func (q *spyQS) All(container interface{}, cols ...string) (int64, error) {
	// poblar slice de productos con base en si se aplicó el filtro
	if out, ok := container.(*[]models.Producto); ok {
		if q.filterField == "ESTADO_PRODUCTO" && len(q.filterValues) == 1 && q.filterValues[0] == models.EstadoProductoDisponible {
			*out = []models.Producto{{NOMBRE: "Activo", ESTADO_PRODUCTO: models.EstadoProductoDisponible}}
		} else {
			*out = []models.Producto{{NOMBRE: "Cualquiera", ESTADO_PRODUCTO: models.EstadoProductoNoDisponible}}
		}
		return int64(len(*out)), nil
	}
	return 0, nil
}

func Test_queryProductosAll_DefaultBody_OnlyActiveTrue(t *testing.T) {
	base := orm.NewOrm() // no se usará porque el spy intercepta QueryTable
	o := &spyOrmer{Ormer: base}
	var productos []models.Producto

	n, err := queryProductosAll(o, true, &productos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 || len(productos) != 1 {
		t.Fatalf("expected 1 producto, got n=%d len=%d", n, len(productos))
	}
	if o.lastQS == nil || o.lastQS.filterField != "ESTADO_PRODUCTO" || len(o.lastQS.filterValues) != 1 || o.lastQS.filterValues[0] != models.EstadoProductoDisponible {
		t.Fatalf("expected Filter on ESTADO_PRODUCTO=DISPONIBLE to be applied")
	}
	if productos[0].ESTADO_PRODUCTO != models.EstadoProductoDisponible {
		t.Fatalf("expected producto disponible, got %v", productos[0].ESTADO_PRODUCTO)
	}
}

func Test_queryProductosAll_DefaultBody_OnlyActiveFalse(t *testing.T) {
	base := orm.NewOrm()
	o := &spyOrmer{Ormer: base}
	var productos []models.Producto

	n, err := queryProductosAll(o, false, &productos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 || len(productos) != 1 {
		t.Fatalf("expected 1 producto, got n=%d len=%d", n, len(productos))
	}
	if o.lastQS == nil {
		t.Fatalf("expected QueryTable to be called")
	}
	if o.lastQS.filterField != "" {
		t.Fatalf("did not expect Filter to be applied when onlyActive=false")
	}
}

func Test_readProductoFn_DefaultCallsRead(t *testing.T) {
	o := &spyOrmer{Ormer: orm.NewOrm()}
	p := &models.Producto{PK_ID_PRODUCTO: 1}
	if err := readProductoFn(o, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !o.readCalled {
		t.Fatalf("expected Read to be called")
	}
	if p.NOMBRE == "" || p.PRECIO == 0 {
		t.Fatalf("expected producto to be populated by spy Read")
	}
}

func Test_insertProductoFn_DefaultCallsInsert(t *testing.T) {
	o := &spyOrmer{Ormer: orm.NewOrm()}
	p := &models.Producto{NOMBRE: "X", PRECIO: 10, ESTADO_PRODUCTO: models.EstadoProductoDisponible}
	n, err := insertProductoFn(o, p)
	if err != nil || n != 1 {
		t.Fatalf("unexpected result n=%d err=%v", n, err)
	}
	if o.insertCalled != 1 {
		t.Fatalf("expected Insert to be called once, got %d", o.insertCalled)
	}
	if p.PK_ID_PRODUCTO == 0 {
		t.Fatalf("expected PK to be set by spy Insert")
	}
}

func Test_insertPrecioHistFn_DefaultCallsInsert(t *testing.T) {
	o := &spyOrmer{Ormer: orm.NewOrm()}
	h := &models.PrecioProductoHist{Precio: 10}
	n, err := insertPrecioHistFn(o, h)
	if err != nil || n != 1 {
		t.Fatalf("unexpected result n=%d err=%v", n, err)
	}
	if o.insertCalled != 1 {
		t.Fatalf("expected Insert to be called once, got %d", o.insertCalled)
	}
}

func Test_updateProductoFn_DefaultCallsUpdate(t *testing.T) {
	o := &spyOrmer{Ormer: orm.NewOrm()}
	p := &models.Producto{PK_ID_PRODUCTO: 1, NOMBRE: "Y", PRECIO: 20, ESTADO_PRODUCTO: models.EstadoProductoDisponible}
	n, err := updateProductoFn(o, p, "NOMBRE", "PRECIO")
	if err != nil || n != 1 {
		t.Fatalf("unexpected result n=%d err=%v", n, err)
	}
	if !o.updateCalled {
		t.Fatalf("expected Update to be called")
	}
	if len(o.updateCols) != 2 || o.updateCols[0] != "NOMBRE" || o.updateCols[1] != "PRECIO" {
		t.Fatalf("unexpected cols: %+v", o.updateCols)
	}
}
