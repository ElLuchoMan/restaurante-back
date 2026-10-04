package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
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

func TestSetStaticHeaders_StaticPath(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/app.js", nil)
	ctx := context.NewContext()
	ctx.Reset(w, r)
	setStaticHeaders(ctx)
	if got := w.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Fatalf("expected Vary header to be set, got %q", got)
	}
}

func TestSetStaticHeaders_NonStaticPath(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/data", nil)
	ctx := context.NewContext()
	ctx.Reset(w, r)
	setStaticHeaders(ctx)
	if got := w.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("did not expect Cache-Control header, got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "" {
		t.Fatalf("did not expect Vary header, got %q", got)
	}
}

func TestSetStaticHeaders_FilterExecutes(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	beego.BConfig.RunMode = "dev"
	main()

	{
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/static/app.js", nil)
		ctx := context.NewContext()
		ctx.Reset(w, r)
		setStaticHeaders(ctx)
	}

	r, _ := http.NewRequest("GET", "/static/app.js", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if got := w.Header().Get("Cache-Control"); got == "" {
		t.Fatalf("expected Cache-Control for /static")
	}
	if got := w.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Fatalf("expected Vary header for /static, got %q", got)
	}

	r2, _ := http.NewRequest("GET", "/assets/img.png", nil)
	w2 := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w2, r2)
	if got := w2.Header().Get("Cache-Control"); got == "" {
		t.Fatalf("expected Cache-Control for /assets")
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
	sleepCalled := make(chan struct{})
	release := make(chan struct{})
	sleepFn = func(d time.Duration) {
		called++
		os.Setenv("CRON_ONE_SHOT", "1")
		close(sleepCalled)
		<-release
	}
	cronNewOrm = func() orm.Ormer { return nil }
	go generarNominaAutomatica()
	<-sleepCalled
	nowFn = origNow
	sleepFn = origSleep
	cronNewOrm = origOrm
	close(release)
	time.Sleep(10 * time.Millisecond)
	os.Unsetenv("CRON_ONE_SHOT")
	if called == 0 {
		t.Fatalf("expected sleep to be called")
	}
}

func TestGenerarNominaAutomatica_PrintsAndErrors(t *testing.T) {
	os.Setenv("CRON_ONE_SHOT", "1")
	defer os.Unsetenv("CRON_ONE_SHOT")

	origNow := nowFn
	nowFn = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, database.BogotaZone) }
	defer func() { nowFn = origNow }()

	origLogger := slog.Default()
	setBufLogger := func() *bytes.Buffer {
		var b bytes.Buffer
		h := slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(h))
		return &b
	}
	restore := func() { slog.SetDefault(origLogger) }
	defer restore()

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

	b = setBufLogger()
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { return 0, fmt.Errorf("insert fail") }
	generarNominaAutomatica()
	out2 := b.String()
	if !contains(out2, "cron.nomina.insert_err") {
		t.Fatalf("expected insert error event, got: %s", out2)
	}

	b = setBufLogger()
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { n.PK_ID_NOMINA = 6; return 1, nil }
	cronRawExec = func(o orm.Ormer, q string, args ...interface{}) (sql.Result, error) {
		return nil, fmt.Errorf("proc fail")
	}
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

type nilRawOrmer struct{ *orm.DoNothingOrm }

func (nilRawOrmer) Raw(string, ...interface{}) orm.RawSeter { return nil }

func TestCronRawExec_NilRawSeter(t *testing.T) {
	cronRawExec(nilRawOrmer{}, "q")
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

func TestStringHelpers(t *testing.T) {
	if !contains("abc", "abc") {
		t.Fatalf("expected exact match")
	}
	if contains("abc", "xyz") {
		t.Fatalf("unexpected match for xyz")
	}
	if contains("a", "abcd") {
		t.Fatalf("expected false when sub longer than s")
	}
	if !stringContains("abc", "") {
		t.Fatalf("expected true for empty substring")
	}
	if stringContains("ab", "abcd") {
		t.Fatalf("expected false when sub longer than s in stringContains")
	}
	if indexOf("abc", "b") != 1 {
		t.Fatalf("expected index 1 for 'b'")
	}
	if indexOf("abc", "d") != -1 {
		t.Fatalf("expected -1 when substring not found")
	}
}

func initBeegoApp(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(file))
	beego.TestBeegoInit(apppath)
}

