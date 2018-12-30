package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestTransfersService_Complete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/1/files/1/upload-complete", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id": "1", "retries": 0, "name": "pony1.txt", "size": 2, "chunk_size": 2}`)
	})
	mux.HandleFunc("/transfers/1/files/2/upload-complete", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id": "2", "retries": 1, "name": "pony2.txt", "size": 4, "chunk_size": 4}`)
	})

	tx := &Transfer{
		ID: String("1"),
		Files: []*File{
			&File{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(2),
				},
				ID: String("1"),
			},
			&File{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(2),
				},
				ID: String("2"),
			},
		},
	}

	want := []*CompletedTransfer{
		&CompletedTransfer{
			ID:        String("1"),
			Retries:   Int64(0),
			Name:      String("pony1.txt"),
			Size:      Int64(2),
			ChunkSize: Int64(2),
		},
		&CompletedTransfer{
			ID:        String("2"),
			Retries:   Int64(1),
			Name:      String("pony2.txt"),
			Size:      Int64(4),
			ChunkSize: Int64(4),
		},
	}

	completed, err := client.Transfers.Complete(context.Background(), tx)
	if err != nil {
		t.Errorf("TransfersService.Complete returned an error: %v", err)
	}

	if !reflect.DeepEqual(completed, want) {
		t.Errorf("TransfersService.Complete returned %v, want %v", completed, want)
	}
}

func TestTransfersService_createTransfer(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	filename := "pony.txt"
	message := "My first pony"

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "x-api-key", testApiKey)
		testHeader(t, r, "Authorization", fmt.Sprintf("Bearer %v", testJWTAuthToken))

		fmt.Fprintf(w, `
			{
			  "success" : true,
			  "id" : "random-hash",
			  "message" : "%v",
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
				  "name" : "%v",
				  "id" : "random-hash-1"
				}
			  ]
			}
		`, message, filename)
	})

	tfile := setupTestFile(t, filename, message)
	defer tfile.Close()

	file, err := NewBufferedFile(tfile.Name())
	if err != nil {
		t.Errorf("NewBufferedFile returned an error: %v", err)
	}
	defer file.Close()

	transfer, err := client.Transfers.createTransfer(context.Background(), &message, file)
	if err != nil {
		t.Errorf("TransfersService.createTransfer returned an error: %v", err)
	}

	want := &Transfer{
		Success:   Bool(true),
		ID:        String("random-hash"),
		Message:   &message,
		State:     String("uploading"),
		URL:       nil,
		ExpiresAt: String("2019-01-01T00:00:00Z"),
		Files: []*File{
			&File{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Size: Int64(195906),
				Type: String("file"),
				Name: &filename,
				ID:   String("random-hash-1"),
			},
		},
	}

	if !reflect.DeepEqual(transfer, want) {
		t.Errorf("TransfersService.Create returned %v, want %v", transfer, want)
	}
}

func TestTransfersService_createTransfer_badRequest(t *testing.T) {
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

	buf := NewBuffer("kitty.txt", []byte("meow"))
	_, err := client.Transfers.createTransfer(context.Background(), nil, buf)

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
			  "url": "https://we.tl/t-meowdEFgHi12",
			  "expires_at": "2018-01-01T00:00:00Z",
			  "files": [
				{
				  "id": "another-random-hash",
				  "type": "file",
				  "name": "kitty.txt",
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
