package util

import "testing"

func TestEscape(t *testing.T) {
	source := string([]byte{'\\', '\t', '\f', '\n', '\r', '"', '\'', '1', '2', '3'})
	target := string([]byte{'\\', '\\', '\\', 't', '\\', 'f', '\\', 'n', '\\', 'r', '"', '\\', '\'', '1', '2', '3'})
	s := SingleQuoteStringEscape(source)
	if s != target {
		t.Fatal(s)
	}
}
