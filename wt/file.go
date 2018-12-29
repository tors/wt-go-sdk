package wt

import (
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

// BufferedFile implements the Transferable interface. It represents
// a file on disk to be sent as a file transfer.
type BufferedFile struct {
	name string
	size int64
	file *os.File
}

func (f *BufferedFile) GetName() string {
	return f.name
}

func (f *BufferedFile) GetSize() int64 {
	return f.size
}

func (f *BufferedFile) GetFile() *os.File {
	return f.file
}

func (f *BufferedFile) Close() error {
	return f.file.Close()
}

func NewBufferedFile(f string) (*BufferedFile, error) {
	name, size, err := fileInfo(f)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	return &BufferedFile{
		name: name,
		size: size,
		file: file,
	}, nil
}

// Buffer implements the Transferable interface. It represents a buffered data
// (usually created on the fly) to be sent as a file object
type Buffer struct {
	name string
	size int64
	b    []byte
}

func (f *Buffer) GetName() string {
	return f.name
}

func (f *Buffer) GetSize() int64 {
	return f.size
}

func (f *Buffer) GetBytes() []byte {
	return f.b
}

func NewBuffer(name string, b []byte) *Buffer {
	size := len(b)
	return &Buffer{
		name: name,
		size: int64(size),
		b:    b,
	}
}

// fileObject represents the parameter serialized in JSON format to be sent to
// create a file transfer
type fileObject struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// toFileObject converts a Transferable into a serializable file object
func toFileObject(t Transferable) fileObject {
	return fileObject{
		Name: t.GetName(),
		Size: t.GetSize(),
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
