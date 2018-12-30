package wt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "https://dev.wetransfer.com/v2/"
	userAgent      = "go-wt"
	contentType    = "application/json"
)

type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public WeTransfer API.
	// Base URL should always be specified with a trailing slash.
	BaseURL *url.URL

	// WeTransfer API key
	APIKey string

	// WeTransfer JWT Authorization token
	JWTAuthToken string

	// User agent used when communicating with the API.
	UserAgent string

	// Reuse a single struct instead of allocating one for each service on the heap.
	common service

	// Services used for talking to different parts of the API.
	Transfers *TransfersService
	Boards    *BoardsService

	// Service that allow for multipart file uploads
	uploader *uploaderService
}

type service struct {
	client *Client
}

// NewClient returns a new WeTransfer unauthorized API client. If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(apiKey string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if apiKey == "" {
		return nil, fmt.Errorf("APIKey must not be blank")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		APIKey:    apiKey,
		UserAgent: userAgent,
	}

	c.common.client = c

	c.Transfers = (*TransfersService)(&c.common)
	c.Boards = (*BoardsService)(&c.common)
	c.uploader = (*uploaderService)(&c.common)

	return c, nil
}

// NewAuthorizedClient returns a new WeTransfer authorized API client.
func NewAuthorizedClient(ctx context.Context, apiKey string, httpClient *http.Client) (*Client, error) {
	client, err := NewClient(apiKey, nil)
	if err != nil {
		return nil, err
	}

	err = Authorize(ctx, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	if c.JWTAuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.JWTAuthToken))
	}

	req.Header.Set("x-api-key", c.APIKey)

	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

// CheckResponse checks the API response for errors, and returns them if
// present.
// WeTransfer API docs: https://developers.wetransfer.com/documentation
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errResp := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if data != nil {
		err := json.Unmarshal(data, errResp)
		if err != nil {
			return err
		}
	}

	return errResp
}

// An ErrorResponse reports the errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }

type Errors struct {
	message string
	errors  []error
}

func (e *Errors) Append(err error) {
	e.errors = append(e.errors, err)
}

func (e *Errors) Error() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v:\n", e.message)
	for _, er := range e.errors {
		fmt.Fprintf(buf, "%v", er.Error())
	}
	return buf.String()
}

func (e *Errors) Len() int {
	return len(e.errors)
}

func (e *Errors) GetErrors() []error {
	return e.errors
}

func NewErrors(m string) *Errors {
	return &Errors{
		message: m,
		errors:  make([]error, 0),
	}
}
