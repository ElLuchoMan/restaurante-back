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
	os.Setenv("cocina-de-maria", "test-secret")
	os.Chdir(appPath)
	beego.TestBeegoInit(appPath)

	// Initialize the database using the test configuration
	if err := database.InitDB(); err != nil {
		log.Println("Skipping tests: database unavailable:", err)
		os.Exit(0)
	}

	code := m.Run()
	os.Exit(code)
}
