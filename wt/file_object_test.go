package wt

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewFileObject(t *testing.T) {
	name := "abc.txt"
	content := "content"

	str := strings.NewReader(content)
	fo := NewFileObject(name, 1, str)

	if fo.Name() != name {
		t.Errorf("Name returned %v, want %v", fo.Name(), name)
	}

	if fo.Size() != 1 {
		t.Errorf("Size returned %v, want %v", fo.Size(), 1)
	}

	b, _ := ioutil.ReadAll(str)

	if string(b) != content {
		t.Errorf("Reader content returned %v, want %v", string(b), content)
	}
}

func TestFromString_withEmojis(t *testing.T) {

	tests := []struct {
		filename, want string
	}{
		{"filename-🙈.jpg", "filename-.jpg"},
		{"🙊-filename-🙈.jpg", "-filename-.jpg"},
		{"🙊-file-🙉-name-🙈.jpg", "-file--name-.jpg"},
	}

	for _, c := range tests {
		object, _ := FromString("content", c.filename)
		if object.Name() != c.want {
			t.Errorf("Name returned %v, want %v", object.Name(), c.want)
		}
	}
}

func TestFromString_empty(t *testing.T) {
	_, err := FromString("", "")

	if err != ErrBlankSize {
		t.Errorf("FromString returned error %v", err)
	}

	_, err = FromString("content", "")
	if err != ErrBlankName {
		t.Errorf("FromString returned error %v", err)
	}
}

func TestFromLocalFile(t *testing.T) {

	tests := []struct {
		file, name string
		size       int64
	}{
		{"../example/files/Japan-01🇯🇵.jpg", "Japan-01.jpg", 13370099},
		{"../example/files/Japan-02.jpg", "Japan-02.jpg", 275639},
		{"../example/files/Japan-03.jpg", "Japan-03.jpg", 432557},
	}

	for _, c := range tests {
		object, closer, _ := FromLocalFile(c.file)
		defer closer()

		if object.Name() != c.name {
			t.Errorf("Name returned %v, want %v", object.Name(), c.name)
		}

		if object.Size() != c.size {
			t.Errorf("Size returned %v, want %v", object.Size(), c.size)
		}
	}
}