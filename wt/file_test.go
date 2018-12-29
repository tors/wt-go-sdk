package wt

import (
	"fmt"
	"testing"
)

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

func ExampleNewBufferedFile() {
	buf, err := NewBufferedFile("../example/files/Japan-02.jpg")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(buf.GetName())
	fmt.Println(buf.GetSize())

	// Output:
	// Japan-02.jpg
	// 275639
}

func ExampleNewBuffer() {
	buf := NewBuffer("pony.txt", []byte("yehaaa"))

	fmt.Println(buf.GetName())
	fmt.Println(buf.GetSize())

	// Output:
	// pony.txt
	// 6
}
