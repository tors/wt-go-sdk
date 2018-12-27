package wt

import (
	"context"
	"fmt"
	"sync"
)

// Multipart is info on the chunks of data to be uploaded
type Multipart struct {
	PartNumbers *int64 `json:"part_numbers"`
	ChunkSize   *int64 `json:"chunk_size"`
}

func (m Multipart) String() string {
	return ToString(m)
}

// RemoteFile represents a WeTransfer file object transfer response
type RemoteFile struct {
	Multipart *Multipart `json:"multipart"`
	Size      *int64     `json:"size"`
	Type      *string    `json:"type"`
	Name      *string    `json:"name"`
	ID        *string    `json:"id"`
}

// PartNumbers retrieves part numbers of a multipart file. It returns 0
// when multipart is nil
func (r RemoteFile) GetPartNumbers() int64 {
	if r.Multipart == nil || r.Multipart.PartNumbers == nil {
		return 0
	}
	return *r.Multipart.PartNumbers
}

func (r RemoteFile) GetID() string {
	if r.ID == nil {
		return ""
	}
	return *r.ID
}

func (r RemoteFile) String() string {
	return ToString(r)
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

func (t Transfer) String() string {
	return ToString(t)
}

// UploadURL represents the response once a request for the URL destination of
// the local file
type UploadURL struct {
	Success *bool   `json:"success"`
	URL     *string `json:"url"`
	tid     string  `json:"-"`
	fid     string  `json:"-"`
	partNum int64   `json:"-"`
	err     error   `json:"-"`
}

func (u *UploadURL) SetError(err error) {
	u.err = err
}

func (u UploadURL) GetError() error {
	return u.err
}

func (u UploadURL) String() string {
	return ToString(u)
}

// TransfersService handles communication with the classic related methods of the
// WeTransfer API
type TransfersService service

// Create informs the API that we want to create a transfer (with at
// least one file). There are no actual files being sent here.
func (t *TransfersService) Create(ctx context.Context, message *string, fo []*FileObject) (*Transfer, error) {
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

// GetAllUploadURL retrieves all upload URLs as remote destination addresses
// for the file. A file can be divided into parts and is uploaded by chunks.
// Each part has it's own destination address which is a presigned S3 URL. The
// S3 URL will expire after 1 hour.
func (t *TransfersService) GetAllUploadURL(ctx context.Context, tid string, file *RemoteFile) []*UploadURL {
	var all []*UploadURL

	var wg sync.WaitGroup
	mux := &sync.Mutex{}

	for no := int64(1); no <= file.GetPartNumbers(); no++ {
		wg.Add(1)
		go func(n int64) {
			defer wg.Done()
			uurl := t.GetUploadURL(ctx, tid, file.GetID(), n)
			mux.Lock()
			all = append(all, uurl)
			mux.Unlock()
		}(no)
	}

	wg.Wait()

	return all
}

// GetUploadURL retrieves an upload URL as remote destination address for the
// file. The URL that is returned comes in a form of presigned S3 URL which
// is only valid for an hour.
func (t *TransfersService) GetUploadURL(ctx context.Context, tid, fid string, partNum int64) *UploadURL {
	path := fmt.Sprintf("transfers/%s/files/%s/upload-url/%d", tid, fid, partNum)
	req, err := t.client.NewRequest("POST", path, nil)
	if err != nil {
		uurl := &UploadURL{tid: tid, fid: fid, partNum: partNum}
		uurl.SetError(err)
		return uurl
	}
	var uurl UploadURL
	if _, err = t.client.Do(ctx, req, &uurl); err != nil {
		uurl = UploadURL{tid: tid, fid: fid, partNum: partNum}
		uurl.SetError(err)
	}
	return &uurl
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
