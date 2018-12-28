package wt

import (
	"context"
	"fmt"
	"os"
)

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

func (t *Transfer) GetID() string {
	if t == nil || t.ID == nil {
		return ""
	}
	return *t.ID
}

func (t Transfer) String() string {
	return ToString(t)
}

// TransferRequest represents the parameters to create a transfer
type TransferRequest struct {
	Message *string       `json:"message"`
	Files   []*FileObject `json:"files"`
}

// AppendAsFileObject creates a new file object and appends it into
// the file objects array
func (t *TransferRequest) AppendAsFileObject(name string, size int64) {
	f := NewFileObject(name, size)
	if f != nil {
		t.Files = append(t.Files, f)
	}
}

func NewTransferRequest(message *string) *TransferRequest {
	return &TransferRequest{
		Message: message,
	}
}

// UploadURL represents the response once a request for the URL destination of
// the local file
type UploadURL struct {
	Success *bool   `json:"success"`
	URL     *string `json:"url"`
}

func (u UploadURL) String() string {
	return ToString(u)
}

// TransfersService handles communication with the classic related methods of the
// WeTransfer API
type TransfersService service

// Create informs the API that we want to create a transfer (with at
// least one file).
func (t *TransfersService) Create(ctx context.Context, in interface{}, message *string) (*Transfer, error) {
	var tid string
	switch v := in.(type) {
	case string:
		ts, err := t.createTransfer(ctx, v, message)
		if err != nil {
			return nil, err
		}
		tid = ts.GetID()
		file := ts.Files[0]
		local, _ := os.Open(v)
		defer local.Close()
		err = t.client.uploader.uploadFile(ctx, ts, local, file)
		if err != nil {
			return nil, err
		}
	}

	return t.Find(ctx, tid)
}

func (t *TransfersService) createTransfer(ctx context.Context, path string, message *string) (*Transfer, error) {
	name, size, err := fileInfo(path)
	if err != nil {
		return nil, err
	}

	trq := NewTransferRequest(message)
	trq.AppendAsFileObject(name, size)

	req, err := t.client.NewRequest("POST", "transfers", trq)
	if err != nil {
		return nil, err
	}

	var ts Transfer
	if _, err = t.client.Do(ctx, req, &ts); err != nil {
		return nil, err
	}

	return &ts, nil
}

// Find retrieves transfer information given an ID.
func (t *TransfersService) Find(ctx context.Context, id string) (*Transfer, error) {
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
