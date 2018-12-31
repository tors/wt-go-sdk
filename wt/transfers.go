package wt

import (
	"context"
	"fmt"
	"os"
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

// GetID returns the ID field if it is not nil. Otherwise, it returns
// an empty string.
func (t *Transfer) GetID() string {
	if t == nil || t.ID == nil {
		return ""
	}
	return *t.ID
}

func (t Transfer) String() string {
	return ToString(t)
}

// CompletedTransfer represents a completed file transfer in
// WeTransfer. This step is required after successfully sending the
// files to S3.
type CompletedTransfer struct {
	ID        *string `json:"id"`
	Retries   *int64  `json:"retries"`
	Name      *string `json:"name"`
	Size      *int64  `json:"size"`
	ChunkSize *int64  `json:"chunk_size"`
}

func (c CompletedTransfer) String() string {
	return ToString(c)
}

// TransfersService handles communication with the classic related methods of the
// WeTransfer API
type TransfersService service

// Create attempts to upload data to WeTransfer using S3 as object storage. It
// does the whole ceremony - create a transfer request, get the S3 signed URLs,
// actually upload the file to S3, and complete and finalize the upload.
//
// Create parameter data types can be string, *os.File, *Buffer, *BufferedFile.
// Slices can be passed but will have to be unpacked.
func (t *TransfersService) Create(ctx context.Context, message *string, in ...interface{}) (*Transfer, error) {
	if in == nil {
		return nil, fmt.Errorf("empty files")
	}

	tx := make([]Transferable, len(in))

	for i, obj := range in {
		switch v := obj.(type) {
		case string, *os.File:
			buf, err := BuildBufferedFile(v)
			if err != nil {
				return nil, err
			}
			tx[i] = buf
		case *Buffer:
			tx[i] = (*Buffer)(v)
		case *BufferedFile:
			tx[i] = (*BufferedFile)(v)
		default:
			return nil, fmt.Errorf(`allowed types are string string *Buffer *BufferedFile`)
		}
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

// Complete informs WeTransfer that all the uploading for our files is done.
// When the files are uploaded to S3, WeTransfer has no way of determining if
// the transfer is successful or not. After the call to the endpoint is made,
// this method returns the list of completed transfer responses which length is
// equal to the number of files specified in the transfer request.
func (t *TransfersService) Complete(ctx context.Context, tx *Transfer) ([]*CompletedTransfer, error) {
	tid := tx.GetID()
	errors := NewErrors(fmt.Sprintf("complete transfer %v errors", tid))
	completed := make([]*CompletedTransfer, 0)

	for _, file := range tx.Files {
		fid := file.GetID()
		path := fmt.Sprintf("transfers/%v/files/%v/upload-complete", tx.GetID(), fid)
		partNum := file.Multipart.GetPartNumbers()
		req, err := t.client.NewRequest("PUT", path, &struct {
			PartNumbers int64 `json:"file_numbers"`
		}{
			PartNumbers: partNum,
		})
		if err != nil {
			errors.Append(fmt.Errorf(`file %v completion request error: %v`, fid, err.Error()))
			continue
		}
		var ct CompletedTransfer
		if _, err = t.client.Do(ctx, req, &ct); err != nil {
			errors.Append(fmt.Errorf(`file %v completion error: %v`, fid, err.Error()))
		}
		completed = append(completed, &ct)
	}

	if errors.Len() > 0 {
		return nil, errors
	}

	return completed, nil
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
