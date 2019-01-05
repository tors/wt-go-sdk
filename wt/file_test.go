package wt

import (
	"fmt"
)

func ExampleLocalFile() {
	japan, _ := NewLocalFile("../example/files/Japan-01🇯🇵.jpg")
	name, size := japan.Stat()

	fmt.Println(name)
	fmt.Println(size)

	// Output:
	// Japan-01.jpg
	// 13370099
}

func ExampleNewBuffer() {
	buf := NewBuffer("pony.txt", []byte("yehaaa"))
	name, size := buf.Stat()

	fmt.Println(name)
	fmt.Println(size)

	// Output:
	// pony.txt
	// 6
}
