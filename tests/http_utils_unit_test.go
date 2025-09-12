package test

import (
	"net/http"
	"testing"
)

func TestSendRequest_NilBodyUnitBranch(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/x", nil)
	if w := sendRequest(r); w == nil {
		t.Fatalf("expected recorder")
	}
	if r.Body == nil {
		t.Fatalf("expected body to be initialized")
	}
}
