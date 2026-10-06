package services

import (
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestPushConfigurado(t *testing.T) {
	vars := []string{"VAPID_PUBLIC_KEY", "VAPID_PRIVATE_KEY", "FIREBASE_PROJECT_ID"}
	clear := func() {
		for _, v := range vars {
			t.Setenv(v, "")
		}
	}

	clear()
	if PushConfigurado() {
		t.Fatal("sin variables no debe estar configurado")
	}

	clear()
	t.Setenv("VAPID_PUBLIC_KEY", "pub")
	if PushConfigurado() {
		t.Fatal("VAPID incompleto (solo pública) no basta")
	}
	t.Setenv("VAPID_PRIVATE_KEY", "  priv ")
	if !PushConfigurado() {
		t.Fatal("VAPID completo debe contar")
	}

	clear()
	t.Setenv("FIREBASE_PROJECT_ID", "proyecto")
	if !PushConfigurado() {
		t.Fatal("FIREBASE_PROJECT_ID debe contar")
	}

	clear()
	t.Setenv("FIREBASE_PROJECT_ID", "   ")
	if PushConfigurado() {
		t.Fatal("valor en blanco no cuenta")
	}

	// app.conf cuando no hay variable de entorno
	clear()
	_ = web.AppConfig.Set("firebase_project_id", " desde-conf ")
	t.Cleanup(func() { _ = web.AppConfig.Set("firebase_project_id", "") })
	if !PushConfigurado() {
		t.Fatal("firebase_project_id de app.conf debe contar")
	}
}
