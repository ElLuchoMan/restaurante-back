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

var (
	initDBFunc       = database.InitDB
	initTimezoneFunc = database.InitTimezone
)

func init() {
	appInit()
}

func appInit() {
	if err := initDBFunc(); err != nil {
		log.Println("Error al conectar a la base de datos:", err)
	}
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

	// Iniciar cron job si no está deshabilitado
	if os.Getenv("SKIP_CRON") != "1" {
		go generarNominaAutomatica()
	}

	// Ejecutar aplicación
	web.Run()
}
