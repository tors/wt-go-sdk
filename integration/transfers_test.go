// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/tors/wt-go-sdk/wt"
)

func TestTransfersService(t *testing.T) {
	ctx := context.Background()

	message := "My first pony!"

	pony := wt.NewBuffer("pony.txt", []byte("yeehaaa"))
	japan, _ := wt.NewLocalFile("../example/files/Japan-01🇯🇵.jpg")

	transfer, err := client.Transfers.Create(ctx, &message, pony, japan)
	if err != nil {
		t.Errorf("Transfers.Create returned an error %v", err)
	}

	url := transfer.GetURL()
	logf("Got URL: %v", url)

	if url == "" {
		t.Errorf("Transfer.GetURL returned an empty string")
	}

	id := transfer.GetID()
	_, err = client.Transfers.Find(ctx, id)

	if err != nil {
		t.Errorf("Transfers.Find returned an error %v", err)
	}
}
