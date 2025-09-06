package main

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

type fakeResult struct{}
func (fakeResult) LastInsertId() (int64, error) { return 1, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

func TestGenerarNominaAutomatica_OneShot(t *testing.T) {
	// Forzar que sea medianoche
	origNow := nowFn
	origSleep := sleepFn
	origInsert := cronInsertNom
	origExec := cronRawExec
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	sleepFn = func(d time.Duration) {}
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { return 1, nil }
	cronRawExec = func(o orm.Ormer, query string, args ...interface{}) (sql.Result, error) { return fakeResult{}, nil }
	t.Cleanup(func(){ nowFn = origNow; sleepFn = origSleep; cronInsertNom = origInsert; cronRawExec = origExec })

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
	nowFn = func() time.Time { return time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC) }
	sleepFn = func(d time.Duration) {}
	called := 0
	cronInsertNom = func(o orm.Ormer, n *models.Nomina) (int64, error) { called++; return 0, nil }
	t.Cleanup(func(){ nowFn = origNow; sleepFn = origSleep; cronInsertNom = origInsert })

	os.Setenv("CRON_ONE_SHOT", "1")
	defer os.Unsetenv("CRON_ONE_SHOT")

	go generarNominaAutomatica()
	time.Sleep(10 * time.Millisecond)
	if called != 0 {
		t.Fatalf("expected no insert when not midnight, got %d", called)
	}
}
