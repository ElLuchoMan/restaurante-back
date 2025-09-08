package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"restaurante/models"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

var (
	newOrm                 = orm.NewOrm
	compareHashAndPassword = bcrypt.CompareHashAndPassword
)

type LoginController struct {
	web.Controller
}

// Estructura para los claims del JWT
type Claims struct {
	Documento int64  `json:"documento"`
	Rol       string `json:"rol"`
	Nombre    string `json:"nombre"`
	jwt.StandardClaims
}

// Llave secreta para firmar el token
var jwtSecret = []byte(os.Getenv("cocina-de-maria"))

// @Title Login
// @Summary Iniciar sesión para clientes o trabajadores
// @Description Permite iniciar sesión utilizando el documento y la contraseña, devuelve un JWT con el rol.
// @Tags login
// @Accept json
// @Produce json
// @Param   body  body   models.LoginRequest  true  "Documento y Contraseña"
// @Success 200 {object} models.ApiResponse "Inicio de sesión exitoso con token JWT"
// @Failure 400 {object} models.ApiResponse "Solicitud incorrecta"
// @Failure 401 {object} models.ApiResponse "Credenciales inválidas"
// @Router /login [post]
func (c *LoginController) Login() {
	var loginRequest models.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &loginRequest); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	o := newOrm()

	// Primero, intenta encontrar al usuario como trabajador
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: loginRequest.Documento}
	err := o.Read(&trabajador)

	if err == nil {
		// Verificar la contraseña
		if err := compareHashAndPassword([]byte(trabajador.PASSWORD), []byte(loginRequest.Password)); err != nil {
			c.Ctx.Output.SetStatus(http.StatusUnauthorized)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "Credenciales inválidas",
			}
			_ = c.ServeJSON()
			return
		}
		nombre := trabajador.NOMBRE + " " + trabajador.APELLIDO
		// Generar JWT con el rol específico del trabajador (admin, mesero, mensajero, etc.)
		generateJWT(c, trabajador.PK_DOCUMENTO_TRABAJADOR, string(trabajador.ROL), nombre)
		return
	}

	// Si no es un trabajador, intenta como cliente
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: loginRequest.Documento}
	err = o.Read(&cliente)

	if err == nil {
		// Verificar la contraseña
		if err := compareHashAndPassword([]byte(cliente.PASSWORD), []byte(loginRequest.Password)); err != nil {
			c.Ctx.Output.SetStatus(http.StatusUnauthorized)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "Credenciales inválidas",
			}
			_ = c.ServeJSON()
			return
		}
		nombre := cliente.NOMBRE + " " + cliente.APELLIDO
		// Generar JWT con rol de "Cliente"
		generateJWT(c, cliente.PK_DOCUMENTO_CLIENTE, "Cliente", nombre)
		return
	}

	// Si no se encontró ni como trabajador ni como cliente
	c.Ctx.Output.SetStatus(http.StatusUnauthorized)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusUnauthorized,
		Message: "Credenciales inválidas",
	}
	_ = c.ServeJSON()
}

// Función para generar y devolver un token JWT
func generateJWT(c *LoginController, documento int64, rol string, nombre string) {
	now := time.Now()
	expirationTime := now.Add(24 * time.Hour)

	claims := &Claims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al generar el token",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Inicio de sesión exitoso",
		Data: map[string]string{
			"token":  tokenString,
			"nombre": nombre,
		},
	}
	_ = c.ServeJSON()
}

func ValidateToken(ctx *context.Context) {
	// 1) Permitir CORS preflight
	if ctx.Input.Method() == "OPTIONS" {
		fmt.Println("Solicitud OPTIONS recibida y permitida")
		ctx.Output.Status = http.StatusOK
		return
	}

	// 2) Permitir crear cliente sin token (registro público)
	if ctx.Input.Method() == "POST" && ctx.Input.URL() == "/restaurante/v1/clientes" {
		fmt.Println("POST público /restaurante/v1/clientes: sin validación de token")
		return
	}

	// 2b) En desarrollo, permitir solicitudes iniciadas desde Swagger UI sin token
	if web.BConfig.RunMode == "dev" {
		referer := ctx.Input.Header("Referer")
		if strings.Contains(referer, "/swagger/") {
			fmt.Println("Bypass de token por Swagger UI en modo dev")
			return
		}
		// Por seguridad adicional, si se pidiera a rutas de swagger (estático)
		if strings.HasPrefix(ctx.Input.URL(), "/swagger/") {
			return
		}
	}

	// 3) Resto de rutas: exigir token
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		fmt.Println("No se proporcionó el token")
		ctx.Output.SetStatus(http.StatusUnauthorized)
		if err := ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token no proporcionado",
		}, false, false); err != nil {
			// no hay otro canal aquí; simplemente no propagamos
		}
		return
	}
	fmt.Println("Token recibido:", authHeader)

	// Normalizar prefijo Bearer
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		fmt.Println("Token inválido:", err)
		ctx.Output.SetStatus(http.StatusUnauthorized)
		if err := ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido",
		}, false, false); err != nil {
			// noop si falla escribir la respuesta
		}
		return
	}

	fmt.Println("Token válido. Claims:", claims)
}
