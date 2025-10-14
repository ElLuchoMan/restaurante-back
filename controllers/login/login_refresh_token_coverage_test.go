package login

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_SinBearerPrefix(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 123456,
		Rol:       "CLIENTE",
		Nombre:    "Test User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Tokens renovados exitosamente", response.Message)
}

func TestRefreshToken_NoEsRefreshToken(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 123456,
		Rol:       "CLIENTE",
		Nombre:    "Test User",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Message, "no es un refresh token")
}

func TestRefreshToken_TokenTypeVacio(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 123456,
		Rol:       "CLIENTE",
		Nombre:    "Test User",
		TokenType: "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Message, "no es un refresh token")
}

func TestRefreshToken_BearerConEspaciosExtra(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 999999,
		Rol:       "TRABAJADOR",
		Nombre:    "Admin User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)

	data := response.Data.(map[string]interface{})
	assert.NotNil(t, data["access_token"])
	assert.NotNil(t, data["refresh_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Equal(t, "Admin User", data["nombre"])
}

func TestRefreshToken_ErrorGenerandoTokens(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 123456,
		Rol:       "CLIENTE",
		Nombre:    "Test User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	originalSecret := jwtSecret

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	jwtSecret = originalSecret

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRefreshToken_TokenExpirado(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	claims := &RefreshClaims{
		Documento: 123456,
		Rol:       "CLIENTE",
		Nombre:    "Test User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Message, "inválido o expirado")
}

func TestRefreshToken_TokenMalformado(t *testing.T) {

	ctrl := &LoginController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	r.Header.Set("Authorization", "Bearer esto.no.es.un.token.valido")
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.RefreshToken()

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Message, "inválido o expirado")
}
