// +build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/tors/wt-go-sdk/wt"
)

var (
	client  *wt.Client
	keyName = []byte("WETRANSFER_API_TOKEN")
)

func init() {
	apiKey, err := readToken(".env")
	if err != nil {
		log.Fatal(err)
	}
	logf(`Using key "%v"`, apiKey)
	client, err = wt.NewAuthorizedClient(context.Background(), apiKey, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func readToken(file) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(b[:len(keyName)], keyName) {
		return "", fmt.Errorf("Key must be %v", string(keyName))
	}
	return string(bytes.Trim(b[len(keyName)+1:], "\n")), nil
}

func logf(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}