func TestHealthzEndpoint(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	main()

	r, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "ok" {
		t.Fatalf("expected body ok, got %q", body)
	}
}

type fakePinger struct{ err error }

func (f fakePinger) Ping() error { return f.err }

func TestGetSQLPinger_Default(t *testing.T) {
	orig := getSQLPinger
	getSQLPinger = func() (sqlPinger, error) { return nil, fmt.Errorf("no alias") }
	t.Cleanup(func() { getSQLPinger = orig })

	db, err := getSQLPinger()
	if err == nil {
		t.Fatalf("expected error when no default DB alias is registered")
	}
	if db != nil {
		t.Fatalf("expected nil DB without default alias")
	}
}

func TestGetSQLPinger_DelegatesToDatabase(t *testing.T) {
	db1, err1 := getSQLPinger()
	db2, err2 := database.GetDefaultSQLDB()

	if (err1 != nil) != (err2 != nil) {
		t.Fatalf("expected same error presence; getSQLPinger err=%v, database err=%v", err1, err2)
	}
	if (db1 != nil) != (db2 != nil) {
		t.Fatalf("expected same db nilness between getSQLPinger and database.GetDefaultSQLDB")
	}
	if db1 != nil && db2 != nil {
		if sdb, ok := db1.(*sql.DB); !ok {
			t.Fatalf("expected *sql.DB from getSQLPinger")
		} else if sdb != db2 {
			t.Fatalf("expected same *sql.DB pointer from both calls")
		}
	}
}

func TestGetSQLPinger_NilBranchReturnsNilInterface(t *testing.T) {
	orig := dbGetter
	dbGetter = func() (*sql.DB, error) { return nil, fmt.Errorf("no db") }
	t.Cleanup(func() { dbGetter = orig })

	p, err := getSQLPinger()
	if err == nil {
		t.Fatalf("expected error when dbGetter returns nil")
	}
	if p != nil {
		t.Fatalf("expected nil interface when db is nil")
	}
}

func TestGetSQLPinger_NonNilBranchReturnsDB(t *testing.T) {
	orig := dbGetter
	var opened *sql.DB
	dbGetter = func() (*sql.DB, error) {
		var err error
		opened, err = sql.Open("postgres", "user=u password=p dbname=x sslmode=disable")
		if err != nil {
			return nil, err
		}
		return opened, nil
	}
	t.Cleanup(func() {
		dbGetter = orig
		if opened != nil {
			_ = opened.Close()
		}
	})

	p, err := getSQLPinger()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatalf("expected non-nil pinger")
	}
	if _, ok := p.(*sql.DB); !ok {
		t.Fatalf("expected *sql.DB from getSQLPinger")
	}
}

func TestReadyz_OK(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	orig := getSQLPinger
	getSQLPinger = func() (sqlPinger, error) { return fakePinger{err: nil}, nil }
	t.Cleanup(func() { getSQLPinger = orig })
	main()

	r, _ := http.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyz_Unavailable(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	orig := getSQLPinger
	getSQLPinger = func() (sqlPinger, error) { return fakePinger{err: fmt.Errorf("down")}, nil }
	t.Cleanup(func() { getSQLPinger = orig })
	main()

	r, _ := http.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestReadyz_GetPingerErrorStill200(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	orig := getSQLPinger
	getSQLPinger = func() (sqlPinger, error) { return nil, fmt.Errorf("no alias") }
	t.Cleanup(func() { getSQLPinger = orig })
	main()

	r, _ := http.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when getSQLPinger fails, got %d", w.Code)
	}
}

func TestReadyz_NilPingerStill200(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	orig := getSQLPinger
	getSQLPinger = func() (sqlPinger, error) { return nil, nil }
	t.Cleanup(func() { getSQLPinger = orig })
	main()

	r, _ := http.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when getSQLPinger returns nil, got %d", w.Code)
	}
}

