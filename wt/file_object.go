package wt

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

var (
	ErrBlankFiles = errors.New("blank files")
	ErrBlankSize  = errors.New("blank size")
	ErrBlankName  = errors.New("blank name")
)

// FileObject represents a file object in WeTransfer
type FileObject struct {
	name   string
	size   int64
	reader io.Reader
}

func (f *FileObject) Size() int64 {
	return f.size
}

func (f *FileObject) Name() string {
	return f.name
}

func (f *FileObject) Reader() io.Reader {
	return f.reader
}

func NewFileObject(name string, size int64, r io.Reader) *FileObject {
	return &FileObject{
		name:   name,
		size:   size,
		reader: r,
	}
}

// FromLocalFile performs an os.Stat on the path and returns a WeTransfer file
// object.
func FromLocalFile(path string) (*FileObject, func() error, error) {
	name, size, err := fileInfo(path)
	if err != nil {
		return nil, nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	reader := bufio.NewReader(f)
	fo := NewFileObject(name, size, reader)

	return fo, f.Close, nil
}

// FromString returns a WeTransfer file object given a string and a filename
func FromString(content, filename string) (*FileObject, error) {
	reader := strings.NewReader(content)

	if reader.Size() == 0 {
		return nil, ErrBlankSize
	}

	if len(filename) == 0 {
		return nil, ErrBlankName
	}

	newName := sanitizeString(filename)
	fo := NewFileObject(newName, reader.Size(), reader)

	return fo, nil
}

func fileInfo(name string) (string, int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return "", 0, err
	}
	newName := sanitizeString(info.Name())
	return newName, info.Size(), nil
}

type fileObjectParam struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func toFileObjectParam(f *FileObject) *fileObjectParam {
	return &fileObjectParam{
		Name: f.Name(),
		Size: f.Size(),
	}
}
