package login

import (
	"net/http/httptest"
	"testing"
	"time"
)

// Al alcanzar loginMaxEntries se descartan las ventanas vencidas y se conservan las vigentes.
func TestAllowLoginPurgaVentanasVencidas(t *testing.T) {
	origRL, origMax := loginRL, loginMaxEntries
	loginRL = newRateLimiter()
	loginMaxEntries = 2
	t.Cleanup(func() { loginRL, loginMaxEntries = origRL, origMax })

	loginRL.m["vencida"] = &rateEntry{count: 1, reset: time.Now().Add(-time.Minute)}
	loginRL.m["vigente"] = &rateEntry{count: 1, reset: time.Now().Add(time.Minute)}

	r := httptest.NewRequest("POST", "/login", nil)
	r.RemoteAddr = "198.51.100.7:4000"
	if !allowLogin(r) {
		t.Fatal("el primer intento debe permitirse")
	}
	if _, ok := loginRL.m["vencida"]; ok {
		t.Fatal("la ventana vencida debía purgarse")
	}
	if _, ok := loginRL.m["vigente"]; !ok {
		t.Fatal("la ventana vigente debía conservarse")
	}
	if len(loginRL.m) != 2 {
		t.Fatalf("esperaba 2 entradas, hay %d", len(loginRL.m))
	}
}
