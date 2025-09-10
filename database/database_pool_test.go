package database

import (
	"database/sql"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

func TestGetDefaultSQLDB(t *testing.T) {
	db := &sql.DB{}
	orig := getDB
	getDB = func(aliasNames ...string) (*sql.DB, error) {
		if len(aliasNames) != 1 || aliasNames[0] != "default" {
			t.Fatalf("unexpected alias: %v", aliasNames)
		}
		return db, nil
	}
	t.Cleanup(func() { getDB = orig })

	got, err := GetDefaultSQLDB()
	if err != nil {
		t.Fatalf("GetDefaultSQLDB returned error: %v", err)
	}
	if got != db {
		t.Fatalf("unexpected db: %v", got)
	}
}

func TestInitDBConfiguresPool(t *testing.T) {
	t.Setenv("QUIET_TESTS", "1")
	t.Setenv("DB_MAX_OPEN", "5")
	t.Setenv("DB_MAX_IDLE", "4")
	t.Setenv("DB_CONN_MAX_LIFETIME_MIN", "3")
	t.Setenv("DB_CONN_MAX_IDLE_TIME_MIN", "2")
	t.Setenv("SKIP_DB_SEED", "1")

	_ = web.AppConfig.Set("db_host", "localhost")
	_ = web.AppConfig.Set("db_port", "5432")
	_ = web.AppConfig.Set("db_user", "user")
	_ = web.AppConfig.Set("db_pass", "pass")
	_ = web.AppConfig.Set("db_name", "db")

	sqlDB := &sql.DB{}
	origReg := registerDataBase
	registerDataBase = func(alias, driver, conn string, params ...orm.DBOption) error { return nil }
	origGet := getDB
	getDB = func(aliasNames ...string) (*sql.DB, error) { return sqlDB, nil }
	t.Cleanup(func() { registerDataBase = origReg; getDB = origGet })

	if err := InitDB(); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	if sqlDB.Stats().MaxOpenConnections != 5 {
		t.Fatalf("expected MaxOpenConnections 5, got %d", sqlDB.Stats().MaxOpenConnections)
	}
}
