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
	board, _ := client.Boards.Create(context.Background(), &name, nil)

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
