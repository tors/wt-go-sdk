package wt

import (
	"context"
	"fmt"
	"net/url"
)

// Meta describes information on an item that is of link type
type Meta struct {
	Title *string `json:"title,omitempty"`
}

func (m Meta) String() string {
	return ToString(m)
}

// Item represents a board item. A board item can one of two
// types - link or file.
type Item struct {
	ID        *string    `json:"id"`
	Name      *string    `json:"name,omitempty"`
	URL       *string    `json:"url,omitempty"`
	Size      *int64     `json:"size,omitempty"`
	Type      *string    `json:"type"`
	Multipart *Multipart `json:"multipart,omitempty"`
	Meta      *Meta      `json:"meta,omitempty"`
}

func (i Item) String() string {
	return ToString(i)
}

// Board represents a board object. Each board can have 0 to many board items.
type Board struct {
	ID    *string `json:"id"`
	Name  *string `json:"name"`
	Desc  *string `json:"description"`
	State *string `json:"state"`
	URL   *string `json:"url"`
	Items []*Item `json:"items"`
}

// GetID returns the ID field if it is not nil. Otherwise, it returns
// an empty string.
func (b *Board) GetID() string {
	if b == nil || b.ID == nil {
		return ""
	}
	return *b.ID
}

func (b Board) String() string {
	return ToString(b)
}

// BoardsService handles communication with the board related methods of the
// WeTransfer API
type BoardsService service

// Create creates an empty WeTransfer board. Name is required but description
// is optional.
func (t *BoardsService) Create(ctx context.Context, name string, desc *string) (*Board, error) {
	req, err := t.client.NewRequest("POST", "boards", &struct {
		Name string  `json:"name"`
		Desc *string `json:"description"`
	}{
		Name: name,
		Desc: desc,
	})

	if err != nil {
		return nil, err
	}

	board := &Board{}
	if _, err := t.client.Do(ctx, req, board); err != nil {
		return nil, err
	}

	return board, nil
}

// Find retrieves a board given an id.
func (t *BoardsService) Find(ctx context.Context, id string) (*Board, error) {
	path := fmt.Sprintf("boards/%v", url.PathEscape(id))

	req, err := t.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	board := &Board{}
	if _, err = t.client.Do(ctx, req, board); err != nil {
		return nil, err
	}

	return board, nil
}
