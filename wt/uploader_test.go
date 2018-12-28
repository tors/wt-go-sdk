package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestUploaderService_getUploadURL(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/1/files/1/upload-url/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-1"}`)
	})
	mux.HandleFunc("/transfers/1/files/1/upload-url/2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-2"}`)
	})

	tests := []struct {
		in  int64
		out *UploadURL
	}{
		{int64(1), &UploadURL{Success: Bool(true), URL: String("https://s3-1")}},
		{int64(2), &UploadURL{Success: Bool(true), URL: String("https://s3-2")}},
	}

	ts := &Transfer{ID: String("1")}

	for _, test := range tests {
		got, _ := client.uploader.getUploadURL(context.Background(), ts, "1", test.in)
		if !reflect.DeepEqual(got, test.out) {
			t.Errorf("Transfers.GetUploadURL returned %+v, want %+v", got, test.out)
		}
	}
}

func TestUploaderService_getUploadURL_fail(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	tests := []struct {
		url       string
		httpCode  int
		partNum   int64
		wantError string
	}{
		{"/transfers/2/files/2/upload-url/1", 404, int64(1), "Invalid transfer or file id."},
		{"/transfers/2/files/2/upload-url/0", 417, int64(0), "Chunk numbers are 1-based"},
	}

	ts := &Transfer{ID: String("2")}

	for _, g := range tests {
		mux.HandleFunc(g.url, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(g.httpCode)
			fmt.Fprintf(w, `{"success":false,"message":"%v"}`, g.wantError)
		})
		_, err := client.uploader.getUploadURL(context.Background(), ts, "2", g.partNum)
		testErrorResponse(t, err, g.wantError)
	}
}
