package wt

import (
	"fmt"
	"os"
)

// File represents a WeTransfer file object transfer response
type File struct {
	Multipart *Multipart `json:"multipart"`
	Size      *int64     `json:"size"`
	Type      *string    `json:"type"`
	Name      *string    `json:"name"`
	ID        *string    `json:"id"`
}

func (f *File) GetMultipartValues() (partNum int64, chunkSize int64) {
	if f == nil || f.Multipart == nil {
		return int64(0), int64(0)
	}
	return f.Multipart.GetPartNumbers(), f.Multipart.GetChunkSize()
}

func (f *File) GetName() string {
	if f == nil || f.Name == nil {
		return ""
	}
	return *f.Name
}

func (f *File) GetID() string {
	if f == nil || f.ID == nil {
		return ""
	}
	return *f.ID
}

func (r File) String() string {
	return ToString(r)
}

// Multipart is info on the chunks of data to be uploaded
type Multipart struct {
	PartNumbers *int64 `json:"part_numbers"`
	ChunkSize   *int64 `json:"chunk_size"`
}

func (m *Multipart) GetPartNumbers() int64 {
	if m == nil || m.PartNumbers == nil {
		return int64(0)
	}
	return *m.PartNumbers
}

func (m *Multipart) GetChunkSize() int64 {
	if m == nil || m.ChunkSize == nil {
		return int64(0)
	}
	return *m.ChunkSize
}

func (m Multipart) String() string {
	return ToString(m)
}

type FileError struct {
	file *File
	err  error
}

func (f *FileError) Error() string {
	return fmt.Sprintf(`%v, %v`, f.file.GetID(), f.err.Error())
}

func NewFileError(file *File, err error) *FileError {
	return &FileError{
		file: file,
		err:  err,
	}
}

// FileObject represents a file object in WeTransfer
type FileObject struct {
	Name *string `json:"name"`
	Size *int64  `json:"size"`
}

func (f *FileObject) GetName() string {
	if f == nil || f.Name == nil {
		return ""
	}
	return *f.Name
}

func (f *FileObject) GetSize() int64 {
	if f == nil || f.Size == nil {
		return int64(0)
	}
	return *f.Size
}

func NewFileObject(name string, size int64) *FileObject {
	return &FileObject{
		Name: &name,
		Size: &size,
	}
}

func fileInfo(name string) (string, int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return "", 0, err
	}
	newName := sanitizeString(info.Name())
	return newName, info.Size(), nil
}
