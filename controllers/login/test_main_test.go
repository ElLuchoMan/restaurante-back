package login

import (
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "testsecret")
	}
	if web.BConfig.RunMode == "" {
		web.BConfig.RunMode = "dev"
	}
	os.Exit(m.Run())
}
