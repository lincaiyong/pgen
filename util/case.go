package util

import (
	"strings"
	"unicode"
)

func toPascalOrCamelCase(s string, isPascalCase bool) string {
	var sb strings.Builder
	shouldUpper := isPascalCase
	for i, r := range s {
		if i != 0 && r == '_' { // 下划线分隔符，将下一个字符转换为大写字母
			shouldUpper = true
		} else {
			if shouldUpper && unicode.IsLetter(r) { // 需要转换为大写字母
				sb.WriteRune(unicode.ToUpper(r))
				shouldUpper = false
			} else {
				sb.WriteRune(unicode.ToLower(r))
			}
		}
	}
	return sb.String()
}

func ToPascalCase(s string) string {
	return toPascalOrCamelCase(s, true)
}

func ToCamelCase(s string) string {
	return toPascalOrCamelCase(s, false)
}
