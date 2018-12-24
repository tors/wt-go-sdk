package wt

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrBlankFiles   = errors.New("blank files")
	ErrBlankContent = errors.New("blank content")
)

// File represents a file object when a successful transfer request
// is issued.
type File struct {
	Multipart *struct {
		PartNumbers *int64 `json:"part_numbers"`
		ChunkSize   *int64 `json:"chunk_size"`
	} `json:"multipart,omitempty"`

	Size *int64  `json:"size"`
	Type *string `json:"type"`
	Name *string `json:"name"`
	ID   *string `json:"id"`
}

// Transfer represents the response when a successful transfer
// request is issued.
type Transfer struct {
	Success   *bool   `json:"success"`
	ID        *string `json:"id"`
	Message   *string `json:"message,omitempty"`
	State     *string `json:"state"`
	ExpiresAt *string `json:"expires_at"`
	URL       *string `json:"url,omitempty"`
	Files     []*File `json:"files"`
}

type TransferService service

// Create informs the API that we want to create a transfer (with at
// least one file). There are no actual files being sent here.
func (t *TransferService) Create(ctx context.Context, tx Transferable) (*Transfer, error) {
	if tx.Len() == 0 {
		return nil, ErrBlankFiles
	}

	param := toTransferableParam(tx)
	fmt.Printf("%+v", param)

	req, err := t.client.NewRequest("POST", "transfers", param)
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{}

	if _, err = t.client.Do(ctx, req, transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// Find retrieves transfer information given an ID.
func (t *TransferService) Find(ctx context.Context, id string) (*Transfer, error) {
	path := fmt.Sprintf("transfers/%v", id)

	req, err := t.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{}
	if _, err = t.client.Do(ctx, req, transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// Transferable describes a transfer payload that is sent to the API
type Transferable interface {
	Message() *string
	Files() []Uploadable
	Len() int
}

// TransferRequest implements the Transferable interface which is used as
// parameter to create a successful transfer.
type TransferRequest struct {
	message *string
	files   []Uploadable
}

func (t *TransferRequest) Message() *string {
	return t.message
}

func (t *TransferRequest) Files() []Uploadable {
	return t.files
}

func (t *TransferRequest) Len() int {
	return len(t.files)
}

func (t *TransferRequest) Add(u Uploadable) {
	t.files = append(t.files, u)
}

func NewTransferRequest(message *string) *TransferRequest {
	return &TransferRequest{
		message: message,
		files:   make([]Uploadable, 0),
	}
}

type transferableParam struct {
	Message *string            `json:"message,omitempty"`
	Files   []*uploadableParam `json:"files"`
}

// Converts transferable to JSON marshalable struct
func toTransferableParam(t Transferable) *transferableParam {
	tx := &transferableParam{
		Message: t.Message(),
		Files:   make([]*uploadableParam, 0),
	}

	for _, f := range t.Files() {
		tx.Files = append(tx.Files, toUploadableParam(f))
	}

	return tx
}
