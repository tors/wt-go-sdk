package wt

import (
	"context"
	"fmt"
	"net/url"
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

// GetID returns the ID field if it is not nil. Otherwise, it returns
// an empty string.
func (t *Transfer) GetID() string {
	if t == nil || t.ID == nil {
		return ""
	}
	return *t.ID
}

// GetURL returns the URL field if it is not nil. Otherwise, it returns
// an empty string.
func (t *Transfer) GetURL() string {
	if t == nil || t.URL == nil {
		return ""
	}
	return *t.URL
}

func (t Transfer) String() string {
	return ToString(t)
}

// File represents a WeTransfer file object transfer response
type File struct {
	Multipart *Multipart `json:"multipart"`
	Size      *int64     `json:"size"`
	Type      *string    `json:"type"`
	Name      *string    `json:"name"`
	ID        *string    `json:"id"`
}

// GetName returns the Name field if it is not nil. Otherwise, it returns
// an empty string.
func (f *File) GetName() string {
	if f == nil || f.Name == nil {
		return ""
	}
	return *f.Name
}

// GetID returns the ID field if it is not nil. Otherwise, it returns
// an empty string.
func (f *File) GetID() string {
	if f == nil || f.ID == nil {
		return ""
	}
	return *f.ID
}

// GetMultipart returns the Multipart field.
func (f *File) GetMultipart() *Multipart {
	if f == nil || f.Multipart == nil {
		return nil
	}
	return f.Multipart
}

func (f File) String() string {
	return ToString(f)
}

// Multipart is info on the chunks of data to be uploaded
type Multipart struct {
	ID          *string `json:"id,omitempty"`
	PartNumbers *int64  `json:"part_numbers"`
	ChunkSize   *int64  `json:"chunk_size"`
}

// GetPartNumbers returns the PartNumbers field.
func (m *Multipart) GetPartNumbers() int64 {
	if m == nil || m.PartNumbers == nil {
		return int64(0)
	}
	return *m.PartNumbers
}

// GetChunkSize returns the ChunkSize field.
func (m *Multipart) GetChunkSize() int64 {
	if m == nil || m.ChunkSize == nil {
		return int64(0)
	}
	return *m.ChunkSize
}

// GetID returns the ID field.
func (m *Multipart) GetID() string {
	if m == nil || m.ID == nil {
		return ""
	}
	return *m.ID
}

func (m Multipart) String() string {
	return ToString(m)
}

// completedTransfer represents a completed file transfer in WeTransfer.
type completedTransfer struct {
	ID        *string `json:"id"`
	Retries   *int64  `json:"retries"`
	Name      *string `json:"name"`
	Size      *int64  `json:"size"`
	ChunkSize *int64  `json:"chunk_size"`
}

func (c completedTransfer) String() string {
	return ToString(c)
}

// TransfersService handles communication with the classic related methods of the
// WeTransfer API
type TransfersService service

// Create attempts to upload data to WeTransfer using S3 as object storage. It
// does the whole ceremony - create a transfer request, get the S3 signed URLs,
// actually upload the file to S3, and complete and finalize the transfer.
//
// Create parameter data types can be string, *os.File, *Buffer, *LocalFile.
// Slices can be passed but will have to be unpacked.
func (t *TransfersService) Create(ctx context.Context, message *string, up ...Uploadable) (*Transfer, error) {
	if len(up) == 0 {
		return nil, fmt.Errorf("empty files")
	}

	// `filemap` keys are file names. We need this mapping to get the
	// actual file or buffer easily when we receive response from the transfer
	// request.
	filemap := make(map[string]Uploadable)

	for _, f := range up {
		name, _ := f.Stat()
		filemap[name] = f
	}

	// Create a transfer object. Note that this does not upload the file or buffer.
	transfer, err := t.createTransfer(ctx, message, up...)
	if err != nil {
		return nil, err
	}

	var errs []error

	// Once we have the files that have been acknowledged by WeTransfer, we
	// map the files with our filemap so we begin the actual uploading.
	for _, f := range transfer.Files {
		name := f.GetName()
		if tx, ok := filemap[name]; ok {
			ft := newFileTransfer(tx, f)
			err = t.client.uploader.upload(ctx, transfer, ft)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Do not complete and finalize the transfer if there are errors
	if len(errs) > 0 {
		return nil, joinErrors(errs, nil)
	}

	// Complete the transfer since there are no errors
	_, err = t.complete(ctx, transfer)
	if err != nil {
		return nil, err
	}

	return t.finalize(ctx, transfer.GetID())
}

// createTransfer returns a transfer object after submitting a new transfer
// request to the API
func (t *TransfersService) createTransfer(ctx context.Context, message *string, up ...Uploadable) (*Transfer, error) {
	var fs []fileObject

	for _, obj := range up {
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

// complete informs WeTransfer that all the uploading for our files is done.
// When the files are uploaded to S3, WeTransfer has no way of determining if
// the transfer is successful or not. After the call to the endpoint is made,
// this method returns the list of completed transfer responses which length is
// equal to the number of files specified in the transfer request.
func (t *TransfersService) complete(ctx context.Context, tx *Transfer) ([]*completedTransfer, error) {
	completed := make([]*completedTransfer, 0)

	var errs []error

	tid := url.PathEscape(tx.GetID())

	for _, file := range tx.Files {
		fid := url.PathEscape(file.GetID())
		path := fmt.Sprintf("transfers/%v/files/%v/upload-complete", tid, fid)
		partNum := file.Multipart.GetPartNumbers()
		req, err := t.client.NewRequest("PUT", path, &struct {
			PartNumbers int64 `json:"part_numbers"`
		}{
			PartNumbers: partNum,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var ct completedTransfer
		if _, err = t.client.Do(ctx, req, &ct); err != nil {
			errs = append(errs, err)
		}
		completed = append(completed, &ct)
	}

	if len(errs) > 0 {
		errmsg := fmt.Sprintf("completing transfer %v, with %v error(s)", tx.GetID(), len(errs))
		return nil, joinErrors(errs, &errmsg)
	}

	return completed, nil
}

// Finalize closes a transfer for modification rendering it immutable and downloadable.
func (t *TransfersService) finalize(ctx context.Context, id string) (*Transfer, error) {
	path := fmt.Sprintf("transfers/%v/finalize", url.PathEscape(id))

	req, err := t.client.NewRequest("PUT", path, nil)
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{}
	if _, err = t.client.Do(ctx, req, transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// Find retrieves the transfer object given an ID.
func (t *TransfersService) Find(ctx context.Context, id string) (*Transfer, error) {
	path := fmt.Sprintf("transfers/%v", url.PathEscape(id))

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
