package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestTransfersService_createTransfer(t *testing.T) {
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
				  "name" : "kitty.txt",
				  "id" : "random-hash-1"
				}
			  ]
			}
		`)
	})

	file, _, err := openTestFile("kitty.txt", "meow")
	defer file.Close()

	transfer, err := client.Transfers.createTransfer(context.Background(), file.Name(), nil)

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
		Files: []*File{
			&File{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Size: Int64(195906),
				Type: String("file"),
				Name: String("kitty.txt"),
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

	file, _, err := openTestFile("kitty.txt", "meow")
	defer file.Close()

	_, err = client.Transfers.createTransfer(context.Background(), file.Name(), nil)

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
