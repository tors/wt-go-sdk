# WeTransfer Go SDK
[![GoDoc](https://godoc.org/github.com/tors/wt-go-sdk/wt?status.svg)](https://godoc.org/github.com/tors/wt-go-sdk/wt) [![Build Status](https://travis-ci.org/tors/wt-go-sdk.svg?branch=master)](https://travis-ci.org/tors/wt-go-sdk)

This is an unofficial WeTransfer Go SDK.

**Status**: Work in progress. Do not use in production.

## Installation

Get the SDK with:

```bash
go get -v github.com/tors/wt-go-sdk
```

## Getting started

In order to be able to use the SDK and access our public APIs, you must provide
an API key, which is available in our [Developers
Portal](https://developers.wetransfer.com/).

This is the bare minimum code needed to create a transfer...

```go
// main.go
package main

import (
  "context"
  "fmt"

  "github.com/tors/wt-go-sdk/wt"
)

func main() {
  apiKey := "<your-api-key>"
  client, _ := wt.NewAuthorizedClient(apiKey, nil)

  param := wt.NewTransferParam("First transfer.")
  param.AddFile("big-bobis.jpg", 195906)

  ctx := context.Background()

  resp, _ := client.Transfer.Create(ctx, param)
  fmt.Printf("%+v\n", resp)
}
```
