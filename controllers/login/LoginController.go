package login

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"restaurante/models"

	"github.com/golang-jwt/jwt/v5"
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

type Claims struct {
	Documento int64  `json:"documento"`
	Rol       string `json:"rol"`
	Nombre    string `json:"nombre"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

var signingMethod jwt.SigningMethod = jwt.SigningMethodHS256

func init() {
	jwtSecret = loadJWTSecret()
}

func loadJWTSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	if isTestingProcess() || web.BConfig.RunMode != "prod" {
		b := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, b); err == nil {
			return b
		}
		return []byte("dev-insecure-default")
	}
	panic("JWT_SECRET no configurado")
}

func isTestingProcess() bool {
	if len(os.Args) > 0 {
		exe := strings.ToLower(os.Args[0])
		if strings.HasSuffix(exe, ".test") || strings.HasSuffix(exe, ".test.exe") {
			return true
		}
	}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

var (
	loginRL     = newRateLimiter()
	loginMaxReq = getEnvIntDefault("LOGIN_MAX_REQ_PER_MIN", 10)
	loginWindow = time.Minute
	rlMutex     sync.Mutex
)

type rateEntry struct {
	count int
	reset time.Time
}

type rateLimiter struct {
	m map[string]*rateEntry
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{m: make(map[string]*rateEntry)}
}

func getEnvIntDefault(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	return d
}

func clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func allowLogin(r *http.Request) bool {
	rlMutex.Lock()
	defer rlMutex.Unlock()
	ip := clientIP(r)
	entry, ok := loginRL.m[ip]
	if !ok || time.Now().After(entry.reset) {
		loginRL.m[ip] = &rateEntry{count: 1, reset: time.Now().Add(loginWindow)}
		return true
	}
	if entry.count >= loginMaxReq {
		return false
	}
	entry.count++
	return true
}

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
// @Failure 429 {object} models.ApiResponse "Demasiadas solicitudes"
// @Router /login [post]
func (c *LoginController) Login() {
	if !allowLogin(c.Ctx.Request) {
		c.Ctx.Output.SetStatus(http.StatusTooManyRequests)
		c.Data["json"] = models.ApiResponse{Code: http.StatusTooManyRequests, Message: "Demasiados intentos, intente más tarde"}
		_ = c.ServeJSON()
		return
	}

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

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: loginRequest.Documento}
	err := o.Read(&trabajador)

	if err == nil {
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
		generateJWT(c, trabajador.PK_DOCUMENTO_TRABAJADOR, string(trabajador.ROL), nombre)
		return
	}

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: loginRequest.Documento}
	err = o.Read(&cliente)

	if err == nil {
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
		generateJWT(c, cliente.PK_DOCUMENTO_CLIENTE, "Cliente", nombre)
		return
	}

	c.Ctx.Output.SetStatus(http.StatusUnauthorized)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusUnauthorized,
		Message: "Credenciales inválidas",
	}
	_ = c.ServeJSON()
}

func generateJWT(c *LoginController, documento int64, rol string, nombre string) {
	if len(jwtSecret) == 0 {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al generar el token",
			Cause:   fmt.Errorf("secreto JWT no configurado").Error(),
		}
		_ = c.ServeJSON()
		return
	}
	now := time.Now()
	expirationTime := now.Add(24 * time.Hour)

	claims := &Claims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
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
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.Status = http.StatusOK
		return
	}

	method := ctx.Input.Method()
	path := ctx.Input.URL()
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}

	if method == http.MethodGet || method == http.MethodPost {
		switch path {
		case "/restaurante/v1/productos", "/restaurante/v1/productos/search",
			"/restaurante/v1/restaurantes", "/restaurante/v1/restaurantes/search",
			"/restaurante/v1/reservas", "/restaurante/v1/reservas/search",
			"/restaurante/v1/reservas/parameter",
			"/restaurante/v1/cambios_horario/actual":
			return
		}
	}

	if method == "POST" && path == "/restaurante/v1/clientes" {
		return
	}

	if web.BConfig.RunMode == "dev" {
		referer := ctx.Input.Header("Referer")
		if strings.Contains(referer, "/swagger/") {
			return
		}
		if strings.HasPrefix(path, "/swagger/") {
			return
		}
	}

	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		if err := ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token no proporcionado",
		}, false, false); err != nil {
		}
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		if err := ctx.Output.JSON(models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido",
		}, false, false); err != nil {
		}
		return
	}
}
