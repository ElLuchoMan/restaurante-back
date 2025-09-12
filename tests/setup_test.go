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

func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "testsecret")
	}
	_, file, _, _ := runtime.Caller(0)
	appPath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	os.Setenv("BEEGO_APP_CONFIG_FILE", filepath.Join(appPath, "conf", "app.test.conf"))
	os.Chdir(appPath)
	beego.TestBeegoInit(appPath)

	integration := os.Getenv("INTEGRATION") == "1"
	if !integration {
		_ = os.Setenv("SKIP_DB_SEED", "1")
		_ = os.Setenv("QUIET_TESTS", "1")
	}

	if err := database.InitDB(); err != nil {
		if integration {
			log.Println("Database unavailable, integration tests may fail:", err)
		}
	} else if integration {
		SeedTestData()
	}

	code := m.Run()
	os.Exit(code)
}
