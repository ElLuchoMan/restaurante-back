package webpush

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type Keys struct {
	P256dh string
	Auth   string
}

type Subscription struct {
	Endpoint string
	Keys     Keys
}

type Options struct {
	Subscriber      string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	TTL             int
}

func SendNotificationWithContext(ctx context.Context, payload []byte, subscription *Subscription, options *Options) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}
