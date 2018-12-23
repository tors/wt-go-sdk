package wt

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAuthorize(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"success":true,"token":"abc"}`)
	})

	err := Authorize(client)
	if err != nil {
		t.Errorf("Authorize returned an error: %v", err)
	}

	if client.JWTAuthToken != testApiKey {
		t.Errorf("Client.JWTAuthToken returned %v, want %v", client.JWTAuthToken, testApiKey)
	}
}
