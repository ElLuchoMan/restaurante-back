package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	dbHost, _ := web.AppConfig.String("db_host")
	dbPort, _ := web.AppConfig.String("db_port")
	dbUser, _ := web.AppConfig.String("db_user")
	dbPass, _ := web.AppConfig.String("db_pass")
	dbName, _ := web.AppConfig.String("db_name")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}

	fmt.Println("Conexión a la base de datos exitosa!")
}
