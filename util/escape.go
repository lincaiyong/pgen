package util

import (
	"fmt"
	"strings"
)

func printableAscii(b rune) bool {
	switch b {
	case '!', '%', '&', '(', ')', '*', '+', ',', '.', '/', ':', ';', '<', '=', '>', '?', '@', '[', ']', '^', '{', '|', '}', '~', '#', '$', '-', '`', '\'', '"':
		return true
	}
	if (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') {
		return true
	}
	return false
}

func rune2hex4(r rune) [4]byte {
	var ret [4]byte
	for i := 0; i < 4; i++ {
		val := byte(r % 16)
		r = r / 16
		if val < 10 {
			ret[3-i] = '0' + val
		} else {
			ret[3-i] = 'A' + val - 10
		}
	}
	return ret
}

func escape(r rune) string {
	if printableAscii(r) {
		return string(r)
	}
	switch r {
	case '\t':
		return "\\t"
	case '\n':
		return "\\n"
	case '\r':
		return "\\r"
	case '\f':
		return "\\f"
	case '\\':
		return "\\\\"
	default:
		hex4 := rune2hex4(r)
		return fmt.Sprintf("\\u%s", string([]byte{hex4[0], hex4[1], hex4[2], hex4[3]}))
	}
}

func SingleQuoteStringEscape(s string) string {
	sb := strings.Builder{}
	for _, r := range []rune(s) {
		if r == '\'' {
			sb.WriteString("\\'")
		} else {
			sb.WriteString(escape(r))
		}
	}
	return sb.String()
}

func DoubleQuoteStringEscape(s string) string {
	sb := strings.Builder{}
	for _, r := range []rune(s) {
		if r == '"' {
			sb.WriteString("\\\"")
		} else {
			sb.WriteString(escape(r))
		}
	}
	return sb.String()
}
