package wt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
)

const (
	testAPIKey       = "abc"
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
	client, _ = NewClient(testAPIKey, nil)
	url, _ := url.Parse(server.URL + "/")

	client.BaseURL = url
	client.JWTAuthToken = testJWTAuthToken

	return client, mux, server.URL, server.Close
}

func setupS3() (mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	return mux, server.URL, server.Close
}

// testHeader checks http methods
func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

// testHeader checks for values set in the http header
func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

// testErrorResponse checks the message of an ErrorResponse. If it matches that
// given message string, then it passes.
func testErrorResponse(t *testing.T, err error, message string) {
	v, ok := err.(*ErrorResponse)

	if ok && v.Message != message {
		t.Errorf("ErrorResponse.Message returned %v, want %+v", v.Message, message)
	}

	if !ok {
		t.Errorf("error is not an ErrorResponse kind: %v", err)
	}
}

// setupTestFile creates a new file with the given name and content for testing.
func setupTestFile(t *testing.T, name, content string) *os.File {
	dir, err := ioutil.TempDir("", "wt-go-sdk")
	if err != nil {
		t.Errorf("openTestFile returned an error: %v", err)
	}

	file, err := os.OpenFile(path.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Errorf("openTestFile returned an error: %v", err)
	}

	fmt.Fprint(file, content)

	// close and re-open the file to keep file.Stat() happy
	file.Close()
	file, err = os.Open(file.Name())
	if err != nil {
		t.Errorf("openTestFile returned an error: %v", err)
	}

	return file
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
