package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestTransfersService_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {

		testMethod(t, r, "POST")
		testHeader(t, r, "x-api-key", testApiKey)
		testHeader(t, r, "Authorization", fmt.Sprintf("Bearer %v", testJWTAuthToken))

		fmt.Fprint(w, `
			{
			  "success" : true,
			  "id" : "random-hash",
			  "message" : "Please deliver",
			  "state" : "uploading",
			  "url" : null,
			  "expires_at": "2019-01-01T00:00:00Z",
			  "files" : [
				{
				  "multipart" : {
					"part_numbers" : 1,
					"chunk_size" : 195906
				  },
				  "size" : 195906,
				  "type" : "file",
				  "name" : "filename.jpg",
				  "id" : "random-hash-1"
				}
			  ]
			}
		`)
	})

	object, _ := FromString("This is some content.", "filename.txt")
	fo := []*FileObject{object}

	transfer, err := client.Transfers.Create(context.Background(), nil, fo)

	if err != nil {
		t.Errorf("TransfersService.Create returned an error: %v", err)
	}

	want := &Transfer{
		Success:   Bool(true),
		ID:        String("random-hash"),
		Message:   String("Please deliver"),
		State:     String("uploading"),
		URL:       nil,
		ExpiresAt: String("2019-01-01T00:00:00Z"),
		Files: []*RemoteFile{
			&RemoteFile{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Size: Int64(195906),
				Type: String("file"),
				Name: String("filename.jpg"),
				ID:   String("random-hash-1"),
			},
		},
	}

	if !reflect.DeepEqual(transfer, want) {
		t.Errorf("TransfersService.Create returned %v, want %v", transfer, want)
	}
}

func TestTransfersService_Create_badRequest(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "Bad request"
	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintf(w, `
			{
				"success": false,
				"message": "%v"
			}
		`, wantError)
	})

	object, _ := FromString("abc", "abc.txt")
	fo := []*FileObject{object}

	_, err := client.Transfers.Create(context.Background(), nil, fo)

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	testErrorResponse(t, err, wantError)
}

func TestTransfersService_Find(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/random-hash", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
			{
			  "id": "random-hash",
			  "message": "My very first transfer!",
			  "state": "downloadable",
			  "url": "https://we.tl/t-ABcdEFgHi12",
			  "expires_at": "2018-01-01T00:00:00Z",
			  "files": [
				{
				  "id": "another-random-hash",
				  "type": "file",
				  "name": "filename.jpg",
				  "multipart": {
					"chunk_size": 195906,
					"part_numbers": 1
				  },
				  "size": 195906
				}
			  ]
			}
		`)
	})

	_, err := client.Transfers.Find(context.Background(), "random-hash")

	if err != nil {
		t.Errorf("TransfersService.Find returned an error: %v", err)
	}
}

func TestTransfersService_Find_notFound(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "Couldn't find Transfer. See https://developers.wetransfer.com/documentation"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(404)
		fmt.Fprintf(w, `
			{
			  "success" : false,
			  "message": "%v"
			}
		`, wantError)
	})

	_, err := client.Transfers.Find(context.Background(), "random-hash")

	testErrorResponse(t, err, wantError)
}

func TestTransferService_getAllUploadURL(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/1/files/1/upload-url/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-put-url-111"}`)
	})
	mux.HandleFunc("/transfers/1/files/1/upload-url/2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-put-url-112"}`)
	})

	remoteFile := &RemoteFile{
		Multipart: &Multipart{
			PartNumbers: Int64(2),
			ChunkSize:   Int64(200),
		},
		ID: String("1"),
	}

	want := []*UploadURL{
		&UploadURL{
			Success: Bool(true),
			URL:     String("https://s3-put-url-111"),
			err:     nil,
		},
		&UploadURL{
			Success: Bool(true),
			URL:     String("https://s3-put-url-112"),
			err:     nil,
		},
	}

	got := client.Transfers.getAllUploadURL(context.Background(), "1", remoteFile)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Transfers.getAllUploadURL returned %+v, want %+v", got, want)
	}
}

func TestTransferService_getUploadURL_fail(t *testing.T) {
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

	for _, g := range tests {
		mux.HandleFunc(g.url, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(g.httpCode)
			fmt.Fprintf(w, `{"success":false,"message":"%v"}`, g.wantError)
		})
		uurl := client.Transfers.getUploadURL(context.Background(), "2", "2", g.partNum)
		err := uurl.GetError()
		testErrorResponse(t, err, g.wantError)
	}
}
