package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

// registerDataBase allows tests to stub orm.RegisterDataBase.
var registerDataBase = orm.RegisterDataBase

// loadLocation allows tests to stub time.LoadLocation.
var loadLocation = time.LoadLocation

// newOrmForSeed allows tests to stub orm.NewOrm when seeding.
var newOrmForSeed = orm.NewOrm

// getDB allows tests to stub orm.GetDB.
var getDB = orm.GetDB

// Lightweight indirections to allow unit testing seed without a real DB.
var (
	queryTableFn = func(o orm.Ormer, model interface{}) orm.QuerySeter { return o.QueryTable(model) }
	filterFn     = func(qs orm.QuerySeter, expr string, args ...interface{}) orm.QuerySeter {
		return qs.Filter(expr, args...)
	}
	countFn  = func(qs orm.QuerySeter) (int64, error) { return qs.Count() }
	insertFn = func(o orm.Ormer, model interface{}) (int64, error) { return o.Insert(model) }
)

// countMetodoPagoByTipo encapsula la consulta usada por el seed para poder ser stubbeada en tests.
var countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) {
	qs := queryTableFn(o, new(models.MetodoPago))
	qs = filterFn(qs, "TIPO", tipo)
	return countFn(qs)
}

func InitDB() error {
	quiet := os.Getenv("QUIET_TESTS") == "1"

	// Validar y obtener configuración obligatoria de DB
	dbHost, err := web.AppConfig.String("db_host")
	if err != nil {
		return fmt.Errorf("config db_host: %w", err)
	}
	dbPort, err := web.AppConfig.String("db_port")
	if err != nil {
		return fmt.Errorf("config db_port: %w", err)
	}
	dbUser, err := web.AppConfig.String("db_user")
	if err != nil {
		return fmt.Errorf("config db_user: %w", err)
	}
	dbPass, err := web.AppConfig.String("db_pass")
	if err != nil {
		return fmt.Errorf("config db_pass: %w", err)
	}
	dbName, err := web.AppConfig.String("db_name")
	if err != nil {
		return fmt.Errorf("config db_name: %w", err)
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=UTC",
		dbUser, dbPass, dbHost, dbPort, dbName)

	if err := registerDataBase("default", "postgres", connStr); err != nil {
		return err
	}

	if !quiet {
		fmt.Println("Conexión a la base de datos exitosa!")
		fmt.Println("Conectando a PostgreSQL en:", dbHost, "Puerto:", dbPort, "Base de datos:", dbName)
	}

	// Permitir desactivar el seed en unit tests para evitar depender del alias registrado
	if os.Getenv("SKIP_DB_SEED") == "1" {
		return nil
	}

	// Ejecutar seed solo si el alias realmente existe en el ORM
	if _, err := getDB("default"); err == nil {
		if err := seedMetodoPago(); err != nil {
			if !quiet {
				fmt.Println("Error al poblar METODO_PAGO:", err)
			}
		}
	}

	return nil
}

var BogotaZone *time.Location

func InitTimezone() {
	var err error
	BogotaZone, err = loadLocation("America/Bogota")
	if err != nil {
		if os.Getenv("QUIET_TESTS") != "1" {
			log.Println("Advertencia: Error al cargar el timezone 'America/Bogota'. Usando UTC.")
		}
		BogotaZone = time.FixedZone("UTC-5", -5*60*60)
	}
}

func seedMetodoPago() error {
	o := newOrmForSeed()

	defaults := []models.MetodoPago{
		{TIPO: "Efectivo"},
		{TIPO: "Tarjeta"},
	}

	for _, m := range defaults {
		cnt, err := countMetodoPagoByTipo(o, m.TIPO)
		if err != nil {
			return err
		}
		if cnt == 0 {
			if _, err := insertFn(o, &m); err != nil {
				return err
			}
		}
	}

	return nil
}
