package wt

import (
	"fmt"
	"testing"
)

func Test_sanitizeString(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{"filename-🙈.jpg", "filename-.jpg"},
		{"🙊-filename-🙈.jpg", "-filename-.jpg"},
		{"🙊-file-🙉-name-🙈.jpg", "-file--name-.jpg"},
		{"file-$&+,/:;=?@-name&.jpg", "file--name.jpg"},
		{"file-_.~-name.jpg", "file-_.~-name.jpg"},
	}

	for _, c := range tests {
		str := sanitizeString(c.got)
		if str != c.want {
			t.Errorf("sanitizeString returned %v, want %v", str, c.want)
		}
	}
}

func TestToString_structs(t *testing.T) {
	var tests = []struct {
		got  interface{}
		want string
	}{
		{Board{ID: String("id"), Name: String("board1"), Items: []*Item{}}, `wt.Board{ID:"id", Name:"board1", Items:[]}`},
		{Multipart{PartNumbers: Int64(1), ChunkSize: Int64(2)}, `wt.Multipart{PartNumbers:1, ChunkSize:2}`},
		{File{Size: Int64(2), Type: String("a"), ID: String("c")}, `wt.File{Size:2, Type:"a", ID:"c"}`},
	}

	for i, tt := range tests {
		s := tt.got.(fmt.Stringer).String()
		if s != tt.want {
			t.Errorf("%d. String() => %q, want %q", i, tt.got, tt.want)
		}
	}
}

func TestToString(t *testing.T) {
	var nilPointer *string

	var tests = []struct {
		got  interface{}
		want string
	}{
		{"foo", `"foo"`},
		{123, `123`},
		{1.5, `1.5`},
		{false, `false`},
		{[]string{"a", "b"}, `["a" "b"]`},
		{struct{ A []string }{nil}, `{}`},
		{struct{ A string }{"foo"}, `{A:"foo"}`},
		{nilPointer, `<nil>`},
		{String("foo"), `"foo"`},
		{Int(123), `123`},
		{Int64(123), `123`},
		{Bool(false), `false`},
		{[]*string{String("a"), String("b")}, `["a" "b"]`},
	}

	for i, tt := range tests {
		s := ToString(tt.got)
		if s != tt.want {
			t.Errorf("%d. ToString(%q) => %q, want %q", i, tt.got, s, tt.want)
		}
	}
}
