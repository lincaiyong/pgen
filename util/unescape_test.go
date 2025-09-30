package util

import (
	"testing"
)

func TestUnescape(t *testing.T) {
	source := string([]byte{'\\', '\\', '\\', 't', '\\', 'n', '\\', 'r', '\\', '\'', '"', '\\', 'u', '0', '0', '3', '0', '\\', 'u', '1'})
	target := string([]byte{'\\', '\t', '\n', '\r', '\'', '"', '0', 'u', '1'})
	s := SingleQuoteStringUnescape(source)
	if s != target {
		t.Fatal(s)
	}
}
