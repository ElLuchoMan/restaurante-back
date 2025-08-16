package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestRestauranteSetAndGetDiasLaborales(t *testing.T) {
	r := Restaurante{}
	dias := []string{"Lunes", "Martes"}
	if err := r.SetDiasLaborales(dias); err != nil {
		t.Fatalf("SetDiasLaborales returned error: %v", err)
	}
	// Verify internal JSON representation
	want, _ := json.Marshal(dias)
	if r.DIAS_LABORALES != string(want) {
		t.Errorf("expected DIAS_LABORALES %s, got %s", string(want), r.DIAS_LABORALES)
	}
	// Verify GetDiasLaborales returns original slice
	got, err := r.GetDiasLaborales()
	if err != nil {
		t.Fatalf("GetDiasLaborales returned error: %v", err)
	}
	if !reflect.DeepEqual(got, dias) {
		t.Errorf("expected %v, got %v", dias, got)
	}
}

func TestRestauranteGetDiasLaboralesInvalidJSON(t *testing.T) {
	r := Restaurante{DIAS_LABORALES: "not-json"}
	if _, err := r.GetDiasLaborales(); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

func TestRestauranteSetDiasLaboralesError(t *testing.T) {
	original := jsonMarshal
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal error")
	}
	defer func() { jsonMarshal = original }()

	r := Restaurante{}
	if err := r.SetDiasLaborales([]string{"Lunes"}); err == nil {
		t.Errorf("expected error but got nil")
	}
}
