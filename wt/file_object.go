package wt

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"unicode/utf8"
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

	newName := sanitizeName(filename)
	fo := NewFileObject(newName, reader.Size(), reader)

	return fo, nil
}

func fileInfo(name string) (string, int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return "", 0, err
	}
	newName := sanitizeName(info.Name())
	return newName, info.Size(), nil
}

func sanitizeName(str string) string {
	origLen := utf8.RuneCountInString(str)
	newLen := origLen

	for _, r := range str {
		if isSanitizable(r) {
			newLen = newLen - utf8.RuneLen(r)
		}
	}

	if origLen == newLen {
		return str
	}

	newStr := make([]rune, 0, newLen)

	for _, r := range str {
		if !isSanitizable(r) {
			newStr = append(newStr, r)
		}
	}

	return string(newStr)
}

func isSanitizable(r rune) bool {
	return isSpecial(r) || isEmoji(r)
}

// Emojis rune range from WhatsApp stickers Swift repo
// https://github.com/WhatsApp/stickers/blob/master/iOS/WAStickersThirdParty/Sticker.swift#L42-L48
func isEmoji(r rune) bool {
	switch {
	case r >= 0x1F600 && r <= 0x1F64F, // Emoticons
		r >= 0x1F300 && r <= 0x1F5FF, // Misc Symbols and Pictographs
		r >= 0x1F680 && r <= 0x1F6FF, // Transport and maps
		r >= 0x2600 && r <= 0x26FF,   // Misc symbols
		r >= 0x2700 && r <= 0x27BF,   // Dingbats
		r >= 0x1F1E6 && r <= 0x1F1FF, // Flags
		r >= 0x1F900 && r <= 0x1F9FF: // Supplemental Symbols and Pictographs
		return true
	default:
		return false
	}
}

func isSpecial(c rune) bool {
	switch c {
	case '-', '_', '.', '~':
		return false
	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@':
		return true
	}
	return false
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
