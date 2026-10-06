package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Las listas de las respuestas de telemetría y push se serializan como [] y
// nunca como null.
func TestListasNilSeSerializanComoArregloVacio(t *testing.T) {
	casos := map[string]any{
		"SalesData":            SalesData{},
		"ProductsData":         ProductsData{},
		"UsersData":            UsersData{},
		"TimeAnalysisData":     TimeAnalysisData{},
		"RentabilidadData":     RentabilidadData{},
		"SegmentacionData":     SegmentacionData{},
		"EficienciaData":       EficienciaData{},
		"ReservasAnalisisData": ReservasAnalisisData{},
		"PedidosAnalisisData":  PedidosAnalisisData{},
		"PushDispositivo":      PushDispositivo{},
	}
	for nombre, v := range casos {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%s: %v", nombre, err)
		}
		if strings.Contains(string(b), "null") {
			t.Errorf("%s contiene null: %s", nombre, b)
		}
	}
}

func TestListasConDatosSeConservan(t *testing.T) {
	b, _ := json.Marshal(SalesData{TendenciaVentas: []VentaPorFecha{{Fecha: "01-01-2026", Total: 5}}})
	if !strings.Contains(string(b), `"tendenciaVentas":[{"fecha":"01-01-2026","total":5,"cantidad":0}]`) || !strings.Contains(string(b), `"ventasPorMetodoPago":[]`) {
		t.Errorf("serialización inesperada: %s", b)
	}
}

func TestPushDispositivoMarshalJSON_SinPasswordYConIDs(t *testing.T) {
	ahora := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	d := PushDispositivo{
		PkIdPushDispositivo:   1,
		Plataforma:            PlataformaWeb,
		PkDocumentoCliente:    &Cliente{PK_DOCUMENTO_CLIENTE: 10, PASSWORD: "x"},
		PkDocumentoTrabajador: &Trabajador{PK_DOCUMENTO_TRABAJADOR: 20, PASSWORD: "x"},
		CreatedAt:             ahora,
		LastSeenAt:            &ahora,
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["documentoCliente"] != float64(10) || m["documentoTrabajador"] != float64(20) {
		t.Errorf("documentos inesperados: %s", b)
	}
	if m["lastSeenAt"] == nil || strings.Contains(strings.ToLower(string(b)), "password") {
		t.Errorf("respuesta inesperada: %s", b)
	}
}

func TestPushEnvioMarshalJSON_IDNumerico(t *testing.T) {
	e := PushEnvio{PkIdPushEnvio: 2, PkIdPushDispositivo: &PushDispositivo{PkIdPushDispositivo: 9, PkDocumentoCliente: &Cliente{PASSWORD: "x"}}}
	b, _ := json.Marshal(e)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["pushDispositivoId"] != float64(9) || strings.Contains(strings.ToLower(string(b)), "password") {
		t.Errorf("respuesta inesperada: %s", b)
	}

	b, _ = json.Marshal(PushEnvio{})
	_ = json.Unmarshal(b, &m)
	if m["pushDispositivoId"] != float64(0) {
		t.Errorf("sin dispositivo debe ser 0: %s", b)
	}
}
