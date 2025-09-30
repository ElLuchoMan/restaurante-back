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

type RefreshClaims struct {
	Documento int64  `json:"documento"`
	Rol       string `json:"rol"`
	Nombre    string `json:"nombre"`
	TokenType string `json:"token_type"` // "refresh"
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
// @Success 200 {object} models.ApiResponse{data=models.AuthResponse} "Inicio de sesión exitoso con tokens JWT"
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

func generateTokens(documento int64, rol, nombre string) (string, string, error) {
	if len(jwtSecret) == 0 {
		return "", "", fmt.Errorf("secreto JWT no configurado")
	}

	now := time.Now()

	// Access Token (30 minutos)
	accessClaims := &Claims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Refresh Token (30 días)
	refreshClaims := &RefreshClaims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(signingMethod, accessClaims)
	refreshToken := jwt.NewWithClaims(signingMethod, refreshClaims)

	accessString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("error al generar access token: %w", err)
	}

	refreshString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("error al generar refresh token: %w", err)
	}

	return accessString, refreshString, nil
}

func generateJWT(c *LoginController, documento int64, rol string, nombre string) {
	accessToken, refreshToken, err := generateTokens(documento, rol, nombre)
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
			"token":         accessToken, // Compatibilidad hacia atrás
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    "1800", // 30 minutos en segundos
			"nombre":        nombre,
		},
	}
	_ = c.ServeJSON()
}

// @Title RefreshToken
// @Summary Renovar access token usando refresh token
// @Description Permite obtener un nuevo access token utilizando un refresh token válido
// @Tags auth
// @Accept json
// @Produce json
// @Param   Authorization  header  string  true  "Refresh Token en formato: Bearer {token}"
// @Success 200 {object} models.ApiResponse{data=models.AuthResponse} "Tokens renovados exitosamente"
// @Failure 400 {object} models.ApiResponse "Solicitud incorrecta"
// @Failure 401 {object} models.ApiResponse "Refresh token inválido o expirado"
// @Router /auth/refresh [post]
func (c *LoginController) RefreshToken() {
	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Refresh token no proporcionado",
		}
		_ = c.ServeJSON()
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	tokenString := authHeader[len("Bearer "):]

	refreshClaims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Refresh token inválido o expirado",
		}
		_ = c.ServeJSON()
		return
	}

	// Verificar que sea un refresh token
	if refreshClaims.TokenType != "refresh" {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido: no es un refresh token",
		}
		_ = c.ServeJSON()
		return
	}

	// Generar nuevos tokens
	accessToken, newRefreshToken, err := generateTokens(refreshClaims.Documento, refreshClaims.Rol, refreshClaims.Nombre)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al generar nuevos tokens",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Tokens renovados exitosamente",
		Data: map[string]string{
			"token":         accessToken, // Compatibilidad hacia atrás
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"token_type":    "Bearer",
			"expires_in":    "1800", // 30 minutos en segundos
			"nombre":        refreshClaims.Nombre,
		},
	}
	_ = c.ServeJSON()
}

// GetJWTSecret retorna el secreto JWT para uso en otros controladores
func GetJWTSecret() []byte {
	return jwtSecret
}

// ParseTokenClaims parsea un token JWT y retorna los claims
func ParseTokenClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

func ValidateToken(ctx *context.Context) {
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.Status = http.StatusOK
		return
	}

	method := ctx.Input.Method()
	path := ctx.Input.URI()
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}

	if method == http.MethodGet || method == http.MethodPost {
		switch path {
		case "/restaurante/v1/productos", "/restaurante/v1/productos/search",
			"/restaurante/v1/restaurantes", "/restaurante/v1/restaurantes/search",
			"/restaurante/v1/reservas", "/restaurante/v1/reservas/search",
			"/restaurante/v1/reservas/parameter", "/restaurante/v1/reservas/cliente",
			"/restaurante/v1/reservas/documento",
			"/restaurante/v1/cambios_horario/actual",
			"/restaurante/v1/ofertas/activas":
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
