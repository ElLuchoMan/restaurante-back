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

func InitDB() error {
	quiet := os.Getenv("QUIET_TESTS") == "1"

	dbHost, _ := web.AppConfig.String("db_host")
	dbPort, _ := web.AppConfig.String("db_port")
	dbUser, _ := web.AppConfig.String("db_user")
	dbPass, _ := web.AppConfig.String("db_pass")
	dbName, _ := web.AppConfig.String("db_name")

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
	if _, err := orm.GetDB("default"); err == nil {
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
	o := orm.NewOrm()

	defaults := []models.MetodoPago{
		{TIPO: "Efectivo"},
		{TIPO: "Tarjeta"},
	}

	for _, m := range defaults {
		cnt, err := o.QueryTable(new(models.MetodoPago)).Filter("TIPO", m.TIPO).Count()
		if err != nil {
			return err
		}
		if cnt == 0 {
			if _, err := o.Insert(&m); err != nil {
				return err
			}
		}
	}

	return nil
}
