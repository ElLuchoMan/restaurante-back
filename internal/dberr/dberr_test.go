package dberr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestIsUnique(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"pq code", &pq.Error{Code: "23505"}, true},
		{"pq wrapped", fmt.Errorf("x: %w", &pq.Error{Code: "23505"}), true},
		{"pq other code", &pq.Error{Code: "22001", Message: "boom"}, false},
		{"text duplicate", errors.New("ERROR: duplicate key value"), true},
		{"text unique", errors.New("UNIQUE constraint failed"), true},
		{"text code", errors.New("sqlstate 23505"), true},
		{"other", errors.New("boom"), false},
	}
	for _, c := range cases {
		if got := IsUnique(c.err); got != c.want {
			t.Errorf("%s: IsUnique=%v, want %v", c.name, got, c.want)
		}
	}
}

func TestIsForeignKey(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"pq code", &pq.Error{Code: "23503"}, true},
		{"pq other code", &pq.Error{Code: "22001", Message: "boom"}, false},
		{"text fk", errors.New("violates foreign key constraint"), true},
		{"text code", errors.New("sqlstate 23503"), true},
		{"other", errors.New("boom"), false},
	}
	for _, c := range cases {
		if got := IsForeignKey(c.err); got != c.want {
			t.Errorf("%s: IsForeignKey=%v, want %v", c.name, got, c.want)
		}
	}
}

func TestMentions(t *testing.T) {
	if Mentions(nil, "x") {
		t.Fatal("nil no menciona")
	}
	err := errors.New(`duplicate key value violates unique constraint "UQ_Cliente_Correo"`)
	if !Mentions(err, "correo") || Mentions(err, "telefono") {
		t.Fatal("Mentions incorrecto")
	}
}
