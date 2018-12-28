package wt

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
		ts, err := t.CreateFromFile(ctx, v, message)
		if err != nil {
			return nil, err
		}
		tid = ts.GetID()
		file := ts.Files[0]
		local, _ := os.Open(v)
		defer local.Close()
		err = t.uploadFile(ctx, tid, local, file)
		if err != nil {
			return nil, err
		}
	}

	return t.Find(ctx, tid)
}

func (t *TransfersService) CreateFromFile(ctx context.Context, path string, message *string) (*Transfer, error) {
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

func (t *TransfersService) uploadFile(ctx context.Context, tid string, localFile *os.File, file *File) error {
	fid := file.GetID()
	name := file.GetName()
	partNum, chunkSize := file.GetMultipartValues()

	reader := bufio.NewReader(localFile)
	errors := NewErrors(fmt.Sprintf(`file %v, %v errors`, fid, name))

	buf := make([]byte, 0, chunkSize)

	var err error
	var n int

	for i := int64(1); i <= partNum; i++ {
		n, err = reader.Read(buf[:chunkSize])
		if n == 0 {
			break
		}
		buf = buf[:n]
		uurl, nerr := t.GetUploadURL(ctx, tid, fid, i)
		if nerr != nil {
			errors.Append(fmt.Errorf(`request upload URL part %v error, %v`, i, nerr.Error()))
			continue
		}
		nerr = t.uploadBytes(ctx, uurl, buf)
		if nerr != nil {
			errors.Append(fmt.Errorf(`upload part %v error, %v`, i, nerr.Error()))
		}
	}

	if err != nil && err != io.EOF {
		return err
	}

	if errors.Len() > 0 {
		return errors
	}

	return nil
}

func (t *TransfersService) uploadBytes(ctx context.Context, uurl *UploadURL, b []byte) error {
	// todo
	return nil
}

func (t *TransfersService) GetUploadURL(ctx context.Context, tid, fid string, partNum int64) (*UploadURL, error) {
	path := fmt.Sprintf("transfers/%s/files/%s/upload-url/%d", tid, fid, partNum)

	req, err := t.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	var uurl UploadURL
	if _, err = t.client.Do(ctx, req, &uurl); err != nil {
		return nil, err
	}

	return &uurl, nil
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
