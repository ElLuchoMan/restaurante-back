package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"restaurante/models"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

type LoginController struct {
	web.Controller
}

// Estructura para los claims del JWT
type Claims struct {
	Documento int    `json:"documento"`
	Rol       string `json:"rol"`
	jwt.StandardClaims
}

// Llave secreta para firmar el token
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()

	// Primero, intenta encontrar al usuario como trabajador
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(loginRequest.Documento)}
	err := o.Read(&trabajador)

	if err == nil {
		// Verificar la contraseña
		if err := bcrypt.CompareHashAndPassword([]byte(trabajador.PASSWORD), []byte(loginRequest.Password)); err != nil {
			c.Ctx.Output.SetStatus(http.StatusUnauthorized)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "Credenciales inválidas",
			}
			c.ServeJSON()
			return
		}

		// Generar JWT con el rol específico del trabajador (admin, mesero, mensajero, etc.)
		generateJWT(c, int(trabajador.PK_DOCUMENTO_TRABAJADOR), trabajador.ROL)
		return
	}

	// Si no es un trabajador, intenta como cliente
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: loginRequest.Documento}
	err = o.Read(&cliente)

	if err == nil {
		// Verificar la contraseña
		if err := bcrypt.CompareHashAndPassword([]byte(cliente.PASSWORD), []byte(loginRequest.Password)); err != nil {
			c.Ctx.Output.SetStatus(http.StatusUnauthorized)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "Credenciales inválidas",
			}
			c.ServeJSON()
			return
		}

		// Generar JWT con rol de "cliente"
		generateJWT(c, cliente.PK_DOCUMENTO_CLIENTE, "cliente")
		return
	}

	// Si no se encontró ni como trabajador ni como cliente
	c.Ctx.Output.SetStatus(http.StatusUnauthorized)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusUnauthorized,
		Message: "Credenciales inválidas",
	}
	c.ServeJSON()
}

// Función para generar y devolver un token JWT
func generateJWT(c *LoginController, documento int, rol string) {
	// Obtener la fecha y hora actual
	now := time.Now()

	// Establecer la hora de expiración a las 11:59 p.m. del mismo día
	expirationTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())

	// Crear los claims con la expiración calculada
	claims := &Claims{
		Documento: documento,
		Rol:       rol,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Generar el token con los claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al generar el token",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Respuesta exitosa con el token
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Inicio de sesión exitoso",
		Data: map[string]string{
			"token": tokenString,
		},
	}
	c.ServeJSON()
}

func ValidateToken(ctx *context.Context) {
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token no proporcionado",
		}, false, false)
		return
	}

	// Verificar si ya contiene el prefijo 'Bearer'
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}

	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido",
		}, false, false)
		return
	}

	// Se podría usar el rol aquí para autorizaciones más avanzadas
}
