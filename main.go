package main

import (
	"fmt"
	"log"
	_ "restaurante/docs"
	_ "restaurante/routers"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	dbUser, _ := beego.AppConfig.String("db_user")
	dbPass, _ := beego.AppConfig.String("db_pass")
	dbHost, _ := beego.AppConfig.String("db_host")
	dbPort, _ := beego.AppConfig.String("db_port")
	dbName, _ := beego.AppConfig.String("db_name")

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	err := orm.RegisterDataBase("default", "postgres", connStr)
	if err != nil {
		panic(err)
	}
}

// @title Restaurante API
// @version 0.0.1
// @description API para gestionar el sistema de un restaurante para "El fogón de María"
// @contact.email baluisto96@gmail.com
// @basePath /restaurante/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @Security BearerAuth
func main() {
	// Habilitar CORS para todas las rutas
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Habilitar la documentación de Swagger
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.Handler("/swagger/*", httpSwagger.WrapHandler)

	// Iniciar el servidor
	go func() {
		err := sendWhatsAppMessage("573042449339@s.whatsapp.net", "Hola, este es un mensaje desde go-whatsapp!")
		if err != nil {
			log.Fatalf("Error enviando mensaje de WhatsApp: %v", err)
		}
	}()

	beego.Run()
}

func sendWhatsAppMessage(recipient string, text string) error {
	// Crear una nueva instancia de WhatsApp
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		return fmt.Errorf("error creando conexión de WhatsApp: %v", err)
	}

	// Escanear el código QR para iniciar sesión en WhatsApp Web
	qr := make(chan string)
	go func() {
		fmt.Printf("Escanea este código QR en tu WhatsApp Web:\n%s\n", <-qr)
	}()

	session, err := wac.Login(qr)
	if err != nil {
		return fmt.Errorf("error iniciando sesión en WhatsApp: %v", err)
	}

	fmt.Println("Inicio de sesión exitoso", session.Wid)

	// Enviar el mensaje
	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: recipient,
		},
		Text: text,
	}

	_, err = wac.Send(msg)
	if err != nil {
		return fmt.Errorf("error enviando mensaje: %v", err)
	}

	return nil
}
