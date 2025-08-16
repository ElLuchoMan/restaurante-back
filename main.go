package main

import (
	"fmt"
	"log"
	"restaurante/database"
	_ "restaurante/docs"
	_ "restaurante/routers"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

var dbReady bool

func init() {
	// Inicializar la base de datos y la zona horaria
	if err := database.InitDB(); err != nil {
		log.Println("Error al conectar a la base de datos:", err)
		dbReady = false
	} else {
		dbReady = true
	}
	database.InitTimezone()
	fmt.Println("Loaded timezone:", database.BogotaZone)
}

// Función para ejecutar la generación automática de nómina
func generarNominaAutomatica() {
	o := orm.NewOrm() // Usar la conexión existente

	for {
		// Ejecutar la función de nómina cada día a las 00:00
		now := time.Now().In(database.BogotaZone)
		if now.Hour() == 0 && now.Minute() == 0 {
			fmt.Println("Ejecutando generación automática de nómina...")

			// Llamar a la función de nómina
			_, err := o.Raw("CALL generar_nomina_automatica()").Exec()
			if err != nil {
				fmt.Println("Error al generar la nómina automática:", err)
			} else {
				fmt.Println("Nómina generada automáticamente con éxito.")
			}
		}

		// Esperar 1 minuto antes de verificar de nuevo
		time.Sleep(1 * time.Minute)
	}
}

func setStaticHeaders(ctx *context.Context) {
	// Solo aplicar cache y compresión en archivos estáticos (imagenes, CSS, JS)
	if ctx.Input.URL() != "" && (ctx.Input.URL() == "/assets/" || ctx.Input.URL() == "/static/") {
		ctx.Output.Header("Cache-Control", "public, max-age=31536000, immutable")
		ctx.Output.Header("Content-Encoding", "gzip")
		ctx.Output.Header("Vary", "Accept-Encoding")
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
	// Aplicar compresión y cacheo SOLO a archivos estáticos
	web.InsertFilter("/*", web.BeforeRouter, setStaticHeaders)

	// Habilitar la documentación de Swagger
	web.BConfig.WebConfig.DirectoryIndex = true
	web.Handler("/swagger/*", httpSwagger.WrapHandler)

	// Iniciar el cron job en un goroutine solo si la DB está disponible
	if dbReady {
		go generarNominaAutomatica()
	}

	// Iniciar el servidor
	web.Run()
}
