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
