package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestSetStaticHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/assets/", nil)
	ctx := context.NewContext()
	ctx.Reset(w, r)
	setStaticHeaders(ctx)
	if got := w.Header().Get("Cache-Control"); got == "" {
		t.Fatalf("expected Cache-Control header to be set")
	}
}

func TestMainNoRun(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	main()
}

func TestAppInitSuccess(t *testing.T) {
	origDB := initDBFunc
	origTZ := initTimezoneFunc
	initDBFunc = func() error { return nil }
	initTimezoneFunc = func() {}
	t.Cleanup(func() { initDBFunc = origDB; initTimezoneFunc = origTZ; dbReady = false })

	appInit()
	if !dbReady {
		t.Fatalf("expected dbReady true")
	}
}

func TestAppInitErrorSetsDbReadyFalse(t *testing.T) {
	origDB := initDBFunc
	origTZ := initTimezoneFunc
	initDBFunc = func() error { return fmt.Errorf("db fail") }
	initTimezoneFunc = func() {}
	t.Cleanup(func() { initDBFunc = origDB; initTimezoneFunc = origTZ; dbReady = false })

	appInit()
	if dbReady {
		t.Fatalf("expected dbReady false when initDB fails")
	}
}

func TestMainStartsCronWhenDBReady(t *testing.T) {
	origDB := initDBFunc
	initDBFunc = func() error { return nil }
	initTimezoneFunc = func() {}
	t.Cleanup(func() { initDBFunc = origDB; dbReady = false })
	appInit()

	os.Setenv("SKIP_WEB_RUN", "1")
	os.Unsetenv("SKIP_CRON")
	os.Setenv("CRON_ONE_SHOT", "1")
	defer func() { os.Unsetenv("CRON_ONE_SHOT"); os.Unsetenv("SKIP_WEB_RUN") }()

	origOrm := cronNewOrm
	origNow := nowFn
	origSleep := sleepFn
	cronNewOrm = func() orm.Ormer { return nil }
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC) }
	sleepFn = func(time.Duration) {}
	t.Cleanup(func() { cronNewOrm = origOrm; nowFn = origNow; sleepFn = origSleep })

	main()
	time.Sleep(10 * time.Millisecond)
}

func TestMainRunWeb(t *testing.T) {
	os.Unsetenv("SKIP_WEB_RUN")
	os.Setenv("SKIP_CRON", "1")
	defer func() { os.Unsetenv("SKIP_CRON") }()

	called := 0
	orig := webRun
	webRun = func(...string) { called++ }
	t.Cleanup(func() { webRun = orig })

	main()
	if called == 0 {
		t.Fatalf("expected webRun to be called")
	}
}

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 1, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

func TestGenerarNominaAutomatica_OneShot(t *testing.T) {
	// Forzar que sea medianoche
	origNow := nowFn
	origSleep := sleepFn
	origInsert := cronInsertNom
	origExec := cronRawExec
	origOrm := cronNewOrm
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	sleepFn = func(d time.Duration) {}
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { return 1, nil }
	cronRawExec = func(o orm.Ormer, query string, args ...interface{}) (sql.Result, error) { return fakeResult{}, nil }
	cronNewOrm = func() orm.Ormer { return nil }
	t.Cleanup(func() {
		nowFn = origNow
		sleepFn = origSleep
		cronInsertNom = origInsert
		cronRawExec = origExec
		cronNewOrm = origOrm
	})

	os.Setenv("CRON_ONE_SHOT", "1")
	defer os.Unsetenv("CRON_ONE_SHOT")

	go generarNominaAutomatica()
	// dar un respiro
	time.Sleep(10 * time.Millisecond)
}

func TestGenerarNominaAutomatica_NotMidnightOneShot(t *testing.T) {
	origNow := nowFn
	origSleep := sleepFn
	origInsert := cronInsertNom
	origOrm := cronNewOrm
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC) }
	sleepFn = func(d time.Duration) {}
	called := 0
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { called++; return 0, nil }
	cronNewOrm = func() orm.Ormer { return nil }
	t.Cleanup(func() { nowFn = origNow; sleepFn = origSleep; cronInsertNom = origInsert; cronNewOrm = origOrm })

	os.Setenv("CRON_ONE_SHOT", "1")
	defer os.Unsetenv("CRON_ONE_SHOT")

	go generarNominaAutomatica()
	time.Sleep(10 * time.Millisecond)
	if called != 0 {
		t.Fatalf("expected no insert when not midnight, got %d", called)
	}
}

