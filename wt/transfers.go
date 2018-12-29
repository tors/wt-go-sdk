package wt

import (
	"context"
	"fmt"
)

// Transferable describes a file object in WeTransfer
type Transferable interface {
	GetName() string
	GetSize() int64
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

func (t *Transfer) GetID() string {
	if t == nil || t.ID == nil {
		return ""
	}
	return *t.ID
}

func (t Transfer) String() string {
	return ToString(t)
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

// Create uploads a file, set of file, and buffered data. Uploadable data
// includes these types:
//
//	string - *BufferedFile is created from this string
//	[]string - []*BufferedFile is created from the string slice
//	*Buffer
//	[]*Buffer
//	*BufferedFile
//	[]*BufferedFile
func (t *TransfersService) Create(ctx context.Context, message *string, in interface{}) (*Transfer, error) {
	if in == nil {
		return nil, fmt.Errorf("empty files")
	}

	tx := make([]Transferable, 0)

	switch v := in.(type) {
	case string:
		buf, err := NewBufferedFile(v)
		if err != nil {
			return nil, err
		}
		tx = append(tx, buf)
	case []string:
		var err error
		var buf *BufferedFile
		for _, p := range v {
			buf, err = NewBufferedFile(p)
			if err != nil {
				break
			}
			tx = append(tx, buf)
		}
		if err != nil {
			return nil, err
		}
	case *Buffer:
		tx = append(tx, v)
	case []*Buffer:
		for _, b := range v {
			tx = append(tx, b)
		}
	case *BufferedFile:
		tx = append(tx, v)
	case []*BufferedFile:
		for _, b := range v {
			tx = append(tx, b)
		}
	default:
		return nil, fmt.Errorf(`allowed types are string []string *Buffer []*Buffer *BufferedFile []*BufferedFile`)
	}

	return t.createTransfer(ctx, message, tx...)
}

// createTransfer returns a transfer object after submitting a new transfer
// request to the API
func (t *TransfersService) createTransfer(ctx context.Context, message *string, tx ...Transferable) (*Transfer, error) {
	var fs []fileObject

	for _, obj := range tx {
		fs = append(fs, toFileObject(obj))
	}

	req, err := t.client.NewRequest("POST", "transfers", &struct {
		Message *string      `json:"message"`
		Files   []fileObject `json:"files"`
	}{
		Message: message,
		Files:   fs,
	})

	if err != nil {
		return nil, err
	}

	var ts Transfer
	if _, err = t.client.Do(ctx, req, &ts); err != nil {
		return nil, err
	}

	return &ts, nil
}

// Find retrieves the transfer object given an ID.
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
