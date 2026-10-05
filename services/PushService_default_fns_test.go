package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/assert"
)

// Capturadas antes de que otros tests sustituyan las variables por dobles.
var (
	defaultWebpushSendFn = webpushSendNotificationWithContextFn
	defaultHTTPDoFn      = httpDoFn
	defaultTokenFn       = tokenSourceTokenFn
)

func TestDefaultWebpushSendFn_ErrorConSuscripcionInvalida(t *testing.T) {
	sub := &webpush.Subscription{Endpoint: "http://127.0.0.1:1/push", Keys: webpush.Keys{P256dh: "no-valida", Auth: "no-valida"}}
	resp, err := defaultWebpushSendFn(context.Background(), []byte("{}"), sub, &webpush.Options{TTL: 1})
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.Error(t, err)
}

func TestDefaultHTTPDoFn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	assert.NoError(t, err)
	resp, err := defaultHTTPDoFn(&http.Client{Timeout: time.Second}, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestDefaultTokenSourceTokenFn_NilTokenSource(t *testing.T) {
	tok, err := defaultTokenFn(nil)
	assert.Nil(t, tok)
	assert.EqualError(t, err, "token source nil")
}
