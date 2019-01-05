// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/tors/wt-go-sdk/wt"
)

func TestBoardsService(t *testing.T) {
	ctx := context.Background()

	// Create board
	desc := "Testing board service..."
	board, err := client.Boards.Create(ctx, "Pony board!", &desc)
	if err != nil {
		t.Errorf("Boards.Create returned an error %v", err)
	}

	// Add links to board
	title := "WeTransfer website"
	link, _ := wt.NewLink("https://wetransfer.com", &title)
	_, err = client.Boards.AddLinks(ctx, board, link)
	if err != nil {
		t.Errorf("Boards.AddLinks returned an error %v", err)
	}

	// Add files to board
	pony := wt.NewBuffer("pony.txt", []byte("yeehaaa"))
	_, err = client.Boards.AddFiles(ctx, board, pony)
	if err != nil {
		t.Errorf("Boards.AddFiles returned an error %v", err)
	}

	// Find board
	id := board.GetID()
	foundBoard, err := client.Boards.Find(ctx, id)
	if err != nil {
		t.Errorf("Boards.Find returned an error %v", err)
	}

	url := foundBoard.GetURL()
	logf("Got URL: %v", url)

	if url == "" {
		t.Errorf("Board.GetURL returned an empty string")
	}
}
