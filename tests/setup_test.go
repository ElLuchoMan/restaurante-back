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

        // Initialize the database using the test configuration
        if err := database.InitDB(); err != nil {
                // Log the error but allow tests that do not require the
                // database to run. This avoids reporting a coverage of
                // "[no statements]" when the database is unavailable.
                log.Println("Database unavailable, tests will run without DB:", err)
        }

        code := m.Run()
        os.Exit(code)
}
