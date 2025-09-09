package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"restaurante/database"
	_ "restaurante/docs"
	"restaurante/logging"
	"restaurante/models"
	_ "restaurante/routers"
	"strconv"
	"strings"
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
			slog.Info("cron.nomina.start")

			nomina := models.Nomina{
				FECHA:         now,
				ESTADO_NOMINA: models.EstadoNominaNoPago,
				MONTO:         0,
			}
			if _, err := cronInsertNom(o, &nomina); err != nil {
				slog.Error("cron.nomina.insert_err", slog.String("error", err.Error()))
			} else {
				if _, err := cronRawExec(o, "CALL generar_nomina_automatica(?, ?)", nomina.PK_ID_NOMINA, nomina.FECHA); err != nil {
					slog.Error("cron.nomina.exec_err", slog.String("error", err.Error()))
				} else {
					slog.Info("cron.nomina.success")
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
	// Solo aplicar cache a archivos estáticos (imagenes, CSS, JS)
	url := ctx.Input.URL()
	if strings.HasPrefix(url, "/assets/") || strings.HasPrefix(url, "/static/") {
		ctx.Output.Header("Cache-Control", "public, max-age=31536000, immutable")
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
	// Logging
	runMode := web.BConfig.RunMode
	logging.Setup(runMode)
	slog.Info("app.start", slog.String("runmode", runMode))

	// Middleware de logging: inicio/fin de request + errores
	web.InsertFilter("*", web.BeforeRouter, logging.StartTimer)
	web.InsertFilter("*", web.FinishRouter, logging.LogRequest)
	web.InsertFilter("*", web.FinishRouter, logging.LogIfError)

	// Healthcheck
	web.Get("/healthz", func(ctx *context.Context) {
		ctx.Output.SetStatus(200)
		_ = ctx.Output.Body([]byte("ok"))
	})
	web.Get("/readyz", func(ctx *context.Context) {
		status := 200
		if sqlDB, err := database.GetDefaultSQLDB(); err == nil && sqlDB != nil {
			if err := sqlDB.Ping(); err != nil {
				status = http.StatusServiceUnavailable
			}
		}
		ctx.Output.SetStatus(status)
		_ = ctx.Output.Body([]byte(http.StatusText(status)))
	})

	// Limitar tamaño de request y multipart
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			web.BConfig.MaxMemory = n
		}
	}
	if v := os.Getenv("MULTIPART_MAX_MEMORY_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			web.BConfig.CopyRequestBody = false
			web.BConfig.MaxMemory = int64(n) * 1024 * 1024
		}
	}

	// Configurar CORS según entorno
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS") // Coma-separado
	corsOpts := &cors.Options{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	if runMode == "prod" {
		// En prod, exigir orígenes explícitos; si no hay, bloquear cross-origin
		if strings.TrimSpace(allowedOriginsEnv) == "" {
			corsOpts.AllowAllOrigins = false
			corsOpts.AllowOrigins = []string{}
		} else {
			parts := strings.Split(allowedOriginsEnv, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			corsOpts.AllowAllOrigins = false
			corsOpts.AllowOrigins = parts
		}
	} else {
		// En dev/test: si se define CORS_ALLOWED_ORIGINS, respétalo con credenciales.
		// Si no se define, permite todos los orígenes pero SIN credenciales para evitar bloqueo del navegador.
		if strings.TrimSpace(allowedOriginsEnv) == "" {
			corsOpts.AllowAllOrigins = true
			corsOpts.AllowCredentials = false
		} else {
			parts := strings.Split(allowedOriginsEnv, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			corsOpts.AllowAllOrigins = false
			corsOpts.AllowOrigins = parts
			corsOpts.AllowCredentials = true
		}
	}
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(corsOpts))

	// Aplicar cache en estáticos
	web.InsertFilter("/*", web.BeforeRouter, setStaticHeaders)

	// Swagger y listado de directorios sólo fuera de prod
	if runMode != "prod" {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.Handler("/swagger/*", httpSwagger.WrapHandler)
	}

	// Iniciar el cron job solo si la DB está disponible y no se ha pedido omitirlo
	if dbReady && os.Getenv("SKIP_CRON") != "1" {
		go generarNominaAutomatica()
	}

	// Iniciar el servidor salvo que se pida omitir (para tests/unit)
	if os.Getenv("SKIP_WEB_RUN") != "1" {
		webRun()
	}
}
