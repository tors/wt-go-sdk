package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func getValidTransferHandler(cb func(*http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
				  "name" : "big-bobis.jpg",
				  "id" : "random-hash-1"
				}
			  ]
			}
		`)
		cb(r)
	}
}

func TestTransferService_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := getValidTransferHandler(func(r *http.Request) {
		testMethod(t, r, "POST")
	})

	mux.HandleFunc("/transfers", handler)

	param := NewTransferParam("")
	param.AddFile("big-bobis.jpg", 195906)

	transfer, err := client.Transfer.Create(context.Background(), param)

	if err != nil {
		t.Errorf("TransferService.Create returned an error: %v", err)
	}

	want := &Transfer{
		Success:   Bool(true),
		ID:        String("random-hash"),
		Message:   String("Please deliver"),
		State:     String("uploading"),
		URL:       nil,
		ExpiresAt: String("2019-01-01T00:00:00Z"),
		Files: []*File{
			&File{
				Multipart: &struct {
					PartNumbers *int64 `json:"part_numbers"`
					ChunkSize   *int64 `json:"chunk_size"`
				}{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Size: Int64(195906),
				Type: String("file"),
				Name: String("big-bobis.jpg"),
				ID:   String("random-hash-1"),
			},
		},
	}

	if !reflect.DeepEqual(transfer, want) {
		t.Errorf("TransferService.Create returned %v, want %v", transfer, want)
	}
}

func TestTransferService_Create_emptyFiles(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := getValidTransferHandler(func(r *http.Request) {
		testMethod(t, r, "POST")
	})

	mux.HandleFunc("/transfers", handler)

	param := NewTransferParam("")
	_, err := client.Transfer.Create(context.Background(), param)

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if err != ErrEmptyFiles {
		t.Errorf("Error is %+v, want %+v", err, ErrEmptyFiles)
	}
}

func TestTransferService_Find(t *testing.T) {
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
				  "name": "big-bobis.jpg",
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

	_, err := client.Transfer.Find(context.Background(), "random-hash")

	if err != nil {
		t.Errorf("TransferService.Find returned an error: %v", err)
	}
}

func TestTransferService_Find_notFound(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "Couldn't find Transfer. See https://developers.wetransfer.com/documentation"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(404)
		fmt.Fprint(w, fmt.Sprintf(`
			{
			  "success" : false,
			  "message": "%v"
			}
		`, wantError))
	})

	_, err := client.Transfer.Find(context.Background(), "random-hash")

	if err, ok := err.(*ErrorResponse); !ok && err.Message != wantError {
		t.Errorf("ErrorResponse.Message returned %v, want %+v", err.Message, wantError)
	}
}
