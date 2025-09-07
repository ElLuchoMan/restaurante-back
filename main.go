package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"restaurante/database"
	_ "restaurante/docs"
	"restaurante/models"
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

// Funciones variables para testear init sin efectos secundarios complejos
var (
	initDBFunc       = database.InitDB
	initTimezoneFunc = database.InitTimezone
)

func init() {
	appInit()
}

func appInit() {
	// Inicializar la base de datos y la zona horaria
	if err := initDBFunc(); err != nil {
		log.Println("Error al conectar a la base de datos:", err)
		dbReady = false
	} else {
		dbReady = true
	}
	initTimezoneFunc()
	fmt.Println("Loaded timezone:", database.BogotaZone)
}

// Funciones variables para facilitar pruebas del cron
var (
	nowFn         = time.Now
	sleepFn       = time.Sleep
	cronNewOrm    = orm.NewOrm
	cronInsertNom = ormInsert
	cronRawExec   = ormRawExec
	webRun        = web.Run
)

func ormInsert(o orm.Ormer, n *models.Nomina) (int64, error) {
	if o == nil {
		return 0, nil
	}
	return o.Insert(n)
}

func ormRawExec(o orm.Ormer, query string, args ...interface{}) (sql.Result, error) {
	if o == nil {
		return nil, nil
	}
	rs := o.Raw(query, args...)
	if rs == nil {
		return nil, nil
	}
	return rs.Exec()
}

// Función para ejecutar la generación automática de nómina
func generarNominaAutomatica() {
	o := cronNewOrm() // Usar la conexión existente

	for {
		// Ejecutar la función de nómina cada día a las 00:00
		now := nowFn().In(database.BogotaZone)
		if now.Hour() == 0 && now.Minute() == 0 {
			fmt.Println("Ejecutando generación automática de nómina...")

			nomina := models.Nomina{
				FECHA:         now,
				ESTADO_NOMINA: models.EstadoNominaNoPago,
				MONTO:         0,
			}
			if _, err := cronInsertNom(o, &nomina); err != nil {
				fmt.Println("Error al crear la nómina:", err)
			} else {
				if _, err := cronRawExec(o, "CALL generar_nomina_automatica(?, ?)", nomina.PK_ID_NOMINA, nomina.FECHA); err != nil {
					fmt.Println("Error al generar la nómina automática:", err)
				} else {
					fmt.Println("Nómina generada automáticamente con éxito.")
				}
			}
		}

		// Modo de una sola iteración para pruebas
		if os.Getenv("CRON_ONE_SHOT") == "1" {
			return
		}

		// Esperar 1 minuto antes de verificar de nuevo
		sleepFn(1 * time.Minute)
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

	// Iniciar el cron job solo si la DB está disponible y no se ha pedido omitirlo
	if dbReady && os.Getenv("SKIP_CRON") != "1" {
		go generarNominaAutomatica()
	}

	// Iniciar el servidor salvo que se pida omitir (para tests/unit)
	if os.Getenv("SKIP_WEB_RUN") != "1" {
		webRun()
	}
}
