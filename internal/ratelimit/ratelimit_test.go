package ratelimit

import (
	"testing"
	"time"
)

func TestAllowLimitaYReinicia(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(2, time.Minute, 10)
	l.now = func() time.Time { return now }
	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("las primeras peticiones deben pasar")
	}
	if l.Allow("a") {
		t.Fatal("la tercera debe bloquearse")
	}
	if !l.Allow("b") {
		t.Fatal("otra clave tiene su propia ventana")
	}
	now = now.Add(2 * time.Minute)
	if !l.Allow("a") {
		t.Fatal("tras vencer la ventana vuelve a permitir")
	}
}

func TestAllowPurgaVencidas(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(1, time.Minute, 2)
	l.now = func() time.Time { return now }
	l.Allow("a")
	l.Allow("b")
	now = now.Add(2 * time.Minute)
	l.Allow("c") // mapa lleno: purga a y b
	if len(l.m) != 1 {
		t.Fatalf("esperaba 1 entrada tras la purga, hay %d", len(l.m))
	}
	// con el mapa lleno pero sin vencidas, no se purga nada
	l.Allow("d")
	l.Allow("e")
	if len(l.m) != 3 {
		t.Fatalf("no debía purgar ventanas vigentes: %d", len(l.m))
	}
}

func TestEnvInt(t *testing.T) {
	t.Setenv("RL_TEST", "")
	if EnvInt("RL_TEST", 7) != 7 {
		t.Fatal("vacío usa el defecto")
	}
	t.Setenv("RL_TEST", "x")
	if EnvInt("RL_TEST", 7) != 7 {
		t.Fatal("inválido usa el defecto")
	}
	t.Setenv("RL_TEST", "0")
	if EnvInt("RL_TEST", 7) != 7 {
		t.Fatal("no positivo usa el defecto")
	}
	t.Setenv("RL_TEST", "3")
	if EnvInt("RL_TEST", 7) != 3 {
		t.Fatal("válido se respeta")
	}
}
