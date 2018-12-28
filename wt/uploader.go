package wt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

// identifiable describes an object that can return an id
type identifiable interface {
	GetID() string
}

// uploaderService is a common file upload service for transfers
// and boards
type uploaderService service

func (u *uploaderService) uploadFile(ctx context.Context, id identifiable, l *os.File, f *File) error {
	fid := f.GetID()
	name := f.GetName()
	partNum, chunkSize := f.GetMultipartValues()

	reader := bufio.NewReader(l)
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
		uurl, nerr := u.getUploadURL(ctx, id, fid, i)
		if nerr != nil {
			errors.Append(fmt.Errorf(`request upload URL part %v error, %v`, i, nerr.Error()))
			continue
		}
		nerr = u.uploadBytes(ctx, uurl, buf)
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

func (t *uploaderService) uploadBytes(ctx context.Context, uurl *UploadURL, b []byte) error {
	return nil
}

func (t *uploaderService) getUploadURL(ctx context.Context, id identifiable, fid string, partNum int64) (*UploadURL, error) {
	var pathPrefix string

	switch id.(type) {
	case *Transfer:
		pathPrefix = "transfers"
	case *Board:
		pathPrefix = "boards"
	default:
		return nil, fmt.Errorf("identifiable type not supported")
	}

	path := fmt.Sprintf("%s/%s/files/%s/upload-url/%d", pathPrefix, id.GetID(), fid, partNum)

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
