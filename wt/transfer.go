package wt

import (
	"context"
	"errors"
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

func (t *TransferService) Create(ctx context.Context, param *TransferParam) (*Transfer, error) {
	if len(param.Files) == 0 {
		return nil, fmt.Errorf("Files must not be empty")
	}

	req, err := t.client.NewRequest("POST", "transfers", param)

	if err != nil {
		return nil, err
	}

	transfer := &Transfer{}

	_, err = t.client.Do(ctx, req, transfer)
	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (t *TransferService) Find() (*Transfer, error) {
	return nil, errors.New("not implemented error")
}
