# WeTransfer Go SDK

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

This is the bare minimum code needed to create a transfer. Copy and paste into
a file, place your API Key there, and run with `node path/to/file.js`. Voilà,
you just created your very first transfer!

```go
// main.go
package main

import (
  "fmt"
  "github.com/tors/wt-go-sdk/wt"
)

func main() {
  apiKey := "<your-api-key>"
  client, _ := wt.NewAuthorizedClient(apiKey, nil)
  resp, _ := client.Transfer.Create()
  fmt.Printf("%+v\n", resp)
}
```
