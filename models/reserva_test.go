package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestReservaMarshalJSON(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.September, 12, 0, 0, 0, 0, loc)
	estado := EstadoReservaConfirmada
	r := Reserva{FECHA: fecha, ESTADO_RESERVA: &estado}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaReserva"] != "12-09-2024" {
		t.Errorf("expected fechaReserva 12-09-2024, got %v", data["fechaReserva"])
	}
	if data["estadoReserva"] != string(EstadoReservaConfirmada) {
		t.Errorf("expected estadoReserva %s, got %v", EstadoReservaConfirmada, data["estadoReserva"])
	}
}

func TestReservaMarshalJSONCreatedUpdatedBy(t *testing.T) {
	r := Reserva{}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if _, ok := data["createdBy"]; ok {
		t.Errorf("createdBy should be omitted when nil")
	}
	if _, ok := data["updatedBy"]; ok {
		t.Errorf("updatedBy should be omitted when nil")
	}

	cb, ub := "creator", "updater"
	r.CREATED_BY = &cb
	r.UPDATED_BY = &ub

	b, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	data = map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["createdBy"] != cb {
		t.Errorf("expected createdBy %s, got %v", cb, data["createdBy"])
	}
	if data["updatedBy"] != ub {
		t.Errorf("expected updatedBy %s, got %v", ub, data["updatedBy"])
	}
}

func TestReservaTableName(t *testing.T) {
	r := Reserva{}
	if r.TableName() != "reserva" {
		t.Errorf("expected table name reserva, got %s", r.TableName())
	}
}

func TestReservaResponseConRelaciones(t *testing.T) {
	estado := EstadoReservaPendiente
	ind := "ventana"
	fecha := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	hora := time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC)
	cli := &Cliente{PK_DOCUMENTO_CLIENTE: 7, PASSWORD: "secreto"}
	r := Reserva{
		PK_ID_RESERVA:     4,
		FECHA:             fecha,
		HORA:              hora,
		PERSONAS:          3,
		ESTADO_RESERVA:    &estado,
		INDICACIONES:      &ind,
		PK_ID_CONTACTO:    &ReservaContacto{PKIDContacto: 2, NombreCompleto: "Ana", PKDocumentoCliente: cli},
		PK_ID_RESTAURANTE: &Restaurante{PK_ID_RESTAURANTE: 1, NOMBRE_RESTAURANTE: "Sazón", HORA_APERTURA: hora},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "secreto") || strings.Contains(string(b), "password") {
		t.Fatalf("filtración de contraseña: %s", b)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["fechaReserva"] != "31-01-2025" || got["horaReserva"] != "18:30:00" || got["estadoReserva"] != "PENDIENTE" || got["indicaciones"] != "ventana" {
		t.Fatalf("campos inesperados: %s", b)
	}
	c := got["contactoId"].(map[string]any)
	if c["nombreCompleto"] != "Ana" || c["documentoCliente"].(map[string]any)["documentoCliente"] != float64(7) {
		t.Fatalf("contacto inesperado: %v", c)
	}
	rest := got["restauranteId"].(map[string]any)
	if rest["restauranteId"] != float64(1) || rest["nombreRestaurante"] != "Sazón" || rest["horaApertura"] != "18:30:00" {
		t.Fatalf("restaurante inesperado: %v", rest)
	}
	if _, ok := rest["cambioHorarioId"]; ok {
		t.Fatalf("cambioHorarioId no debe viajar en la reserva: %v", rest)
	}
}

func TestReservaResponseSinRelaciones(t *testing.T) {
	got := (Reserva{}).Response()
	if got.ContactoID != nil || got.RestauranteID != nil || got.EstadoReserva != nil {
		t.Fatalf("esperaba relaciones nulas: %+v", got)
	}
}
