package wt

import (
	"fmt"
	"testing"
)

func TestBuildBufferedFile(t *testing.T) {
	name := "kitty.txt"
	content := "meow"
	size := len(content)

	file := setupTestFile(t, name, content)
	defer file.Close()

	tests := []interface{}{
		file.Name(),
		file,
	}

	for _, f := range tests {
		fo, err := BuildBufferedFile(f)
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
}

func ExampleBuildBufferedFile() {
	buf, err := BuildBufferedFile("../example/files/Japan-02.jpg")
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
