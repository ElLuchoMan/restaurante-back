package main

import (
	"restaurante/database"
	_ "restaurante/docs" // Importa la documentación generada de Swagger
	_ "restaurante/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// Inicializar la base de datos
	database.InitDB()

	// Habilitar Swagger en modo de desarrollo
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.Run()
}
