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

// Duración real de los tokens. expires_in de las respuestas se deriva de
// accessTokenTTL para que documentación y token no se desfasen.
const (
	accessTokenTTL  = 120 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour

	tokenTypeRefresh = "refresh"
)

// Claims son los claims del access token. TokenType solo viene informado
// ("refresh") cuando se presenta por error un refresh token como access token.
type Claims struct {
	Documento int64  `json:"documento"`
	Rol       string `json:"rol"`
	Nombre    string `json:"nombre"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Documento int64  `json:"documento"`
	Rol       string `json:"rol"`
	Nombre    string `json:"nombre"`
	TokenType string `json:"token_type"`
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
	// loginMaxEntries es el tamaño del mapa a partir del cual se purgan las ventanas vencidas.
	loginMaxEntries = 10000
	rlMutex         sync.Mutex
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
	now := time.Now()
	if len(loginRL.m) >= loginMaxEntries {
		// el mapa no puede crecer sin límite: descarta las ventanas vencidas
		for k, e := range loginRL.m {
			if now.After(e.reset) {
				delete(loginRL.m, k)
			}
		}
	}
	entry, ok := loginRL.m[ip]
	if !ok || now.After(entry.reset) {
		loginRL.m[ip] = &rateEntry{count: 1, reset: now.Add(loginWindow)}
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
// @Description Permite iniciar sesión utilizando el documento y la contraseña (se busca primero entre trabajadores y luego entre clientes). Devuelve un access token JWT (`token` y `access_token`, mismo valor, rol incluido en el claim `rol`; "Cliente" para clientes), un `refresh_token`, `token_type` ("Bearer") y `expires_in` (segundos de vida del access token, 7200 = 120 min, como string). Límite: 10 intentos por minuto y por IP.
// @Tags login
// @Accept json
// @Produce json
// @Param   body  body   models.LoginRequest  true  "Documento (número) y contraseña"
// @Success 200 {object} models.ApiResponse{data=models.AuthResponse} "Inicio de sesión exitoso con tokens JWT"
// @Failure 400 {object} models.ApiResponse "JSON inválido, o documento/password ausentes"
// @Failure 401 {object} models.ApiResponse "Credenciales inválidas"
// @Failure 429 {object} models.ApiResponse "Demasiadas solicitudes"
// @Failure 500 {object} models.ApiResponse "Error de base de datos o al generar el token"
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

	if loginRequest.Documento == 0 || loginRequest.Password == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Los campos 'documento' y 'password' son obligatorios",
		}
		_ = c.ServeJSON()
		return
	}

	o := newOrm()

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: loginRequest.Documento}
	err := o.Read(&trabajador)
	if err == nil {
		if compareHashAndPassword([]byte(trabajador.PASSWORD), []byte(loginRequest.Password)) != nil {
			c.invalidCredentials()
			return
		}
		generateJWT(c, trabajador.PK_DOCUMENTO_TRABAJADOR, string(trabajador.ROL), trabajador.NOMBRE+" "+trabajador.APELLIDO)
		return
	}
	if err != orm.ErrNoRows {
		c.loginDBError(err)
		return
	}

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: loginRequest.Documento}
	err = o.Read(&cliente)
	if err == nil {
		if compareHashAndPassword([]byte(cliente.PASSWORD), []byte(loginRequest.Password)) != nil {
			c.invalidCredentials()
			return
		}
		generateJWT(c, cliente.PK_DOCUMENTO_CLIENTE, "Cliente", cliente.NOMBRE+" "+cliente.APELLIDO)
		return
	}
	if err != orm.ErrNoRows {
		c.loginDBError(err)
		return
	}

	c.invalidCredentials()
}

func (c *LoginController) invalidCredentials() {
	c.Ctx.Output.SetStatus(http.StatusUnauthorized)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusUnauthorized,
		Message: "Credenciales inválidas",
	}
	_ = c.ServeJSON()
}

func (c *LoginController) loginDBError(err error) {
	c.Ctx.Output.SetStatus(http.StatusInternalServerError)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusInternalServerError,
		Message: "Error al consultar el usuario",
		Cause:   err.Error(),
	}
	_ = c.ServeJSON()
}

func generateTokens(documento int64, rol, nombre string) (string, string, error) {
	if len(jwtSecret) == 0 {
		return "", "", fmt.Errorf("secreto JWT no configurado")
	}

	now := time.Now()

	accessClaims := &Claims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshClaims := &RefreshClaims{
		Documento: documento,
		Rol:       rol,
		Nombre:    nombre,
		TokenType: tokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
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

func newAuthResponse(accessToken, refreshToken, nombre string) models.AuthResponse {
	return models.AuthResponse{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    strconv.Itoa(int(accessTokenTTL / time.Second)),
		Nombre:       nombre,
	}
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
		Data:    newAuthResponse(accessToken, refreshToken, nombre),
	}
	_ = c.ServeJSON()
}

// @Title RefreshToken
// @Summary Renovar access token usando refresh token
// @Description Permite obtener un nuevo access token (y un nuevo refresh token) utilizando un refresh token válido enviado en el header Authorization (con o sin prefijo "Bearer "). Un access token no sirve como refresh token. `expires_in` son los segundos de vida del access token (7200 = 120 min) como string.
// @Tags auth
// @Accept json
// @Produce json
// @Param   Authorization  header  string  true  "Refresh Token en formato: Bearer {token}"
// @Success 200 {object} models.ApiResponse{data=models.AuthResponse} "Tokens renovados exitosamente"
// @Failure 400 {object} models.ApiResponse "Solicitud incorrecta"
// @Failure 401 {object} models.ApiResponse "Refresh token inválido, expirado o no es un refresh token"
// @Failure 500 {object} models.ApiResponse "Error al generar los tokens"
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

	if refreshClaims.TokenType != tokenTypeRefresh {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Token inválido: no es un refresh token",
		}
		_ = c.ServeJSON()
		return
	}

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
		Data:    newAuthResponse(accessToken, newRefreshToken, refreshClaims.Nombre),
	}
	_ = c.ServeJSON()
}

func GetJWTSecret() []byte {
	return jwtSecret
}

// ParseTokenClaims valida un access token. Un refresh token presentado como
// access token se rechaza.
func ParseTokenClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid || claims.TokenType == tokenTypeRefresh {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

// rutas públicas (sin token) para GET.
var publicGetPaths = map[string]bool{
	"/restaurante/v1/productos":              true,
	"/restaurante/v1/productos/search":       true,
	"/restaurante/v1/restaurantes":           true,
	"/restaurante/v1/restaurantes/search":    true,
	"/restaurante/v1/reservas":               true,
	"/restaurante/v1/reservas/search":        true,
	"/restaurante/v1/reservas/parameter":     true,
	"/restaurante/v1/reservas/cliente":       true,
	"/restaurante/v1/reservas/documento":     true,
	"/restaurante/v1/cambios_horario/actual": true,
	"/restaurante/v1/ofertas/activas":        true,
}

// rutas públicas (sin token) para POST: reservas públicas y registro de clientes.
var publicPostPaths = map[string]bool{
	"/restaurante/v1/reservas": true,
	"/restaurante/v1/clientes": true,
}

func requestPath(ctx *context.Context) string {
	path := ctx.Input.URL()
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func isPublicRoute(method, path string) bool {
	switch method {
	case http.MethodGet:
		return publicGetPaths[path]
	case http.MethodPost:
		return publicPostPaths[path]
	}
	return false
}

func writeAuthError(ctx *context.Context, status int, message string) {
	ctx.Output.SetStatus(status)
	_ = ctx.Output.JSON(models.ApiResponse{Code: status, Message: message}, false, false)
}

// authenticate exige un access token Bearer válido; si falla responde 401 y
// devuelve nil.
func authenticate(ctx *context.Context) *Claims {
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		writeAuthError(ctx, http.StatusUnauthorized, "Token no proporcionado")
		return nil
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	claims, err := ParseTokenClaims(authHeader[len("Bearer "):])
	if err != nil {
		writeAuthError(ctx, http.StatusUnauthorized, "Token inválido")
		return nil
	}
	return claims
}

// ValidateToken exige un access token válido salvo en las rutas públicas.
func ValidateToken(ctx *context.Context) {
	method := ctx.Input.Method()
	if method == http.MethodOptions {
		ctx.Output.Status = http.StatusOK
		return
	}

	path := requestPath(ctx)
	if isPublicRoute(method, path) {
		return
	}

	if web.BConfig.RunMode == "dev" {
		referer := ctx.Input.Header("Referer")
		if strings.Contains(referer, "/swagger/") || strings.HasPrefix(path, "/swagger/") {
			return
		}
	}

	authenticate(ctx)
}

// ValidateAdmin exige un access token válido cuyo claim `rol` sea Administrador:
// 401 sin token o con token inválido, 403 si el rol no es administrador. No tiene
// excepciones de rutas públicas ni de Swagger en modo dev.
func ValidateAdmin(ctx *context.Context) {
	if ctx.Input.Method() == http.MethodOptions {
		ctx.Output.Status = http.StatusOK
		return
	}

	claims := authenticate(ctx)
	if claims == nil {
		return
	}
	if claims.Rol != string(models.RolAdministrador) {
		writeAuthError(ctx, http.StatusForbidden, "Se requiere rol Administrador")
	}
}
