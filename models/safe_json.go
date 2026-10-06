package models

import (
	"bytes"
	"encoding/json"
)

// marshalSeguro serializa v y elimina cualquier clave "password" (a cualquier
// profundidad), de modo que una relación embebida (p. ej. un Cliente cargado
// con RelatedSel) nunca filtre contraseñas en una respuesta.
func marshalSeguro(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return quitarPasswords(b), nil
}

// quitarPasswords devuelve b sin claves "password". Si b no es JSON válido se
// devuelve tal cual.
func quitarPasswords(b []byte) []byte {
	var g any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&g); err != nil {
		return b
	}
	// Un valor compuesto solo de tipos JSON siempre se puede volver a serializar.
	out, _ := json.Marshal(sinPassword(g))
	return out
}

func sinPassword(v any) any {
	switch t := v.(type) {
	case map[string]any:
		delete(t, "password")
		for k, e := range t {
			t[k] = sinPassword(e)
		}
		return t
	case []any:
		for i, e := range t {
			t[i] = sinPassword(e)
		}
		return t
	}
	return v
}
