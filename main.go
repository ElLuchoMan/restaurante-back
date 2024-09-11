package main

import (
	"restaurante/database"
	_ "restaurante/docs"
	_ "restaurante/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// Inicializar la base de datos
	database.InitDB()

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
