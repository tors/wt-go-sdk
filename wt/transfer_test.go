package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestTransferServiceCreate(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{
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
	})

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
