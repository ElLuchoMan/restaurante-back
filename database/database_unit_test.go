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
	t.Log("database unit test ok")
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
