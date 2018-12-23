package wt

import (
	"context"
	"fmt"
)

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

type Transfer struct {
	Success   *bool   `json:"success"`
	ID        *string `json:"id"`
	Message   *string `json:"message,omitempty"`
	State     *string `json:"state"`
	ExpiresAt *string `json:"expires_at"`
	URL       *string `json:"url,omitempty"`
	Files     []*File `json:"files"`
}

type TransferParam struct {
	Message string `json:"message"`
	Files   []M    `json:"files"`
}

func (t *TransferParam) AddFile(name string, size int64) {
	t.Files = append(t.Files, M{
		"name": name,
		"size": size,
	})
}

func NewTransferParam(message string) *TransferParam {
	return &TransferParam{
		Message: message,
		Files:   make([]M, 0),
	}
}

type TransferService service

// Create informs the API that we want to create a transfer (with at
// least one file). There are no actual files being sent here.
func (t *TransferService) Create(ctx context.Context, param *TransferParam) (*Transfer, error) {
	if len(param.Files) == 0 {
		return nil, fmt.Errorf("Files must not be empty")
	}

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
