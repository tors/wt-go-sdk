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

// Link represents a link item in WeTransfer.
type Link struct {
	URL   *string `json:"url"`
	Title *string `json:"title"`
}

func (l Link) String() string {
	return ToString(l)
}

// NewLink returns a new link given a URL and an optional title. It
// automatically validates the URL string and returns an error if it fails
// validation.
func NewLink(u string, title *string) (*Link, error) {
	_, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	link := &Link{
		URL:   &u,
		Title: title,
	}
	return link, nil
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
func (b *BoardsService) Create(ctx context.Context, name string, desc *string) (*Board, error) {
	req, err := b.client.NewRequest("POST", "boards", &struct {
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
	if _, err := b.client.Do(ctx, req, board); err != nil {
		return nil, err
	}

	return board, nil
}

// AddLink creates a link item for a given board. It returns an Item with meta
// information when the request is successful.
func (b *BoardsService) AddLink(ctx context.Context, bid string, links ...*Link) ([]*Item, error) {
	path := fmt.Sprintf("boards/%v/links", url.PathEscape(bid))

	var gotLinks []*Link
	for _, link := range links {
		if link != nil {
			gotLinks = append(gotLinks, link)
		}
	}
	if len(gotLinks) == 0 {
		return nil, fmt.Errorf("no links provided")
	}

	req, err := b.client.NewRequest("POST", path, gotLinks)
	if err != nil {
		return nil, err
	}

	var items []*Item
	if _, err := b.client.Do(ctx, req, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// Find retrieves a board given an id.
func (b *BoardsService) Find(ctx context.Context, id string) (*Board, error) {
	path := fmt.Sprintf("boards/%v", url.PathEscape(id))

	req, err := b.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	board := &Board{}
	if _, err = b.client.Do(ctx, req, board); err != nil {
		return nil, err
	}

	return board, nil
}
