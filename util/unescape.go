package util

func hex4toRune(bs [4]byte) rune {
	ret := 0
	for i := 0; i <= 3; i++ {
		if bs[i] >= '0' && bs[i] <= '9' {
			ret = ret*16 + int(bs[i]-'0')
		} else if bs[i] >= 'A' && bs[i] <= 'F' {
			ret = ret*16 + int(bs[i]-'A'+10)
		} else {
			ret = ret*16 + int(bs[i]-'a'+10)
		}
	}
	return rune(ret)
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

func isUnicodeFormat(s string, i int) bool {
	return len(s) > i+5 && s[i] == '\\' && s[i+1] == 'u' && isHex(s[i+2]) && isHex(s[i+3]) && isHex(s[i+4]) && isHex(s[i+5])
}

func unescape(s string, i int) (rune, int) {
	// 'i' points to '\'
	// result int points to last byte that used
	if isUnicodeFormat(s, i) {
		return hex4toRune([4]byte{s[i+2], s[i+3], s[i+4], s[i+5]}), i + 5
	}
	i++ // now points to x
	switch s[i] {
	case 't':
		return '\t', i
	case 'n':
		return '\n', i
	case 'r':
		return '\r', i
	case 'f':
		return '\f', i
	default:
		return rune(s[i]), i
	}
}

func SingleQuoteStringUnescape(s string) string {
	ret := make([]rune, 0)
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			var r rune
			if s[i+1] == '\'' {
				r = '\''
				i = i + 1
			} else {
				r, i = unescape(s, i)
			}
			ret = append(ret, r)
		} else {
			ret = append(ret, rune(s[i]))
		}
	}
	return string(ret)
}

func DoubleQuoteStringUnescape(s string) string {
	ret := make([]rune, 0)
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			var r rune
			if s[i+1] == '"' {
				r = '"'
				i = i + 1
			} else {
				r, i = unescape(s, i)
			}
			ret = append(ret, r)
		} else {
			ret = append(ret, rune(s[i]))
		}
	}
	return string(ret)
}
