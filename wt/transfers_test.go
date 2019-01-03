package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestTransfersService_Create(t *testing.T) {
	client, mux, srvURL, teardown := setup()
	defer teardown()

	// The Create test case is not designed to be an exhaustive test. We'll
	// leave that to the methods under the helper methods (see below test
	// cases) and uploaderService which it makes use of. We'll just have to
	// make sure that Create touches all endpoints and fulfill the whole
	// ceremony without throwing an single error.
	touchedEndpoints := 0

	// 1. create transfer
	// 2. request for upload URL
	// 3. actual file upload to s3
	// 4. complete transfer
	// 5. finalize transfer
	// 6. profit! (not counted)
	wantTouchedEndpoints := 5

	// We'll still have to output the correct formats, else the whole thing
	// breaks.
	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		touchedEndpoints++
		fmt.Fprintf(w, `
			{
			  "success" : true,
			  "id" : "1",
			  "message" : "My first pony!",
			  "state" : "uploading",
			  "url" : null,
			  "expires_at": "2019-01-01T00:00:00Z",
			  "files" : [
				{
				  "multipart" : {
					"part_numbers" : 1,
					"chunk_size" : 5
				  },
				  "size" : 5,
				  "type" : "file",
				  "name" : "pony.txt",
				  "id" : "1"
				}
			  ]
			}
		`)
	})
	mux.HandleFunc("/transfers/1/files/1/upload-url/1", func(w http.ResponseWriter, r *http.Request) {
		touchedEndpoints++
		fmt.Fprintf(w, `{"success": true, "url": "%v"}`, fmt.Sprintf("%v/part/%v", srvURL, 1))
	})
	mux.HandleFunc("/transfers/1/files/1/upload-complete", func(w http.ResponseWriter, r *http.Request) {
		touchedEndpoints++
		fmt.Fprint(w, `{"id": "1", "retries": 0, "name": "pony1.txt", "size": 2, "chunk_size": 2}`)
	})
	mux.HandleFunc("/transfers/1/finalize", func(w http.ResponseWriter, r *http.Request) {
		touchedEndpoints++
		fmt.Fprintf(w, `
			{
			  "success" : true,
			  "id" : "1",
			  "message" : "My first pony!",
			  "state" : "done",
			  "url" : "https://we.tl/t-12344657",
			  "expires_at": "2019-01-01T00:00:00Z",
			  "files" : [
				{
				  "multipart" : {
					"part_numbers" : 1,
					"chunk_size" : 5
				  },
				  "size" : 5,
				  "type" : "file",
				  "name" : "pony.txt",
				  "id" : "1"
				}
			  ]
			}
		`)
	})
	// The fake S3 endpoint. For convenience, we just use the same httptest server.
	mux.HandleFunc(fmt.Sprintf("/part/%v", 1), func(w http.ResponseWriter, r *http.Request) {
		touchedEndpoints++
		w.WriteHeader(200)
	})

	buf := NewBuffer("pony.txt", []byte("yehaa"))

	_, err := client.Transfers.Create(context.Background(), nil, buf)
	if err != nil {
		t.Errorf("TransfersService.Create returned an error: %v", err)
	}

	if touchedEndpoints != wantTouchedEndpoints {
		t.Errorf("TransfersService.Create number of endpoints touched %v, want %v", touchedEndpoints, wantTouchedEndpoints)
	}
}

func TestTransfersService_complete(t *testing.T) {
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
			{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(2),
				},
				ID: String("1"),
			},
			{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(2),
				},
				ID: String("2"),
			},
		},
	}

	want := []*completedTransfer{
		{
			ID:        String("1"),
			Retries:   Int64(0),
			Name:      String("pony1.txt"),
			Size:      Int64(2),
			ChunkSize: Int64(2),
		},
		{
			ID:        String("2"),
			Retries:   Int64(1),
			Name:      String("pony2.txt"),
			Size:      Int64(4),
			ChunkSize: Int64(4),
		},
	}

	completed, err := client.Transfers.complete(context.Background(), tx)
	if err != nil {
		t.Errorf("TransfersService.complete returned an error: %v", err)
	}

	if !reflect.DeepEqual(completed, want) {
		t.Errorf("TransfersService.complete returned %v, want %v", completed, want)
	}
}

func TestTransfersService_complete_expectationFailed(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "Chunks 1 are still missing."
	mux.HandleFunc("/transfers/1/files/1/upload-complete", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(417)
		fmt.Fprintf(w, `{"success": false, "message": "%v"}`, wantError)
	})

	tx := &Transfer{
		ID: String("1"),
		Files: []*File{
			{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(2),
				},
				ID: String("1"),
			},
		},
	}

	completed, err := client.Transfers.complete(context.Background(), tx)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if len(completed) != 0 {
		t.Errorf("Expected length to be 0")
	}
}

func TestTransfersService_createTransfer(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	filename := "pony.txt"
	message := "My first pony"

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "x-api-key", testAPIKey)
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

	file, err := BuildBufferedFile(tfile)
	if err != nil {
		t.Errorf("BuildBufferedFile returned an error: %v", err)
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
			{
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
		t.Errorf("TransfersService.createTransfer returned %v, want %v", transfer, want)
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

func TestTransfersService_finalize(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/1/finalize", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprintf(w, `
			{
			  "success" : true,
			  "id" : "1",
			  "message" : null,
			  "state" : "done",
			  "url" : "https://we.tl/t-12344657",
			  "expires_at": "2019-01-01T00:00:00Z",
			  "files" : [
				{
				  "multipart" : {
					"part_numbers" : 1,
					"chunk_size" : 195906
				  },
				  "size" : 195906,
				  "type" : "file",
				  "name" : "pony.txt",
				  "id" : "1"
				}
			  ]
			}
		`)
	})

	transfer, err := client.Transfers.finalize(context.Background(), "1")
	if err != nil {
		t.Errorf("TransfersService.finalize returned an error: %v", err)
	}

	want := &Transfer{
		Success:   Bool(true),
		ID:        String("1"),
		Message:   nil,
		State:     String("done"),
		URL:       String("https://we.tl/t-12344657"),
		ExpiresAt: String("2019-01-01T00:00:00Z"),
		Files: []*File{
			{
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Size: Int64(195906),
				Type: String("file"),
				Name: String("pony.txt"),
				ID:   String("1"),
			},
		},
	}

	if !reflect.DeepEqual(transfer, want) {
		t.Errorf("TransfersService.finalize returned %v, want %v", transfer, want)
	}
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
