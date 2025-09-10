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
	_ "github.com/lib/pq"
)

var dbReady bool

// sqlPinger permite testear /readyz sin una *sql.DB real
type sqlPinger interface {
	Ping() error
}

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
	// getSQLPinger permite stubear el acceso a la DB en /readyz
	getSQLPinger = func() (sqlPinger, error) { return database.GetDefaultSQLDB() }
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

// Implementación real (reasignable en tests) para permitir aislar cobertura
var setStaticHeadersFn = func(ctx *context.Context) {
	// Solo aplicar cache a archivos estáticos (imagenes, CSS, JS)
	url := ctx.Input.URL()
	if strings.HasPrefix(url, "/assets/") || strings.HasPrefix(url, "/static/") {
		ctx.Output.Header("Cache-Control", "public, max-age=31536000, immutable")
		ctx.Output.Header("Vary", "Accept-Encoding")
	}
}

func setStaticHeaders(ctx *context.Context) { setStaticHeadersFn(ctx) }

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
	setupAndRun()
}
