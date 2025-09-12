package database

import (
	"testing"
)

func TestGetenvInt_Empty(t *testing.T) {
	if v := getenvInt("NON_EXISTENT_ENV_INT"); v != 0 {
		t.Fatalf("expected 0, got %d", v)
	}
}

func TestGetenvInt_ParseOK(t *testing.T) {
	t.Setenv("X_INT", "42")
	if v := getenvInt("X_INT"); v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestGetenvInt_NegativeAndInvalid(t *testing.T) {
	t.Setenv("X_NEG", "-5")
	if v := getenvInt("X_NEG"); v != 0 {
		t.Fatalf("expected 0 for negative, got %d", v)
	}
	t.Setenv("X_BAD", "bad")
	if v := getenvInt("X_BAD"); v != 0 {
		t.Fatalf("expected 0 for invalid, got %d", v)
	}
}
