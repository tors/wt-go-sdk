package wt

import (
	"context"
	"fmt"
	"net/url"
	"os"
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

// GetName returns the Name field if it is not nil. Otherwise, it returns
// an empty string.
func (i *Item) GetName() string {
	if i == nil || i.Name == nil {
		return ""
	}
	return *i.Name
}

// GetID returns the ID field if it is not nil. Otherwise, it returns
// an empty string.
func (i *Item) GetID() string {
	if i == nil || i.ID == nil {
		return ""
	}
	return *i.ID
}

// GetMultipart returns the Multipart field.
func (i *Item) GetMultipart() *Multipart {
	if i == nil || i.Multipart == nil {
		return nil
	}
	return i.Multipart
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

// AddLinks creates link items for a given board. It returns a list of items
// with meta information.
func (b *BoardsService) AddLinks(ctx context.Context, bid string, links ...*Link) ([]*Item, error) {
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

// AddFiles uploads files to a specified board.
func (b *BoardsService) AddFiles(ctx context.Context, board *Board, in ...interface{}) ([]*Item, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("empty files")
	}

	files := make([]transferable, len(in))

	for i, obj := range in {
		switch v := obj.(type) {
		case string, *os.File:
			buf, err := BuildBufferedFile(v)
			if err != nil {
				return nil, err
			}
			files[i] = buf
		case *Buffer:
			files[i] = (*Buffer)(v)
		case *BufferedFile:
			files[i] = (*BufferedFile)(v)
		default:
			return nil, fmt.Errorf(`allowed types are string *Buffer *BufferedFile`)
		}
	}

	filemap := make(map[string]transferable)
	for _, f := range files {
		name := f.GetName()
		filemap[name] = f
	}

	items, err := b.uploadFiles(ctx, board, files...)
	if err != nil {
		return nil, err
	}

	var errs []error

	for _, f := range items {
		name := f.GetName()
		if tx, ok := filemap[name]; ok {
			ft := newFileTransfer(tx, f)
			err = b.client.uploader.upload(ctx, board, ft)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return nil, joinErrors(errs, nil)
	}

	// If we have reached this stage, that means there were no errors while
	// uploading the files/chunks. Now we attempt to complete it.
	err = b.complete(ctx, board, items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (b *BoardsService) uploadFiles(ctx context.Context, board *Board, tx ...transferable) ([]*Item, error) {
	var fs []fileObject

	for _, obj := range tx {
		fs = append(fs, toFileObject(obj))
	}

	bid := board.GetID()
	path := fmt.Sprintf("boards/%v/files", url.PathEscape(bid))
	req, err := b.client.NewRequest("POST", path, &struct {
		Files []fileObject `json:"files"`
	}{
		Files: fs,
	})

	if err != nil {
		return nil, err
	}

	var items []*Item
	if _, err = b.client.Do(ctx, req, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (b *BoardsService) complete(ctx context.Context, board *Board, items []*Item) error {
	var errs []error

	bid := url.PathEscape(board.GetID())
	for _, item := range items {
		fid := url.PathEscape(item.GetID())
		path := fmt.Sprintf("boards/%v/files/%v/upload-complete", bid, fid)
		req, err := b.client.NewRequest("PUT", path, nil)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, err = b.client.Do(ctx, req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		errmsg := fmt.Sprintf("completing transfer %v, with %v error(s)", bid, len(errs))
		return joinErrors(errs, &errmsg)
	}

	return nil
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
