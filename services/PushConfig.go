package services

import (
	"os"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

// configValor lee primero la variable de entorno y, si está vacía, la clave de
// app.conf; devuelve "" cuando ninguna está definida.
func configValor(env, key string) string {
	if v := strings.TrimSpace(os.Getenv(env)); v != "" {
		return v
	}
	v, _ := web.AppConfig.String(key)
	return strings.TrimSpace(v)
}

// PushConfigurado indica si el servidor puede enviar push: claves VAPID
// (VAPID_PUBLIC_KEY y VAPID_PRIVATE_KEY) para Web Push o FIREBASE_PROJECT_ID
// para FCM. Los envíos automáticos lo consultan para degradar en silencio
// cuando no hay nada configurado.
func PushConfigurado() bool {
	vapid := configValor("VAPID_PUBLIC_KEY", "vapid_public_key") != "" &&
		configValor("VAPID_PRIVATE_KEY", "vapid_private_key") != ""
	return vapid || configValor("FIREBASE_PROJECT_ID", "firebase_project_id") != ""
}
