package models

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func marshalToMap(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestMarshalJSON_ResponsesAndModels(t *testing.T) {
	ts := time.Date(2025, 3, 10, 15, 4, 5, 0, time.UTC) // 10:04:05 Bogota
	hora := time.Date(2000, 1, 1, 8, 30, 0, 0, time.UTC)
	horaFin := time.Date(2000, 1, 1, 18, 45, 10, 0, time.UTC)
	endpoint := "ep"
	code := "E1"
	status := 200

	m := marshalToMap(t, PushDispositivoResponse{PushDispositivoId: 3, Endpoint: &endpoint, CreatedAt: ts, LastSeenAt: &ts})
	if m["createdAt"] != "10-03-2025 10:04:05" || m["lastSeenAt"] != "10-03-2025 10:04:05" || m["endpoint"] != "ep" {
		t.Fatalf("PushDispositivoResponse: %v", m)
	}
	m = marshalToMap(t, PushDispositivoResponse{CreatedAt: ts})
	if _, ok := m["lastSeenAt"]; ok {
		t.Fatalf("lastSeenAt should be omitted: %v", m)
	}

	m = marshalToMap(t, PushEnvioResponse{PushEnvioId: 1, StatusCode: &status, ErrorCode: &code, SentAt: ts})
	if m["sentAt"] != "10-03-2025 10:04:05" || m["errorCode"] != "E1" || m["statusCode"] != float64(200) {
		t.Fatalf("PushEnvioResponse: %v", m)
	}

	m = marshalToMap(t, CuponResponse{CuponId: 5, Codigo: "X", FechaInicio: ts, FechaFin: ts.AddDate(0, 0, 1)})
	if m["fechaInicio"] != "10-03-2025" || m["fechaFin"] != "11-03-2025" || m["codigo"] != "X" {
		t.Fatalf("CuponResponse: %v", m)
	}

	m = marshalToMap(t, OfertaResponse{OfertaId: 2, FechaInicio: ts, FechaFin: ts, HoraInicio: &hora, HoraFin: &horaFin})
	if m["horaInicio"] != "08:30:00" || m["horaFin"] != "18:45:10" || m["fechaInicio"] != "10-03-2025" {
		t.Fatalf("OfertaResponse: %v", m)
	}
	m = marshalToMap(t, OfertaResponse{FechaInicio: ts, FechaFin: ts})
	if _, ok := m["horaInicio"]; ok {
		t.Fatalf("horaInicio should be omitted: %v", m)
	}

	m = marshalToMap(t, ControlNomina{PK_ID_CONTROL_NOMINA: 4, Fecha: ts, Estado: "ABIERTO"})
	if m["controlNominaId"] != float64(4) || m["fecha"] != "10-03-2025" || m["estado"] != "ABIERTO" {
		t.Fatalf("ControlNomina: %v", m)
	}

	m = marshalToMap(t, PrecioProductoHist{PK_ID_PRECIO_HIST: 9, Precio: 1500, FechaVigencia: ts, PKIDProducto: &Producto{}})
	if m["precioHistId"] != float64(9) || m["precio"] != float64(1500) || m["fechaVigencia"] != "10-03-2025" || m["productoId"] == nil {
		t.Fatalf("PrecioProductoHist: %v", m)
	}

	m = marshalToMap(t, Oferta{PkIdOferta: 1, Titulo: "T", FechaInicio: ts, FechaFin: ts, HoraInicio: &hora, HoraFin: &horaFin, DiasSemanaArray: []string{"LUNES"}})
	if m["horaInicio"] != "08:30:00" || m["horaFin"] != "18:45:10" || m["fechaFin"] != "10-03-2025" {
		t.Fatalf("Oferta: %v", m)
	}
	m = marshalToMap(t, Oferta{FechaInicio: ts, FechaFin: ts})
	if _, ok := m["horaInicio"]; ok {
		t.Fatalf("Oferta horaInicio should be omitted: %v", m)
	}

	m = marshalToMap(t, PushDispositivo{PkIdPushDispositivo: 8, CreatedAt: ts, LastSeenAt: &ts, SubscribedTopicsArray: []string{"a"}})
	if m["lastSeenAt"] != "10-03-2025 10:04:05" || m["createdAt"] != "10-03-2025 10:04:05" {
		t.Fatalf("PushDispositivo: %v", m)
	}
	m = marshalToMap(t, PushDispositivo{CreatedAt: ts})
	if _, ok := m["lastSeenAt"]; ok {
		t.Fatalf("PushDispositivo lastSeenAt should be omitted: %v", m)
	}

	apertura := hora
	m = marshalToMap(t, CambiosHorario{FECHA: ts, HORA_APERTURA: &apertura, HORA_CIERRE: horaFin})
	if m["horaApertura"] != "08:30:00" {
		t.Fatalf("CambiosHorario: %v", m)
	}
	m = marshalToMap(t, CambiosHorario{FECHA: ts, HORA_CIERRE: horaFin})
	if _, ok := m["horaApertura"]; ok {
		t.Fatalf("CambiosHorario apertura should be omitted: %v", m)
	}

	m = marshalToMap(t, Incidencia{PK_ID_INCIDENCIA: 1, FECHA: ts, PK_DOCUMENTO_TRABAJADOR: &Trabajador{PK_DOCUMENTO_TRABAJADOR: 77}})
	if m["documentoTrabajador"] != float64(77) {
		t.Fatalf("Incidencia: %v", m)
	}
	m = marshalToMap(t, Incidencia{FECHA: ts})
	if m["documentoTrabajador"] != float64(0) {
		t.Fatalf("Incidencia nil: %v", m)
	}
}

func TestDeserializeSubscribedTopics_BracesAndQuotes(t *testing.T) {
	p := &PushDispositivo{SubscribedTopics: `{"a""b", c}`}
	p.deserializeSubscribedTopics()
	if len(p.SubscribedTopicsArray) != 2 || p.SubscribedTopicsArray[0] != `a"b` || p.SubscribedTopicsArray[1] != "c" {
		t.Fatalf("got %v", p.SubscribedTopicsArray)
	}
	p = &PushDispositivo{SubscribedTopics: `["x","y"]`}
	p.deserializeSubscribedTopics()
	if len(p.SubscribedTopicsArray) != 2 || p.SubscribedTopicsArray[1] != "y" {
		t.Fatalf("got %v", p.SubscribedTopicsArray)
	}
	p = &PushDispositivo{SubscribedTopics: "{}"}
	p.deserializeSubscribedTopics()
	if p.SubscribedTopicsArray == nil || len(p.SubscribedTopicsArray) != 0 {
		t.Fatalf("got %#v", p.SubscribedTopicsArray)
	}
}

func TestFormatTimestampBogota_LocationFallback(t *testing.T) {
	orig := loadLocation
	defer func() { loadLocation = orig }()
	loadLocation = func(string) (*time.Location, error) { return nil, errors.New("no tzdata") }
	ts := time.Date(2025, 3, 10, 15, 4, 5, 0, time.UTC)
	if got := FormatTimestampBogota(ts); got != "10-03-2025 10:04:05" {
		t.Fatalf("got %s", got)
	}
}

func TestParseTimeToUTC_ShortAndInvalid(t *testing.T) {
	got, err := ParseTimeToUTC("08:30")
	if err != nil || got.Hour() != 8 || got.Minute() != 30 || got.Second() != 0 {
		t.Fatalf("got %v err %v", got, err)
	}
	if _, err := ParseTimeToUTC("25:00"); err == nil {
		t.Fatalf("expected error")
	}
}
