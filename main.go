package main

import (
	"log"
	"restaurante/database"
	_ "restaurante/docs"
	_ "restaurante/routers"

	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Declaración global de la variable location
var location *time.Location

func init() {
	// Inicializar la base de datos
	database.InitDB()

	// Cargar el timezone
	var err error
	location, err = time.LoadLocation("America/Bogota")
	if err != nil {
		log.Println("Advertencia: Error al cargar el timezone 'America/Bogota'. Usando manual offset.")
		location = time.FixedZone("America/Bogota", -5*60*60) // UTC -5
	}
}

// @title Restaurante API
// @version 2.0.0
// @description API para gestionar el sistema de un restaurante para "El fogón de María"
// @contact.email baluisto96@gmail.com
// @basePath /restaurante/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @Security BearerAuth
func main() {
	// Habilitar CORS para todas las rutas
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Habilitar la documentación de Swagger
	web.BConfig.WebConfig.DirectoryIndex = true
	web.Handler("/swagger/*", httpSwagger.WrapHandler)

	// Iniciar el servidor
	web.Run()
}
