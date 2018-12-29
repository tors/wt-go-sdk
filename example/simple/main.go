package main

import (
	"context"
	"fmt"

	"github.com/tors/wt-go-sdk/wt"
)

func main() {
	var apiKey string

	fmt.Print("Enter API key: ")
	fmt.Scanf("%s", &apiKey)

	ctx := context.Background()

	client, err := wt.NewAuthorizedClient(ctx, apiKey, nil)
	checkErr(err)

	message := "My first pony!"
	buffer := wt.NewBuffer("pony.txt", []byte("yeehaaa"))

	transfer, err := client.Transfers.Create(ctx, &message, buffer)
	checkErr(err)

	fmt.Println(transfer.String())
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
