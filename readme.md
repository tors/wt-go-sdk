# WeTransfer Go SDK
[![GoDoc](https://godoc.org/github.com/tors/wt-go-sdk/wt?status.svg)](https://godoc.org/github.com/tors/wt-go-sdk/wt) [![Build Status](https://travis-ci.org/tors/wt-go-sdk.svg?branch=master)](https://travis-ci.org/tors/wt-go-sdk) [![Go Report Card](https://goreportcard.com/badge/github.com/tors/wt-go-sdk)](https://goreportcard.com/report/github.com/tors/wt-go-sdk)

This is an unofficial WeTransfer Go SDK.

**Status**: _v1.0.0-rc.1_

If you find bugs, please file an [issue here](https://github.com/tors/wt-go-sdk/issues). Thank you!

## Installation

Get the SDK with:

```bash
go get -v github.com/tors/wt-go-sdk
```

## Getting started

To get started, you'll need an API key. You can get one from WeTransfer's
[Developers Portal](https://developers.wetransfer.com/).

It's a good practice not to hardcode any private keys. In the event that you
think your API key might be compromised, you can revoke it from within the
[developer portal](https://developers.wetransfer.com/).

### Creating a client

You'll need to create an authorized client which automatically requests for a
JWT token. This token is used in subsequent requests to the server.

```go
apiKey := "<your-api-key>"
ctx := context.Background()
client, err := wt.NewAuthorizedClient(ctx, apiKey, nil)
```

For subsequent authorized requests, you'll need to pass a
[context](https://golang.org/pkg/context).

## Transfers

A transfer is a collection of files that can be created once and downloaded
until it expires. The expiry is set to 7 days and the maximum size of data per
transfer is 2GB. Files are immutable once successfully uploaded.

### Create transfers

You can use `*Buffer`, and `*LocalFile` types as file objects to create a
transfer.

```go
// Buffer
ponya := wt.NewBuffer("pony.txt", []byte("yeehaaa"))

// LocalFile
ponyb := wt.NewLocalFile("disk/pony.txt")

client.Transfers.Create(ctx, &message, ponya, ponyb)
```

`Transfers.Create` does the necessary steps to actually transfer the file.
Internally it does the whole ritual - _create new transfer_, _request for upload
URLs_, _actual file upload to S3_, _complete the upload_, and _finalize_ it.

#### Uploadable slices

`Transfers.Create` is a variadic function that accepts structs that implement
the `Uploadable` interface.

```go
var pets []Uploadable

pony := wt.NewBuffer("pony.txt", []byte("yeehaaa"))
kitten, _ := wt.NewLocalFile("disk/kitten.txt")

pets = append(pets, pony, kitten)

client.Transfers.Create(ctx, &message, pets...)
```

### Find a transfer

```go
transfer, _ := client.Transfers.Find(ctx, "transfer-id")
fmt.Println(transfer.Files)
```

## Boards

A board is collection of items that can be links or traditional files. Unlike
`transfers`, boards' items do not have explicit expiry time as long as they
receive activity. If untouched, they expire after 3 months.

### Create boards

To create a board, a name is required. You can also pass a description which is
optional. New boards have 0 items.

```go
board, err := client.Boards.Create(ctx, "My kittens", nil)
fmt.Println(board.GetID())
```

### Add links to a board

In WeTransfer context, a link has two fields - a url which is required
and a title which is optional. Links must added to existing boards.

```go
// Add single link
desc := "Pony wiki"
pony, _ := wt.NewLink("https://en.wikipedia.org/wiki/Pony", &desc)
board, _ := client.Boards.AddLinks(ctx, board, pony)

// Add multiple
links := []*wt.Link{
    pony,
    &Link{
        URL: "https://en.wikipedia.org/wiki/Kitten"
    },
}
board, _ := client.Boards.AddLinks(ctx, board, links...)
fmt.Println(board.Items)
```

### Add files to a board

Files can be added to existing boards too. The way files are uploaded in boards
is the same as how files are uploaded in trasfers. The only difference is that
you can group files using the board and add more files in the future if needed.

```go
pony := wt.NewBuffer("pony.txt", []byte("yehaaa"))
kitten := wt.NewBuffer("kitten.txt", []byte("meowww"))
narwhal, _ := wt.NewLocalFile("disk/narwhal.txt")
unicorn, _ := wt.NewLocalFile("disk/unicorn.txt")

client.Boards.AddFiles(ctx, board, pony, kitten, narwhal, unicorn)

// Pass a slice like Transfers

withHornsOnly := append(pets[:0], narwhal, unicorn)
client.Boards.AddFiles(ctx, board, withHornsOnly...)
```

### Find a board

```go
board, _ := client.Boards.Find(ctx, "board-id")
fmt.Println(board.Items)
```

## Testing

There are 2 types of test suites in this library - unit and integration. The
unit tests do not make network calls so it should run pretty fast.
Integration tests on the other hand verify that this SDK is coded in such a way
that it follows the actual behavior of the live API. Actual data is sent over
to the live API and if there are incompatibilities, integration tests should
hopefully be able to catch those.

### Running the tests

Run the unit tests...

```bash
make test
```

To run the integration tests, you'll need to get an API key and create an .env
file like so:

```bash
echo WETRANSFER_API_TOKEN=key > integration/.env
```

Once you've done that, you can now run the integration test suite.

```bash
make integration
```

### Helpful Links
- [Examples](https://github.com/tors/wt-go-sdk/tree/master/example)
- [Documentation](https://godoc.org/github.com/tors/wt-go-sdk/wt)
- [WeTransfer API in Github](https://wetransfer.github.io/wt-api-docs/index.html)
