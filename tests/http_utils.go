package test

import (
    "net/http"
    "net/http/httptest"

    beego "github.com/beego/beego/v2/server/web"
)

// sendRequest helps execute an HTTP request against the Beego
// application and returns the recorder with the response. It is
// used across tests to reduce boilerplate and provides statements
// so coverage reporting works for this package.
func sendRequest(req *http.Request) *httptest.ResponseRecorder {
    w := httptest.NewRecorder()
    beego.BeeApp.Handlers.ServeHTTP(w, req)
    return w
}
