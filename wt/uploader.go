package wt

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// boardOrTransfer describes either a Transfer or a Board object
type boardOrTransfer interface {
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
func (u *uploaderService) upload(ctx context.Context, bot boardOrTransfer, ft *fileTransfer) error {
	fid := ft.getID()
	mid, partNum, chunkSize := ft.getMulipartValues()

	reader, rerr := ft.getReader()
	if rerr != nil {
		return rerr
	}

	var errs []error

	buf := make([]byte, 0, chunkSize)

	for i := int64(1); i <= partNum; i++ {
		n, err := reader.Read(buf[:chunkSize])
		if err != nil && err != io.EOF {
			errs = append(errs, err)
			break
		}
		buf = buf[:n]
		uurl, err := u.getUploadURL(ctx, bot, fid, i, mid)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = uploadBytes(ctx, uurl, buf)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		errmsg := fmt.Sprintf("upload %v failed with %v error(s)", bot.GetID(), len(errs))
		return joinErrors(errs, &errmsg)
	}

	return nil
}

// getUploadURL retrieves an upload url given if it's a board or a transfer, a
// file id, a part number and corresponding multipart ID if it's item response.
func (u *uploaderService) getUploadURL(ctx context.Context, bot boardOrTransfer, fid string, partNum int64, mid string) (*UploadURL, error) {
	var path string

	id := url.PathEscape(bot.GetID())
	fid = url.PathEscape(fid)

	switch bot.(type) {
	case *Transfer:
		path = fmt.Sprintf("transfers/%s/files/%s/upload-url/%d", id, fid, partNum)
	case *Board:
		path = fmt.Sprintf("boards/%s/files/%s/upload-url/%d/%v", id, fid, partNum, mid)
	default:
		return nil, fmt.Errorf("boardOrTransfer type not supported")
	}

	req, err := u.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	var uurl UploadURL
	if _, err = u.client.Do(ctx, req, &uurl); err != nil {
		return nil, err
	}

	return &uurl, nil
}

func uploadBytes(ctx context.Context, uurl *UploadURL, b []byte) error {
	url := uurl.GetURL()

	if url == "" {
		return fmt.Errorf("blank URL")
	}

	if len(b) == 0 {
		return fmt.Errorf("blank data for URL: %v", uurl)
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

	br := bufio.NewReader(r.Body)

	buf := make([]byte, 0, 512*1024)

	n, err := br.Read(buf[:cap(buf)])
	if err != nil && err != io.EOF {
		return err
	}

	return fmt.Errorf("upload bytes error %v %v: %d %v",
		r.Request.Method, r.Request.URL,
		r.StatusCode, string(buf[:n]),
	)
}
