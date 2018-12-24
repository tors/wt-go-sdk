package wt

// Uploadable describes object (usually a file) that can be sent
// as parameters to the API
type Uploadable interface {
	Size() int64
	Name() string
}

// UploadableFile implements Uploadable
type UploadableFile struct {
	size int64
	name string
	Path string
}

func (u *UploadableFile) Size() int64 {
	return u.size
}

func (u *UploadableFile) Name() string {
	return u.name
}

func NewUploadableFile(path string) (*UploadableFile, error) {
	name, size, err := fileInfo(path)

	if err != nil {
		return nil, err
	}

	return &UploadableFile{
		Path: path,
		name: name,
		size: size,
	}, nil
}

type UploadableBytes struct {
	size int64
	name string
	b    []byte
}

func (u *UploadableBytes) Size() int64 {
	return u.size
}

func (u *UploadableBytes) Name() string {
	return u.name
}

func NewUploadableBytes(name string, b []byte) (*UploadableBytes, error) {
	size := len(b)

	if size == 0 {
		return nil, ErrBlankContent
	}

	return &UploadableBytes{
		b:    b,
		name: name,
		size: int64(size),
	}, nil
}

// uploadableParam represents the JSON payload when an uploadable is
// sent to the server
type uploadableParam struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func toUploadableParam(t Uploadable) *uploadableParam {
	return &uploadableParam{
		Name: t.Name(),
		Size: t.Size(),
	}
}
