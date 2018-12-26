package wt

import (
	"context"
	"fmt"
)

// Multipart is info on the chunks of data to be uploaded
type Multipart struct {
	PartNumbers *int64 `json:"part_numbers"`
	ChunkSize   *int64 `json:"chunk_size"`
}

// RemoteFile represents a WeTransfer file object transfer response
type RemoteFile struct {
	Multipart *Multipart `json:"multipart,omitempty"`
	Size      *int64     `json:"size"`
	Type      *string    `json:"type"`
	Name      *string    `json:"name"`
	ID        *string    `json:"id"`
}

// Transfer represents the response when a successful transfer
// request is issued.
type Transfer struct {
	Success   *bool         `json:"success"`
	ID        *string       `json:"id"`
	Message   *string       `json:"message,omitempty"`
	State     *string       `json:"state"`
	ExpiresAt *string       `json:"expires_at"`
	URL       *string       `json:"url,omitempty"`
	Files     []*RemoteFile `json:"files"`
}

// TransferService handles communication with the classic related methods of the
// WeTransfer API
type TransferService service

// Create informs the API that we want to create a transfer (with at
// least one file). There are no actual files being sent here.
func (t *TransferService) Create(ctx context.Context, message *string, fo []*FileObject) (*Transfer, error) {
	if len(fo) == 0 {
		return nil, ErrBlankFiles
	}

	tr := newTransferRequest(message, fo)

	req, err := t.client.NewRequest("POST", "transfers", tr)
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

// transferRequest specifies the parameters to create a transfer request
type transferRequest struct {
	Message *string            `json:"message"`
	Files   []*fileObjectParam `json:"files"`
}

func newTransferRequest(message *string, fo []*FileObject) *transferRequest {

	var objects []*fileObjectParam

	for _, o := range fo {
		objects = append(objects, toFileObjectParam(o))
	}

	return &transferRequest{
		Message: message,
		Files:   objects,
	}
}
