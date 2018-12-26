package wt

import (
	"context"
	"errors"
)

type Meta struct {
	Title *string `json:"title,omitempty"`
}

type Item struct {
	ID        *string    `json:"id"`
	Name      *string    `json:"name,omitempty"`
	URL       *string    `json:"url,omitempty"`
	Size      *int64     `json:"size,omitempty"`
	Type      *string    `json:"type"`
	Multipart *Multipart `json:"multipart,omitempty"`
	Meta      *Meta      `json:"meta,omitempty"`
}

type Board struct {
	ID    *string `json:"id"`
	Name  *string `json:"name"`
	Desc  *string `json:"description"`
	State *string `json:"state"`
	URL   *string `json:"url"`
	Items []*Item `json:"items"`
}

type BoardService service

func (t *BoardService) Create(ctx context.Context, name *string, desc *string) (*Board, error) {
	if name == nil {
		return nil, ErrBlankName
	}

	param := &struct {
		Name *string `json:"name"`
		Desc *string `json:"description"`
	}{
		Name: name,
		Desc: desc,
	}

	req, err := t.client.NewRequest("POST", "boards", param)
	if err != nil {
		return nil, err
	}

	board := &Board{}
	if _, err := t.client.Do(ctx, req, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (t *BoardService) Find() error {
	return errors.New("not implemented error")
}
