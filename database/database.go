package database

import (
	"fmt"
	"log"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

func InitDB() {
	dbHost, _ := web.AppConfig.String("db_host")
	dbPort, _ := web.AppConfig.String("db_port")
	dbUser, _ := web.AppConfig.String("db_user")
	dbPass, _ := web.AppConfig.String("db_pass")
	dbName, _ := web.AppConfig.String("db_name")

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=UTC",
		dbUser, dbPass, dbHost, dbPort, dbName)

	err := orm.RegisterDataBase("default", "postgres", connStr)

	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}

	fmt.Println("Conexión a la base de datos exitosa!")
	fmt.Println("Conectando a PostgreSQL en:", dbHost, "Puerto:", dbPort, "Base de datos:", dbName)

}

var BogotaZone *time.Location

func InitTimezone() {
	var err error
	BogotaZone, err = time.LoadLocation("America/Bogota")
	if err != nil {
		log.Println("Advertencia: Error al cargar el timezone 'America/Bogota'. Usando UTC.")
		BogotaZone = time.FixedZone("UTC-5", -5*60*60)
	}
}
