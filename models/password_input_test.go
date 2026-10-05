package models

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cliente y Trabajador aceptan `password` como entrada pero nunca lo serializan,
// ni solos ni embebidos en otras estructuras.
func TestClienteTrabajador_PasswordSoloEntrada(t *testing.T) {
	var c Cliente
	if err := json.Unmarshal([]byte(`{"documentoCliente":1,"nombre":"Ana","password":"secreta"}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.PASSWORD != "secreta" || c.PK_DOCUMENTO_CLIENTE != 1 || c.NOMBRE != "Ana" {
		t.Fatalf("entrada no aplicada: %+v", c)
	}
	// sin password en el JSON no se pisa el valor existente
	if err := json.Unmarshal([]byte(`{"nombre":"Bea"}`), &c); err != nil || c.PASSWORD != "secreta" || c.NOMBRE != "Bea" {
		t.Fatalf("merge incorrecto: %+v %v", c, err)
	}
	if err := json.Unmarshal([]byte(`{"nombre":5}`), &c); err == nil {
		t.Fatal("tipo inválido debía fallar")
	}
	b, _ := json.Marshal(c)
	embebido, _ := json.Marshal(map[string]any{"cliente": &c, "lista": []Cliente{c}})
	for _, out := range []string{string(b), string(embebido)} {
		if strings.Contains(strings.ToLower(out), "password") || strings.Contains(out, "secreta") {
			t.Fatalf("el cliente serializó la contraseña: %s", out)
		}
	}

	var tr Trabajador
	if err := json.Unmarshal([]byte(`{"documentoTrabajador":2,"nombre":"Luis","rol":"Mesero","password":"clave"}`), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.PASSWORD != "clave" || tr.PK_DOCUMENTO_TRABAJADOR != 2 || tr.ROL != RolMesero {
		t.Fatalf("entrada no aplicada: %+v", tr)
	}
	if err := json.Unmarshal([]byte(`{"nombre":"Pedro"}`), &tr); err != nil || tr.PASSWORD != "clave" || tr.NOMBRE != "Pedro" {
		t.Fatalf("merge incorrecto: %+v %v", tr, err)
	}
	if err := json.Unmarshal([]byte(`{"sueldo":"x"}`), &tr); err == nil {
		t.Fatal("tipo inválido debía fallar")
	}
	b, _ = json.Marshal(tr)
	embebido, _ = json.Marshal(map[string]any{"trabajador": &tr, "lista": []Trabajador{tr}})
	for _, out := range []string{string(b), string(embebido)} {
		if strings.Contains(strings.ToLower(out), "password") || strings.Contains(out, "clave") {
			t.Fatalf("el trabajador serializó la contraseña: %s", out)
		}
	}
}

func TestTrabajador_MarshalHorarios(t *testing.T) {
	// horarios presentes (aun vacíos) se serializan; ausentes se omiten
	con, _ := json.Marshal(Trabajador{PK_DOCUMENTO_TRABAJADOR: 1, HORARIOS: []HorarioTrabajador{}})
	sin, _ := json.Marshal(Trabajador{PK_DOCUMENTO_TRABAJADOR: 1})
	if !strings.Contains(string(con), `"horarios":[]`) || strings.Contains(string(sin), "horarios") {
		t.Fatalf("horarios mal serializados: %s / %s", con, sin)
	}
}

func TestMarshalSeguro(t *testing.T) {
	// elimina `password` a cualquier profundidad, también dentro de listas
	b, err := marshalSeguro(map[string]any{
		"password": "x",
		"n":        map[string]any{"password": "y", "ok": 1},
		"l":        []any{map[string]any{"password": "z", "ok": 2}, 3},
	})
	if err != nil || strings.Contains(string(b), "password") || !strings.Contains(string(b), `"ok":2`) {
		t.Fatalf("password sin eliminar: %s %v", b, err)
	}
	// valor no serializable
	if _, err := marshalSeguro(make(chan int)); err == nil {
		t.Fatal("un canal no se puede serializar")
	}
	// JSON inválido se devuelve tal cual
	if got := string(quitarPasswords([]byte(`{no json`))); got != `{no json` {
		t.Fatalf("JSON inválido debe devolverse intacto: %s", got)
	}
}
