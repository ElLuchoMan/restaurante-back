package database

import (
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

// TestMain se ejecuta ANTES de los tests de este paquete.
// Ajusta el path de configuración para que Beego encuentre conf/app*.conf
func TestMain(m *testing.M) {
	// Cuando go test corre dentro de /database, el root del repo está "arriba".
	// Usamos BEEGO_CONFIG_PATH para que busque conf/*.conf en el root.
	_ = os.Setenv("BEEGO_CONFIG_PATH", "..")

	// Intenta cargar conf/app.test.conf; si no existe, intenta conf/app.conf.
	if err := web.LoadAppConfig("ini", "conf/app.test.conf"); err != nil {
		_ = web.LoadAppConfig("ini", "conf/app.conf")
	}
	// Fuerza runmode test, por si acaso
	_ = web.BConfig.RunMode == "test"

	code := m.Run()
	os.Exit(code)
}
