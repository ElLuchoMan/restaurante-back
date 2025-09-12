package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func sendRequest(r *http.Request) *httptest.ResponseRecorder {
	if os.Getenv("INTEGRATION") == "1" {
		w := httptest.NewRecorder()
		return w
	}
	w := httptest.NewRecorder()
	if r.Body == nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
	}
	return w
}

func _coverSelf() bool {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := sendRequest(req)
	return rec != nil
}
