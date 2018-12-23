package wt

import "errors"

type File struct {
	Multipart *struct {
		PartNumbers *int64 `json:"part_numbers"`
		ChunkSize   *int64 `json:"chunk_size"`
	} `json:"multipart,omitempty"`

	Size *int64  `json:"size"`
	Type *string `json:"type"`
	Name *string `json:"name"`
	ID   *string `json:"id"`
}

type Transfer struct {
	Success   *string `json:"success"`
	ID        *string `json:"id"`
	Message   *string `json:"message,omitempty"`
	State     *string `json:"state"`
	ExpiresAt *string `json:"expires_at"`
	URL       *string `json:"url,omitempty"`
	Files     []*File `json:"files"`
}

type TransferService service

func (t *TransferService) Create() (*Transfer, error) {
	return nil, errors.New("not implemented error")
}

func (t *TransferService) Find() (*Transfer, error) {
	return nil, errors.New("not implemented error")
}
