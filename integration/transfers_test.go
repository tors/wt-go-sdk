// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/tors/wt-go-sdk/wt"
)

func TestTransfers_Create_buffer(t *testing.T) {
	ctx := context.Background()

	message := "My first pony!"
	buffer := wt.NewBuffer("pony.txt", []byte("yeehaaa"))

	transfer, err := client.Transfers.Create(ctx, &message, buffer)
	if err != nil {
		t.Errorf("Transfers.Create returned an error %v", err)
	}

	url := transfer.GetURL()
	logf("Got URL: %v", url)

	if url == "" {
		t.Errorf("Transfer.GetURL is expected not be empty")
	}
}

func TestTransfers_Create_files(t *testing.T) {
	ctx := context.Background()

	message := "Japan files"

	files := []string{
		"../example/files/Japan-01🇯🇵.jpg",
		"../example/files/Japan-02.jpg",
		"../example/files/Japan-03.jpg",
	}

	transfer, err := client.Transfers.Create(ctx, &message, files)
	if err != nil {
		t.Errorf("Transfers.Create returned an error %v", err)
	}

	url := transfer.GetURL()
	logf("Got URL: %v", url)

	if url == "" {
		t.Errorf("Transfer.GetURL is expected not be empty")
	}
}
