package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

var registerDataBase = orm.RegisterDataBase

var loadLocation = time.LoadLocation

var newOrmForSeed = orm.NewOrm

var getDB = orm.GetDB

var appConfigString = func(key string) (string, error) { return web.AppConfig.String(key) }

var (
	queryTableFn = func(o orm.Ormer, model interface{}) orm.QuerySeter { return o.QueryTable(model) }
	filterFn     = func(qs orm.QuerySeter, expr string, args ...interface{}) orm.QuerySeter {
		return qs.Filter(expr, args...)
	}
	countFn  = func(qs orm.QuerySeter) (int64, error) { return qs.Count() }
	insertFn = func(o orm.Ormer, model interface{}) (int64, error) { return o.Insert(model) }
)

var countMetodoPagoByTipo = func(o orm.Ormer, tipo string) (int64, error) {
	qs := queryTableFn(o, new(models.MetodoPago))
	qs = filterFn(qs, "TIPO", tipo)
	return countFn(qs)
}

func GetDefaultSQLDB() (*sql.DB, error) {
	return getDB("default")
}

func getenvInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func InitDB() error {
	quiet := os.Getenv("QUIET_TESTS") == "1"

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		v, err := appConfigString("db_host")
		if err != nil {
			return fmt.Errorf("config db_host: %w", err)
		}
		dbHost = v
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		v, err := appConfigString("db_port")
		if err != nil {
			return fmt.Errorf("config db_port: %w", err)
		}
		dbPort = v
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		v, err := appConfigString("db_user")
		if err != nil {
			return fmt.Errorf("config db_user: %w", err)
		}
		dbUser = v
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		v, err := appConfigString("db_pass")
		if err != nil {
			return fmt.Errorf("config db_pass: %w", err)
		}
		dbPass = v
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		v, err := appConfigString("db_name")
		if err != nil {
			return fmt.Errorf("config db_name: %w", err)
		}
		dbName = v
	}
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s TimeZone=UTC",
		dbUser, dbPass, dbHost, dbPort, dbName, sslMode)

	if err := registerDataBase("default", "postgres", connStr); err != nil {
		return err
	}

	if !quiet {
		fmt.Println("Conexión a la base de datos exitosa!")
		fmt.Println("Conectando a PostgreSQL en:", dbHost, "Puerto:", dbPort, "Base de datos:", dbName)
	}

	if sqlDB, err := GetDefaultSQLDB(); err == nil && sqlDB != nil {
		if maxOpen := getenvInt("DB_MAX_OPEN"); maxOpen > 0 {
			sqlDB.SetMaxOpenConns(maxOpen)
		}
		if maxIdle := getenvInt("DB_MAX_IDLE"); maxIdle > 0 {
			sqlDB.SetMaxIdleConns(maxIdle)
		}
		if mins := getenvInt("DB_CONN_MAX_LIFETIME_MIN"); mins > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(mins) * time.Minute)
		}
		if mins := getenvInt("DB_CONN_MAX_IDLE_TIME_MIN"); mins > 0 {
			sqlDB.SetConnMaxIdleTime(time.Duration(mins) * time.Minute)
		}
	}

	if os.Getenv("SKIP_DB_SEED") == "1" {
		return nil
	}

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
