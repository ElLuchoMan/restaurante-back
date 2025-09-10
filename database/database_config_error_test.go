package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

type errConfig struct{ web.AppConfiger }

func (e *errConfig) String(key string) (string, error) { return "", errors.New("config missing") }

func TestInitDB_ConfigErrors(t *testing.T) {
	t.Setenv("SKIP_DB_SEED", "1")
	t.Setenv("QUIET_TESTS", "1")

	orig := web.AppConfig
	web.AppConfig = &errConfig{orig}
	t.Cleanup(func() { web.AppConfig = orig })

	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"db_host", nil, "config db_host"},
		{"db_port", map[string]string{"DB_HOST": "h"}, "config db_port"},
		{"db_user", map[string]string{"DB_HOST": "h", "DB_PORT": "1"}, "config db_user"},
		{"db_pass", map[string]string{"DB_HOST": "h", "DB_PORT": "1", "DB_USER": "u"}, "config db_pass"},
		{"db_name", map[string]string{"DB_HOST": "h", "DB_PORT": "1", "DB_USER": "u", "DB_PASS": "p"}, "config db_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			err := InitDB()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}
