package wt

import "testing"

func TestNewFileObject(t *testing.T) {
	name := "filename.txt"
	size := int64(2)

	fo := NewFileObject(name, size)

	if fo.GetName() != name {
		t.Errorf("Name returned %v, want %v", fo.GetName(), name)
	}

	if fo.GetSize() != size {
		t.Errorf("Size returned %v, want %v", fo.GetSize(), size)
	}
}