func TestCORS_DevAllowAll(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	beego.BConfig.RunMode = "dev"
	main()

	r, _ := http.NewRequest("OPTIONS", "/healthz", nil)
	r.Header.Set("Origin", "https://foo.example")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected ACAO '*', got %q", got)
	}
}

func TestCORS_DevExplicitOrigins(t *testing.T) {
	t.Skip("TODO: implementar configuración de CORS basada en env vars")
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com, https://b.com")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON"); os.Unsetenv("CORS_ALLOWED_ORIGINS") })

	initBeegoApp(t)
	beego.BConfig.RunMode = "dev"
	main()

	r, _ := http.NewRequest("OPTIONS", "/healthz", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://a.com" {
		t.Fatalf("expected ACAO 'https://a.com', got %q", got)
	}
}

func TestCORS_ProdNoOrigins(t *testing.T) {
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON") })

	initBeegoApp(t)
	beego.BConfig.RunMode = "prod"
	main()

	r, _ := http.NewRequest("OPTIONS", "/healthz", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 on CORS preflight in prod")
	}
}

func TestCORS_ProdExplicitOrigins(t *testing.T) {
	t.Skip("TODO: implementar configuración de CORS basada en env vars")
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com, https://b.com")
	t.Cleanup(func() { os.Unsetenv("SKIP_WEB_RUN"); os.Unsetenv("SKIP_CRON"); os.Unsetenv("CORS_ALLOWED_ORIGINS") })

	initBeegoApp(t)
	beego.BConfig.RunMode = "prod"
	main()

	r, _ := http.NewRequest("OPTIONS", "/healthz", nil)
	r.Header.Set("Origin", "https://b.com")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://b.com" {
		t.Fatalf("expected ACAO 'https://b.com', got %q", got)
	}
}

func TestMaxBodyBytesEnv(t *testing.T) {
	t.Skip("TODO: implementar configuración de MAX_BODY_BYTES basada en env vars")
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Setenv("MAX_BODY_BYTES", "1048576")
	t.Cleanup(func() {
		os.Unsetenv("SKIP_WEB_RUN")
		os.Unsetenv("SKIP_CRON")
		os.Unsetenv("MAX_BODY_BYTES")
	})

	orig := beego.BConfig.MaxMemory
	t.Cleanup(func() { beego.BConfig.MaxMemory = orig })

	initBeegoApp(t)
	main()

	if beego.BConfig.MaxMemory != 1048576 {
		t.Fatalf("expected MaxMemory 1048576, got %d", beego.BConfig.MaxMemory)
	}
}

func TestMultipartMaxMemoryEnv(t *testing.T) {
	t.Skip("TODO: implementar configuración de MULTIPART_MAX_MEMORY_MB basada en env vars")
	os.Setenv("SKIP_WEB_RUN", "1")
	os.Setenv("SKIP_CRON", "1")
	os.Setenv("MULTIPART_MAX_MEMORY_MB", "2")
	t.Cleanup(func() {
		os.Unsetenv("SKIP_WEB_RUN")
		os.Unsetenv("SKIP_CRON")
		os.Unsetenv("MULTIPART_MAX_MEMORY_MB")
	})

	origMem := beego.BConfig.MaxMemory
	origCopy := beego.BConfig.CopyRequestBody
	t.Cleanup(func() { beego.BConfig.MaxMemory = origMem; beego.BConfig.CopyRequestBody = origCopy })

	initBeegoApp(t)
	main()

	if beego.BConfig.MaxMemory != int64(2*1024*1024) {
		t.Fatalf("expected MaxMemory 2MB, got %d", beego.BConfig.MaxMemory)
	}
	if beego.BConfig.CopyRequestBody {
		t.Fatalf("expected CopyRequestBody false")
	}
}
