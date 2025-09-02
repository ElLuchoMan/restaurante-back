package test

import (
	"testing"
	"time"

	"restaurante/models"
)

func TestNominaTriggerGeneratesMonto(t *testing.T) {
	o, err := getOrmer()
	if err != nil {
		t.Fatalf("orm not available: %v", err)
	}
	trabajador := models.Trabajador{
		PK_DOCUMENTO_TRABAJADOR: 9999,
		NOMBRE:                  "Test",
		APELLIDO:                "User",
		SUELDO:                  1000,
		NUEVO:                   true,
		ROL:                     models.RolCocinero,
		FECHA_INGRESO:           time.Now(),
		PASSWORD:                "pwd",
	}
	if _, err := o.Insert(&trabajador); err != nil {
		t.Fatalf("insert trabajador failed: %v", err)
	}
	defer o.Delete(&trabajador)

	nom := models.Nomina{FECHA: time.Now(), ESTADO_NOMINA: models.EstadoNominaNoPago}
	if _, err := o.Insert(&nom); err != nil {
		t.Fatalf("insert nomina failed: %v", err)
	}
	defer o.Delete(&nom)

	if _, err := o.Raw("CALL generar_nomina_automatica(?, ?)", nom.PK_ID_NOMINA, nom.FECHA).Exec(); err != nil {
		t.Fatalf("call generar_nomina_automatica failed: %v", err)
	}
	if _, err := o.Raw("CALL verificar_nomina()").Exec(); err != nil {
		t.Fatalf("call verificar_nomina failed: %v", err)
	}
	if err := o.Read(&nom); err != nil {
		t.Fatalf("read nomina failed: %v", err)
	}
	if nom.MONTO == 0 {
		t.Errorf("expected monto to be set by trigger, got 0")
	}
}

func TestNominaTriggerExists(t *testing.T) {
	o, err := getOrmer()
	if err != nil {
		t.Fatalf("orm not available: %v", err)
	}
	var count int
	if err := o.Raw("SELECT COUNT(*) FROM pg_trigger WHERE tgrelid = 'nomina'::regclass").QueryRow(&count); err != nil {
		t.Fatalf("cannot query trigger existence: %v", err)
	}
	if count == 0 {
		t.Fatalf("no triggers found for nomina table")
	}
}
