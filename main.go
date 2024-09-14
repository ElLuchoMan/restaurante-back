package main

import (
	"fmt"
	_ "restaurante/docs"
	_ "restaurante/routers"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	dbUser, _ := beego.AppConfig.String("db_user")
	dbPass, _ := beego.AppConfig.String("db_pass")
	dbHost, _ := beego.AppConfig.String("db_host")
	dbPort, _ := beego.AppConfig.String("db_port")
	dbName, _ := beego.AppConfig.String("db_name")

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	err := orm.RegisterDataBase("default", "postgres", connStr)
	if err != nil {
		panic(err)
	}
}

// @title Restaurante API
// @version 0.0.1
// @description API para gestionar el sistema de un restaurante para "El fogón de María"
// @contact.email baluisto96@gmail.com
// @host restaurante-back-production.up.railway.app
// @basePath /restaurante/v1
// @schemes https
func main() {
	// Habilitar CORS para todas las rutas
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Habilitar la documentación de Swagger
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.Handler("/swagger/*", httpSwagger.WrapHandler)

	// Iniciar el servidor
	beego.Run()
}
