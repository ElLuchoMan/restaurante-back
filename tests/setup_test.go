package test

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"restaurante/database"

	beego "github.com/beego/beego/v2/server/web"
)

// TestMain sets up Beego and a dedicated test database before running tests.
func TestMain(m *testing.M) {
	_, file, _, _ := runtime.Caller(0)
	appPath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	os.Setenv("BEEGO_APP_CONFIG_FILE", filepath.Join(appPath, "conf", "app.test.conf"))
	os.Chdir(appPath)
	beego.TestBeegoInit(appPath)

	// Saltar integración por defecto: solo correr con INTEGRATION=1
	integration := os.Getenv("INTEGRATION") == "1"
	if !integration {
		// Desactivar seed y silenciar en modo no integración
		_ = os.Setenv("SKIP_DB_SEED", "1")
		_ = os.Setenv("QUIET_TESTS", "1")
	}

	// Initialize the database using the test configuration
	if err := database.InitDB(); err != nil {
		if integration {
			log.Println("Database unavailable, integration tests may fail:", err)
		}
	} else if integration {
		// Solo poblar datos cuando se ejecuta en modo integración
		SeedTestData()
	}

	code := m.Run()
	os.Exit(code)
}
