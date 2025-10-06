package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"restaurante/database"
	_ "restaurante/docs"
	"restaurante/models"
	_ "restaurante/routers"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	initDBFunc       = database.InitDB
	initTimezoneFunc = database.InitTimezone
	dbReady          = false //nolint:unused // variable usada en tests
)

func init() {
	loadEnvFile()
	appInit()
}

// loadEnvFile carga variables de entorno desde .env si existe (solo para desarrollo local)
func loadEnvFile() {
	// Intentar cargar .env solo si no estamos en CI
	if os.Getenv("CI") != "true" && os.Getenv("SKIP_WEB_RUN") != "1" {
		if err := godotenv.Load(); err != nil {
			// No es error crítico si no existe .env, puede usar variables del sistema o config file
			log.Printf("Info: no se encontró archivo .env, usando variables de entorno del sistema o configuración: %v\n", err)
		} else {
			log.Println("Variables de entorno cargadas desde .env")
		}
	}
}

func appInit() {
	if err := initDBFunc(); err != nil {
		log.Println("Error al conectar a la base de datos:", err)
		dbReady = false
		return
	}
	dbReady = true
	initTimezoneFunc()
}

var (
	nowFn         = time.Now
	sleepFn       = time.Sleep
	cronNewOrm    = orm.NewOrm
	cronInsertNom = ormInsert
	cronRawExec   = ormRawExec
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

func generarNominaAutomatica() {
	o := cronNewOrm()

	for {
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

		if os.Getenv("CRON_ONE_SHOT") == "1" {
			return
		}

		sleepFn(1 * time.Minute)
	}
}

var setStaticHeadersFn = func(ctx *context.Context) {
	url := ctx.Input.URL()
	if strings.HasPrefix(url, "/assets/") || strings.HasPrefix(url, "/static/") {
		ctx.Output.Header("Cache-Control", "public, max-age=31536000, immutable")
		ctx.Output.Header("Vary", "Accept-Encoding")
	}
}

func setStaticHeaders(ctx *context.Context) { setStaticHeadersFn(ctx) }

var webRun = web.Run

// Variables para health checks (inyección de dependencias para testing)
type sqlPinger interface {
	Ping() error
}

var dbGetter = database.GetDefaultSQLDB

var getSQLPinger = func() (sqlPinger, error) {
	db, err := dbGetter()
	if err != nil || db == nil {
		return nil, err
	}
	return db, nil
}

// @title El fogón de María API
// @version 2.0.0
// @description API para gestionar el sistema de "El fogón de María"
// @contact.email baluisto96@gmail.com
// @basePath /restaurante/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @Security BearerAuth
func main() {
	// Configurar timezone
	database.InitTimezone()

	// Configurar headers estáticos
	web.InsertFilter("*", web.BeforeStatic, setStaticHeaders)

	// Registrar health check endpoints
	web.Router("/healthz", &HealthController{}, "get:Healthz")
	web.Router("/readyz", &HealthController{}, "get:Readyz")

	// Iniciar cron job si no está deshabilitado
	if os.Getenv("SKIP_CRON") != "1" {
		go generarNominaAutomatica()
	}

	// Ejecutar aplicación si no está deshabilitado
	if os.Getenv("SKIP_WEB_RUN") != "1" {
		webRun()
	}
}

// HealthController maneja los endpoints de health check
type HealthController struct {
	web.Controller
}

// Healthz verifica disponibilidad básica de la aplicación
// @Summary Verifica la salud básica de la API
// @Description Retorna 200 OK si la aplicación está en ejecución
// @Tags health
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /healthz [get]
func (c *HealthController) Healthz() {
	c.Ctx.WriteString("ok")
}

// Readyz verifica si la aplicación está lista para recibir tráfico
// @Summary Verifica si la aplicación está lista
// @Description Retorna 200 OK si la aplicación puede conectarse a la base de datos
// @Tags health
// @Produce plain
// @Success 200 {string} string "ok"
// @Failure 503 {string} string "unavailable"
// @Router /readyz [get]
func (c *HealthController) Readyz() {
	db, err := getSQLPinger()
	if err != nil || db == nil {
		c.Ctx.WriteString("ok")
		return
	}

	if err := db.Ping(); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(503)
		c.Ctx.WriteString("unavailable")
		return
	}

	c.Ctx.WriteString("ok")
}
