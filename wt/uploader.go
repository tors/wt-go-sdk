package wt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// identifiable describes an object that can return an id
type identifiable interface {
	GetID() string
}

// UploadURL represents a response of an upload URL retrieval request
type UploadURL struct {
	Success *bool   `json:"success"`
	URL     *string `json:"url"`
}

// GetURL returns the URL field if it is not nil. Otherwise, it returns
// an empty string.
func (u *UploadURL) GetURL() string {
	if u == nil || u.URL == nil {
		return ""
	}
	return *u.URL
}

func (u UploadURL) String() string {
	return ToString(u)
}

// uploaderService is a common file upload service for transfers
// and boards
type uploaderService service

// upload attempts to upload a file or a buffer. It does so by using the response
// from the transfer request which has the multipart info. This info is used to
// upload the file or buffer in chunks if needed.
func (u *uploaderService) upload(ctx context.Context, idx identifiable, ft *fileTransfer) error {
	fid := ft.getID()
	name := ft.getName()
	partNum, chunkSize := ft.getMulipartValues()

	reader, rerr := ft.getReader()
	if rerr != nil {
		return rerr
	}

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
		uurl, nerr := u.getUploadURL(ctx, idx, fid, i)
		if nerr != nil {
			errors.Append(fmt.Errorf(`request upload URL part %v error: %v`, i, nerr.Error()))
			continue
		}
		nerr = uploadBytes(ctx, uurl, buf)
		if nerr != nil {
			errors.Append(fmt.Errorf(`upload part %v error: %v`, i, nerr.Error()))
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

func (u *uploaderService) getUploadURL(ctx context.Context, idx identifiable, fid string, partNum int64) (*UploadURL, error) {
	var pathPrefix string

	switch idx.(type) {
	case *Transfer:
		pathPrefix = "transfers"
	case *Board:
		pathPrefix = "boards"
	default:
		return nil, fmt.Errorf("identifiable type not supported")
	}

	path := fmt.Sprintf("%s/%s/files/%s/upload-url/%d", pathPrefix, idx.GetID(), fid, partNum)

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

func uploadBytes(ctx context.Context, uurl *UploadURL, b []byte) error {
	url := uurl.GetURL()

	if url == "" {
		return fmt.Errorf("blank URL entry")
	}

	if len(b) == 0 {
		return fmt.Errorf("blank data")
	}

	reader := bytes.NewReader(b)

	req, err := http.NewRequest("PUT", url, reader)
	if err != nil {
		return err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return err
	}

	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("upload bytes error in %v %v: %d %v",
		r.Request.Method, r.Request.URL,
		r.StatusCode, string(data),
	)
}
