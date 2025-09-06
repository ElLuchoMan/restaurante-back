package database

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// Test mínimo para que el paquete database tenga cobertura sin tocar la DB real.
func TestDatabasePackageLoads(t *testing.T) {
	to := time.Now()
	if to.IsZero() {
		t.Log("noop")
	}
}

// Verifica que InitTimezone cargue correctamente la zona horaria de Bogotá.
func TestInitTimezone(t *testing.T) {
	BogotaZone = nil
	InitTimezone()
	if BogotaZone == nil {
		t.Fatal("BogotaZone no fue inicializada")
	}
	if BogotaZone.String() != "America/Bogota" {
		t.Fatalf("zona horaria inesperada: %s", BogotaZone.String())
	}
}

// Verifica que InitTimezone use la zona UTC-5 cuando no encuentra los datos de zona horaria.
func TestInitTimezoneFallback(t *testing.T) {
	BogotaZone = nil
	orig := loadLocation
	loadLocation = func(name string) (*time.Location, error) { return nil, errors.New("no tz data") }
	t.Cleanup(func() { loadLocation = orig })

	InitTimezone()
	if BogotaZone == nil {
		t.Fatal("BogotaZone no fue inicializada")
	}
	if BogotaZone.String() != "UTC-5" {
		t.Fatalf("zona horaria inesperada: %s", BogotaZone.String())
	}
}

// InitDB debe fallar cuando los datos de conexión son inválidos.
func TestInitDBReturnsError(t *testing.T) {
	t.Setenv("PGCONNECT_TIMEOUT", "1")
	_ = web.AppConfig.Set("db_host", "127.0.0.1")
	_ = web.AppConfig.Set("db_port", "1") // puerto inválido para forzar error
	_ = web.AppConfig.Set("db_user", "postgres")
	_ = web.AppConfig.Set("db_pass", "bad")
	_ = web.AppConfig.Set("db_name", "test")

	if err := InitDB(); err == nil {
		t.Fatal("se esperaba error de InitDB con configuración inválida")
	}
}

// Verifica que InitDB registre la base de datos correctamente cuando la función de registro no falla.
func TestInitDBSuccess(t *testing.T) {
	_ = web.AppConfig.Set("db_host", "localhost")
	_ = web.AppConfig.Set("db_port", "5432")
	_ = web.AppConfig.Set("db_user", "user")
	_ = web.AppConfig.Set("db_pass", "pass")
	_ = web.AppConfig.Set("db_name", "db")

	var (
		called    bool
		gotAlias  string
		gotDriver string
		gotConn   string
	)
	orig := registerDataBase
	registerDataBase = func(alias, driver, conn string, params ...orm.DBOption) error {
		called = true
		gotAlias, gotDriver, gotConn = alias, driver, conn
		return nil
	}
	t.Cleanup(func() { registerDataBase = orig })

	if err := InitDB(); err != nil {
		t.Fatalf("InitDB devolvió error: %v", err)
	}
	if !called {
		t.Fatal("registerDataBase no fue llamada")
	}
	if gotAlias != "default" || gotDriver != "postgres" {
		t.Fatalf("parámetros inesperados: %s %s", gotAlias, gotDriver)
	}
	if !strings.Contains(gotConn, "host=localhost") || !strings.Contains(gotConn, "port=5432") {
		t.Fatalf("connStr inesperado: %s", gotConn)
	}
}

type fakeOrm struct{ orm.Ormer; inserted int; insertErr error }
func (f *fakeOrm) Insert(interface{}) (int64, error) { f.inserted++; return 1, f.insertErr }

func TestSeedMetodoPago_InsertsWhenEmpty(t *testing.T) {
	origNew := newOrmForSeed
	origCount := countMetodoPagoByTipo
	origInsert := insertFn
	t.Cleanup(func(){ newOrmForSeed = origNew; countMetodoPagoByTipo = origCount; insertFn = origInsert })

	fo := &fakeOrm{}
	newOrmForSeed = func() orm.Ormer { return fo }
	countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) { return 0, nil }
	insertFn = func(o orm.Ormer, model interface{}) (int64, error) { return fo.Insert(model) }

	if err := seedMetodoPago(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fo.inserted < 2 {
		t.Fatalf("expected at least 2 inserts, got %d", fo.inserted)
	}
}

func TestSeedMetodoPago_NoInsertWhenExists(t *testing.T) {
	origNew := newOrmForSeed
	origCount := countMetodoPagoByTipo
	origInsert := insertFn
	t.Cleanup(func(){ newOrmForSeed = origNew; countMetodoPagoByTipo = origCount; insertFn = origInsert })

	fo := &fakeOrm{}
	newOrmForSeed = func() orm.Ormer { return fo }
	countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) { return 1, nil }
	insertFn = func(o orm.Ormer, model interface{}) (int64, error) { return fo.Insert(model) }

	if err := seedMetodoPago(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fo.inserted != 0 {
		t.Fatalf("expected no inserts, got %d", fo.inserted)
	}
}

func TestSeedMetodoPago_ErrorOnCount(t *testing.T) {
	origNew := newOrmForSeed
	origCount := countMetodoPagoByTipo
	t.Cleanup(func(){ newOrmForSeed = origNew; countMetodoPagoByTipo = origCount })

	fo := &fakeOrm{}
	newOrmForSeed = func() orm.Ormer { return fo }
	countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) { return 0, errors.New("db") }

	if err := seedMetodoPago(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSeedMetodoPago_ErrorOnInsert(t *testing.T) {
	origNew := newOrmForSeed
	origCount := countMetodoPagoByTipo
	origInsert := insertFn
	t.Cleanup(func(){ newOrmForSeed = origNew; countMetodoPagoByTipo = origCount; insertFn = origInsert })

	fo := &fakeOrm{insertErr: errors.New("insert fail")}
	newOrmForSeed = func() orm.Ormer { return fo }
	countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) { return 0, nil }
	insertFn = func(o orm.Ormer, model interface{}) (int64, error) { return fo.Insert(model) }

	if err := seedMetodoPago(); err == nil {
		t.Fatalf("expected error on insert")
	}
}