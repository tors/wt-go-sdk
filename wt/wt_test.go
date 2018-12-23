package wt

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	testApiKey       = "abc"
	testJWTAuthToken = "jwt-token"
)

// setup sets up a test HTTP server along with a wt.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	server := httptest.NewServer(mux)

	// client configured to use test server
	client, _ = NewClient(testApiKey, nil)
	url, _ := url.Parse(server.URL + "/")

	client.BaseURL = url
	client.JWTAuthToken = testJWTAuthToken

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

func TestNewClient(t *testing.T) {
	c, _ := NewClient("abc", nil)

	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}
