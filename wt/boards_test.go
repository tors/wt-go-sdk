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
		testHeader(t, r, "x-api-key", testAPIKey)
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

func TestBoardsService_AddLink(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/boards/1/links", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `
			[
			  {
				"id": "99",
				"url": "https://wetransfer.com/",
				"meta": {
				  "title": "WeTransfer"
				},
				"type": "link"
			  }
			]
		`)
	})

	title := "WeTransfer website"
	link, _ := NewLink("https://wetransfer.com", &title)

	board := &Board{
		ID: String("1"),
	}
	item, err := client.Boards.AddLinks(context.Background(), board, link)
	if err != nil {
		t.Errorf("BoardsService.AddLinks returned an error %v", err)
	}

	wantItem := []*Item{
		{
			ID:  String("99"),
			URL: String("https://wetransfer.com/"),
			Meta: &Meta{
				Title: String("WeTransfer"),
			},
			Type: String("link"),
		},
	}

	if !reflect.DeepEqual(item, wantItem) {
		t.Errorf("BoardsService.AddLinks returned %v, want %v", item, wantItem)
	}
}

func TestBoardsService_AddLinks_badRequest(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/boards/1/links", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprint(w, `
			{
			  "success": false,
			  "message": "\"board.links\" must be an array."
			}
		`)
	})

	title := "WeTransfer website"
	link, _ := NewLink("https://wetransfer.com", &title)

	board := &Board{
		ID: String("1"),
	}

	_, err := client.Boards.AddLinks(context.Background(), board, link)
	if err == nil {
		t.Errorf("Expected error to be returned.")
	}

	testErrorResponse(t, err, "\"board.links\" must be an array.")
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

func TestBoardsService_complete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	board := &Board{
		ID:    String("1"),
		Items: []*Item{},
	}
	items := []*Item{
		{
			ID: String("1"),
		},
		{
			ID: String("2"),
		},
	}

	for _, item := range items {
		path := fmt.Sprintf("/boards/%v/files/%v/upload-complete", board.GetID(), item.GetID())
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "PUT")
			fmt.Fprint(w, `{
				"success": true,
				"message": "File is marked as complete."
			}`)
		})
	}

	err := client.Boards.complete(context.Background(), board, items)
	if err != nil {
		t.Errorf("Boards.complete returned an error %v", err)
	}
}

func TestBoardsService_complete_badRequest(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/boards/1/files/1/upload-complete", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprint(w, `{
		  "success": false,
		  "message": "expected at least 1 part."
		}`)
	})

	board := &Board{
		ID:    String("1"),
		Items: []*Item{},
	}
	items := []*Item{
		{
			ID: String("1"),
		},
	}

	err := client.Boards.complete(context.Background(), board, items)
	if err == nil {
		t.Error("Expected error to be returned")
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
