package wt

import "testing"

func TestNewBufferedFile(t *testing.T) {
	name := "kitty.txt"
	content := "meow"
	size := len(content)

	file := setupTestFile(t, name, content)
	defer file.Close()

	fo, err := NewBufferedFile(file.Name())
	if err != nil {
		t.Errorf("NewBufferedFile returned an error: %v", err)
	}

	if fo.GetName() != name {
		t.Errorf("Name returned %v, want %v", fo.GetName(), name)
	}

	if fo.GetSize() != int64(size) {
		t.Errorf("Size returned %v, want %v", fo.GetSize(), size)
	}
}
