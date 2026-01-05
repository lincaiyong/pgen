package goparser

import (
	"fmt"
	"testing"
)

func TestHello(t *testing.T) {
	node, err := ParseBytes("test.go", []byte("package main\n"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(DumpNodeIndent(node))
}
