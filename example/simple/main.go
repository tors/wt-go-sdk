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

	client, err := wt.NewAuthorizedClient(apiKey, nil)
	checkErr(err)

	param := wt.NewTransferParam("First transfer.")
	param.AddFile("big-bobis.jpg", 195906)

	ctx := context.Background()

	resp, err := client.Transfer.Create(ctx, param)
	checkErr(err)

	fmt.Printf("%v\n", resp)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
