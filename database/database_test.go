package database

import (
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestMain(m *testing.M) {

	if os.Getenv("BEEGO_APP_CONFIG_FILE") == "" {
		_ = os.Setenv("BEEGO_APP_CONFIG_FILE", "conf/app.test.conf")
	}
	web.BConfig.RunMode = "test"
	code := m.Run()
	os.Exit(code)
}
