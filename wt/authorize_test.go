package wt

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestAuthorize(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, fmt.Sprintf(`{"success":true,"token":"%s"}`, testJWTAuthToken))
	})

	err := Authorize(context.Background(), client)
	if err != nil {
		t.Errorf("Authorize returned an error: %v", err)
	}

	if client.JWTAuthToken != testJWTAuthToken {
		t.Errorf("Client.JWTAuthToken returned %v, want %v", client.JWTAuthToken, testJWTAuthToken)
	}
}

func TestAuthorize_forbidden(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "Forbidden: invalid API Key"

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(403)
		fmt.Fprint(w, fmt.Sprintf(`{"success":false,"message":"%s"}`, wantError))
	})

	err := Authorize(context.Background(), client)

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if err, ok := err.(*ErrorResponse); !ok && err.Message != wantError {
		t.Errorf("ErrorResponse.Message returned %v, want %+v", err.Message, wantError)
	}
}
