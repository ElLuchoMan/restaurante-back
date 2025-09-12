//go:build integration

package database

import (
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestInitDB_Integration(t *testing.T) {
	_ = os.Setenv("BEEGO_CONFIG_PATH", "..")
	if err := web.LoadAppConfig("ini", "conf/app.test.conf"); err != nil {
		_ = web.LoadAppConfig("ini", "conf/app.conf")
	}
	web.BConfig.RunMode = "test"

	if os.Getenv("DB_DSN") == "" &&
		(os.Getenv("DB_HOST") == "" || os.Getenv("DB_PORT") == "" ||
			os.Getenv("DB_USER") == "" || os.Getenv("DB_NAME") == "") {
		t.Skip("Sin configuración real de DB; omitiendo integración")
	}

	if err := InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
}
