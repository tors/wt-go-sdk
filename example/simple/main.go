package main

import (
	"fmt"

	"github.com/tors/wt-go-sdk/wt"
)

func main() {
	var apiKey string

	fmt.Print("Enter API key: ")
	fmt.Scanf("%s", &apiKey)

	client, err := wt.NewAuthorizedClient(apiKey, nil)
	checkErr(err)

	resp, err := client.Transfer.Create()
	checkErr(err)

	fmt.Printf("%v\n", resp)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
