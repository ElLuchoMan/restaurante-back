package database

import (
	"fmt"
	"log"

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

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=America/Bogota",
		dbUser, dbPass, dbHost, dbPort, dbName)

	err := orm.RegisterDataBase("default", "postgres", connStr)
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}

	fmt.Println("Conexión a la base de datos exitosa!")
}
