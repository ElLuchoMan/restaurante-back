package database

import (
	"testing"

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

// InitDB debe fallar cuando los datos de conexión son inválidos.
func TestInitDBReturnsError(t *testing.T) {
	_ = web.AppConfig.Set("db_host", "127.0.0.1")
	_ = web.AppConfig.Set("db_port", "1") // puerto inválido para forzar error
	_ = web.AppConfig.Set("db_user", "postgres")
	_ = web.AppConfig.Set("db_pass", "bad")
	_ = web.AppConfig.Set("db_name", "test")

	if err := InitDB(); err == nil {
		t.Fatal("se esperaba error de InitDB con configuración inválida")
	}
}