func TestGenerarNominaAutomatica_Sleep(t *testing.T) {
	origNow := nowFn
	origSleep := sleepFn
	origOrm := cronNewOrm
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC) }
	called := 0
	sleepFn = func(d time.Duration) { called++; os.Setenv("CRON_ONE_SHOT", "1") }
	cronNewOrm = func() orm.Ormer { return nil }
	t.Cleanup(func() { nowFn = origNow; sleepFn = origSleep; cronNewOrm = origOrm; os.Unsetenv("CRON_ONE_SHOT") })
	go generarNominaAutomatica()
	time.Sleep(10 * time.Millisecond)
	if called == 0 {
		t.Fatalf("expected sleep to be called")
	}
}

// Añadido: cobertura de prints y errores del bloque de nómina automática
// Nota: se usa stdout capturing; mantener simple para estabilidad.
func TestGenerarNominaAutomatica_PrintsAndErrors(t *testing.T) {
	os.Setenv("CRON_ONE_SHOT", "1")
	defer os.Unsetenv("CRON_ONE_SHOT")

	origNow := nowFn
	nowFn = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, database.BogotaZone) }
	defer func() { nowFn = origNow }()

	// Preparar logger de prueba que escribe a un buffer
	origLogger := slog.Default()
	setBufLogger := func() *bytes.Buffer {
		var b bytes.Buffer
		h := slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(h))
		return &b
	}
	restore := func() { slog.SetDefault(origLogger) }
	defer restore()

	// Caso éxito
	b := setBufLogger()
	origI := cronInsertNom
	origR := cronRawExec
	origOrm := cronNewOrm
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { n.PK_ID_NOMINA = 5; return 1, nil }
	cronRawExec = func(o orm.Ormer, q string, args ...interface{}) (sql.Result, error) { return fakeResult{}, nil }
	cronNewOrm = func() orm.Ormer { return nil }
	generarNominaAutomatica()
	out := b.String()
	if !contains(out, "cron.nomina.start") {
		t.Fatalf("expected cron start event, got: %s", out)
	}
	if !contains(out, "cron.nomina.success") {
		t.Fatalf("expected cron success event, got: %s", out)
	}

	// Caso error insert
	b = setBufLogger()
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { return 0, fmt.Errorf("insert fail") }
	generarNominaAutomatica()
	out2 := b.String()
	if !contains(out2, "cron.nomina.insert_err") {
		t.Fatalf("expected insert error event, got: %s", out2)
	}

	// Caso error procedimiento
	b = setBufLogger()
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { n.PK_ID_NOMINA = 6; return 1, nil }
	cronRawExec = func(o orm.Ormer, q string, args ...interface{}) (sql.Result, error) { return nil, fmt.Errorf("proc fail") }
	generarNominaAutomatica()
	out3 := b.String()
	if !contains(out3, "cron.nomina.exec_err") {
		t.Fatalf("expected exec error event, got: %s", out3)
	}
	cronInsertNom = origI
	cronRawExec = origR
	cronNewOrm = origOrm
}

func TestDefaultCronHelpers(t *testing.T) {
	if _, err := cronInsertNom(nil, &models.Nomina{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := cronRawExec(nil, "q"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o := new(orm.DoNothingOrm)
	cronInsertNom(o, &models.Nomina{})
	cronRawExec(o, "q")
	cronRawExec(dummyOrmer{}, "q")
}

type dummyRawSeter struct{}

func (dummyRawSeter) Exec() (sql.Result, error)                               { return fakeResult{}, nil }
func (dummyRawSeter) QueryRow(...interface{}) error                           { return nil }
func (dummyRawSeter) QueryRows(...interface{}) (int64, error)                 { return 0, nil }
func (dummyRawSeter) SetArgs(...interface{}) orm.RawSeter                     { return dummyRawSeter{} }
func (dummyRawSeter) Values(*[]orm.Params, ...string) (int64, error)          { return 0, nil }
func (dummyRawSeter) ValuesList(*[]orm.ParamsList, ...string) (int64, error)  { return 0, nil }
func (dummyRawSeter) ValuesFlat(*orm.ParamsList, ...string) (int64, error)    { return 0, nil }
func (dummyRawSeter) RowsToMap(*orm.Params, string, string) (int64, error)    { return 0, nil }
func (dummyRawSeter) RowsToStruct(interface{}, string, string) (int64, error) { return 0, nil }
func (dummyRawSeter) Prepare() (orm.RawPreparer, error)                       { return nil, nil }

type dummyOrmer struct{ *orm.DoNothingOrm }

func (dummyOrmer) Raw(string, ...interface{}) orm.RawSeter { return dummyRawSeter{} }

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && (stringContains(s, sub))))
}
func stringContains(s, sub string) bool {
	return (len(sub) == 0) || (len(s) >= len(sub) && (indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
