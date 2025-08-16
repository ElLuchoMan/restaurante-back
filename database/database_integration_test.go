//go:build integration

package database

import (
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

// Este test SOLO corre si pasas -tags=integration y tienes config de DB real.
// Si no hay DSN/vars mínimas, se omite.
func TestInitDB_Integration(t *testing.T) {
	// Asegura config cargada (por si llaman este test en forma aislada)
	_ = os.Setenv("BEEGO_CONFIG_PATH", "..")
	if err := web.LoadAppConfig("ini", "conf/app.test.conf"); err != nil {
		_ = web.LoadAppConfig("ini", "conf/app.conf")
	}
	web.BConfig.RunMode = "test"

	// Si no hay datos de conexión, omite en integración.
	if os.Getenv("DB_DSN") == "" &&
		(os.Getenv("DB_HOST") == "" || os.Getenv("DB_PORT") == "" ||
			os.Getenv("DB_USER") == "" || os.Getenv("DB_NAME") == "") {
		t.Skip("Sin configuración real de DB; omitiendo integración")
	}

	// InitDB no retorna valor; solo error vía log.Fatal si falla.
	// Si llegara a hacer panic/fatal, este test lo evidenciaría.
	InitDB()
}
