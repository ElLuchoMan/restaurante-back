package test

import (
	"net/http"
	"os"
	"testing"
)

func TestSendRequest_IntegrationBranch(t *testing.T) {
	os.Setenv("INTEGRATION", "1")
	defer os.Unsetenv("INTEGRATION")
	r, _ := http.NewRequest(http.MethodGet, "/x", nil)
	if w := sendRequest(r); w == nil {
		t.Fatalf("expected recorder")
	}
}
