package wt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

// Uploadable describes a buffer or a file that can be uploaded to WeTransfer.
type Uploadable interface {
	Stat() (string, int64)
}

// LocalFile implements the Uploadable interface. It represents
// a file on disk to be sent as a file transfer.
type LocalFile struct {
	name     string
	size     int64
	filepath string
}

// Stat returns the info (name and size) of the uploadable. If there's
// an error, it will be type *PathError.
func (l *LocalFile) Stat() (string, int64) {
	return l.name, l.size
}

// NewLocalFile returns a LocalFile if file exists given the filepath.
func NewLocalFile(filepath string) (*LocalFile, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	return &LocalFile{
		name:     sanitizeString(info.Name()),
		size:     info.Size(),
		filepath: filepath,
	}, nil
}

// Buffer implements the Uploadable interface. It represents a buffered data
// (usually created on the fly) to be sent as a file object.
type Buffer struct {
	name   string
	buffer []byte
}

// Stat returns returns the name and the size of the uploadable buffer.
func (b *Buffer) Stat() (string, int64) {
	return b.name, int64(len(b.buffer))
}

// GetBytes returns the b field which represents data.
func (b *Buffer) GetBytes() []byte {
	return b.buffer
}

// NewBuffer returns a new Buffer given a string and a slice of bytes.
func NewBuffer(name string, b []byte) *Buffer {
	return &Buffer{
		name:   name,
		buffer: b,
	}
}

// File or item response from boards or transfers.
type fileItem interface {
	GetID() string
	GetName() string
	GetMultipart() *Multipart
}

// fileTransfer wraps a buffer or file and the response object of
// upload request to return necessary data including io.Reader for a
// multipart upload.
type fileTransfer struct {
	up   Uploadable
	file fileItem
}

func (f *fileTransfer) getID() string {
	return f.file.GetID()
}

func (f *fileTransfer) getName() string {
	return f.file.GetName()
}

func (f *fileTransfer) stat() (id string, partNumbers int64, chunkSize int64) {
	id, partNumbers, chunkSize = "", 0, 0

	if f == nil || f.file == nil {
		return id, partNumbers, chunkSize
	}

	m := f.file.GetMultipart()
	if m == nil {
		return id, partNumbers, chunkSize
	}

	return m.GetID(), m.GetPartNumbers(), m.GetChunkSize()
}

func (f *fileTransfer) reader() (io.Reader, *os.File, error) {
	switch v := f.up.(type) {
	case *LocalFile:
		local, err := os.Open(v.filepath)
		return bufio.NewReader(local), local, err
	case *Buffer:
		return bytes.NewReader(v.buffer), nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported Uploadable source")
	}
}

func newFileTransfer(up Uploadable, file fileItem) *fileTransfer {
	return &fileTransfer{
		up:   up,
		file: file,
	}
}

// fileObject represents the parameter serialized in JSON format to be sent to
// create a file transfer.
type fileObject struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// toFileObject converts a Uploadable into a serializable file object.
func toFileObject(t Uploadable) fileObject {
	name, size := t.Stat()
	return fileObject{
		Name: name,
		Size: size,
	}
}
