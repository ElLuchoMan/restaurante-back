package test

import (
	"net/http"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestLoginInvalidJSON ensures that the login endpoint returns
// a 400 Bad Request response when the provided JSON is malformed.
func TestLoginInvalidJSON(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION=1 to run")
	}
	rBody := strings.NewReader("{")
	req, _ := http.NewRequest("POST", "/restaurante/v1/login", rBody)
	req.Header.Set("Content-Type", "application/json")
	w := sendRequest(req)

	Convey("POST /restaurante/v1/login with invalid JSON should return 400", t, func() {
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}
