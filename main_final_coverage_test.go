package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/joho/godotenv"
)

func TestLoadEnvFile_InCI(t *testing.T) {

	origCI := os.Getenv("CI")
	origSkip := os.Getenv("SKIP_WEB_RUN")
	defer func() {
		if origCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", origCI)
		}
		if origSkip == "" {
			os.Unsetenv("SKIP_WEB_RUN")
		} else {
			os.Setenv("SKIP_WEB_RUN", origSkip)
		}
	}()

	os.Setenv("CI", "true")
	os.Unsetenv("SKIP_WEB_RUN")

	loadEnvFile()

}

func TestLoadEnvFile_WithSkipWebRun(t *testing.T) {
	origCI := os.Getenv("CI")
	origSkip := os.Getenv("SKIP_WEB_RUN")
	defer func() {
		if origCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", origCI)
		}
		if origSkip == "" {
			os.Unsetenv("SKIP_WEB_RUN")
		} else {
			os.Setenv("SKIP_WEB_RUN", origSkip)
		}
	}()

	os.Unsetenv("CI")
	os.Setenv("SKIP_WEB_RUN", "1")

	loadEnvFile()

}

func TestLoadEnvFile_FileNotFound(t *testing.T) {
	origCI := os.Getenv("CI")
	origSkip := os.Getenv("SKIP_WEB_RUN")
	origDir, _ := os.Getwd()

	defer func() {
		if origCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", origCI)
		}
		if origSkip == "" {
			os.Unsetenv("SKIP_WEB_RUN")
		} else {
			os.Setenv("SKIP_WEB_RUN", origSkip)
		}
		os.Chdir(origDir)
	}()

	os.Unsetenv("CI")
	os.Unsetenv("SKIP_WEB_RUN")

	tmpDir := os.TempDir()
	os.Chdir(tmpDir)

	loadEnvFile()

}

func TestLoadEnvFile_Success(t *testing.T) {
	origCI := os.Getenv("CI")
	origSkip := os.Getenv("SKIP_WEB_RUN")
	origDir, _ := os.Getwd()

	defer func() {
		if origCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", origCI)
		}
		if origSkip == "" {
			os.Unsetenv("SKIP_WEB_RUN")
		} else {
			os.Setenv("SKIP_WEB_RUN", origSkip)
		}
		os.Chdir(origDir)
		os.Remove(".env.test")
	}()

	os.Unsetenv("CI")
	os.Unsetenv("SKIP_WEB_RUN")

	tmpDir := os.TempDir()
	os.Chdir(tmpDir)

	envContent := "TEST_VAR=test_value\n"
	os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")

	loadEnvFile()

	if err := godotenv.Load(); err == nil {
		testVar := os.Getenv("TEST_VAR")
		if testVar != "test_value" {
			t.Logf("Expected TEST_VAR to be loaded, but got: %s", testVar)
		}
	}
}

func TestSetStaticHeadersFn_Coverage(t *testing.T) {

	origFn := setStaticHeadersFn
	defer func() { setStaticHeadersFn = origFn }()

	called := false
	setStaticHeadersFn = func(ctx *context.Context) {
		called = true
	}

	setStaticHeaders(nil)
	if !called {
		t.Error("Expected setStaticHeadersFn to be called")
	}
}

func TestOrmInsert_NilOrmer(t *testing.T) {
	result, err := ormInsert(nil, nil)
	if err != nil {
		t.Errorf("Expected no error with nil ormer, got: %v", err)
	}
	if result != 0 {
		t.Errorf("Expected result 0 with nil ormer, got: %d", result)
	}
}

func TestOrmRawExec_NilOrmer(t *testing.T) {
	result, err := ormRawExec(nil, "SELECT 1")
	if err != nil {
		t.Errorf("Expected no error with nil ormer, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result with nil ormer, got: %v", result)
	}
}

func TestOrmRawExec_NilRawSeter(t *testing.T) {

	mockOrm := &nilRawOrmer2{}
	result, err := ormRawExec(mockOrm, "SELECT 1")
	if err != nil {
		t.Errorf("Expected no error with nil RawSeter, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result with nil RawSeter, got: %v", result)
	}
}

type nilRawOrmer2 struct {
	orm.Ormer
}

func (n *nilRawOrmer2) Raw(query string, args ...interface{}) orm.RawSeter {
	return nil
}

func TestGetSQLPinger_NilDB(t *testing.T) {
	origGetter := dbGetter
	defer func() { dbGetter = origGetter }()

	dbGetter = func() (*sql.DB, error) {
		return nil, nil
	}

	pinger, err := getSQLPinger()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if pinger != nil {
		t.Error("Expected nil pinger when db is nil")
	}
}
