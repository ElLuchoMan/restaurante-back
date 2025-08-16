package test

import (
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

	// Initialize the database using the test configuration
	database.InitDB()

	code := m.Run()
	os.Exit(code)
}
