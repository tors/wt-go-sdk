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

	message := "My first transfer!"

	object, _ := wt.FromString("abc", "abc.txt")
	fo := []*wt.FileObject{object}

	ctx := context.Background()

	resp, err := client.Transfer.Create(ctx, &message, fo)
	checkErr(err)

	fmt.Printf("%v\n", resp)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
