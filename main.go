package main

import (
	"restaurante/database"
	_ "restaurante/docs" // Importa la documentación generada de Swagger
	_ "restaurante/routers"

	beego "github.com/beego/beego/v2/server/web"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Inicializar la base de datos
	database.InitDB()

	// Habilitar Swagger en modo de desarrollo
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		// Configurar ruta para Swagger UI usando beego.Handler()
		beego.Handler("/swagger/*", httpSwagger.WrapHandler)
	}

	// Ejecutar la aplicación Beego
	beego.Run()
}
