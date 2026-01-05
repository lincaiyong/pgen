package goparser

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	code := `package main
func main() {
	print(12)
}`
	node, err := ParseBytes("main.go", []byte(code))
	if err != nil {
		t.Fatal(err)
	}
	dump := DumpNodeIndent(node)
	fmt.Println(dump)
}
