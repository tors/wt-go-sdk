package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestBoardsService_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/boards", func(w http.ResponseWriter, r *http.Request) {

		testMethod(t, r, "POST")
		testHeader(t, r, "x-api-key", testApiKey)
		testHeader(t, r, "Authorization", fmt.Sprintf("Bearer %v", testJWTAuthToken))

		fmt.Fprint(w, `
			{
			  "id": "random-hash",
			  "name": "Not pinterest",
			  "description": null,
			  "state": "downloadable",
			  "url": "https://we.tl/b-random-hash",
			  "items": []
			}
		`)
	})

	name := "Not pinterest"
	board, _ := client.Boards.Create(context.Background(), name, nil)

	want := &Board{
		ID:    String("random-hash"),
		Name:  String("Not pinterest"),
		Desc:  nil,
		State: String("downloadable"),
		URL:   String("https://we.tl/b-random-hash"),
		Items: []*Item{},
	}

	if !reflect.DeepEqual(board, want) {
		t.Errorf("Board.Create returned %+v, want %+v", board, want)
	}
}

func TestBoardsService_Find(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/boards/board-id", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
			{
			  "id": "board-id",
			  "state": "processing",
			  "url": "https://we.tl/b-the-boards-url",
			  "name": "Little kittens",
			  "description": null,
			  "items": [
				{
				  "id": "random-hash",
				  "name": "kittie.gif",
				  "size": 195906,
				  "multipart": {
					"part_numbers": 1,
					"chunk_size": 195906
				  },
				  "type": "file"
				},
				{
				  "id": "different-random-hash",
				  "url": "https://wetransfer.com",
				  "meta": {
					"title": "WeTransfer"
				  },
				  "type": "link"
				}
			  ]
			}
		`)
	})

	board, err := client.Boards.Find(context.Background(), "board-id")

	if err != nil {
		t.Errorf("TransfersService.Find returned an error: %v", err)
	}

	want := &Board{
		ID:    String("board-id"),
		Name:  String("Little kittens"),
		Desc:  nil,
		State: String("processing"),
		URL:   String("https://we.tl/b-the-boards-url"),
		Items: []*Item{
			{
				ID:   String("random-hash"),
				Name: String("kittie.gif"),
				Size: Int64(195906),
				Multipart: &Multipart{
					PartNumbers: Int64(1),
					ChunkSize:   Int64(195906),
				},
				Type: String("file"),
			},
			{
				ID:  String("different-random-hash"),
				URL: String("https://wetransfer.com"),
				Meta: &Meta{
					Title: String("WeTransfer"),
				},
				Type: String("link"),
			},
		},
	}

	if !reflect.DeepEqual(board, want) {
		t.Errorf("Board.Find returned %+v, want %+v", board, want)
	}
}

func TestBoardsService_Find_forbidden(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	wantError := "You're not a member of this board (123456). See https://developers.wetransfer.com/documentation"
	mux.HandleFunc("/boards/board-id", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(403)

		fmt.Fprintf(w, `{
			"success": false,
			"message": "%s"
		}`, wantError)
	})

	_, err := client.Boards.Find(context.Background(), "board-id")

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	testErrorResponse(t, err, wantError)
}
