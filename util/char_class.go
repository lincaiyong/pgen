package util

import "fmt"

func ParseCharacterClass(s string) ([][]rune, error) {
	ret := make([][]rune, 0)
	var last rune = -1
	var rangeStart rune = -1
	for i := 0; i < len(s); i++ {
		// 连接符，挂起
		if s[i] == '-' && i != len(s)-1 {
			if last == -1 {
				return nil, fmt.Errorf("parse character class: symbol - is misused, [%s]", s)
			}
			rangeStart = last
			last = -1
		} else {
			// 收回普通 item
			if last != -1 {
				ret = append(ret, []rune{last})
			}
			// 解析 item, 但挂起
			if s[i] == '\\' {
				last, i = unescape(s, i)
			} else {
				last = rune(s[i])
			}
			// 如果是连接符之后的 item, 收回
			if rangeStart != -1 {
				ret = append(ret, []rune{rangeStart, last})
				rangeStart = -1
				last = -1
			}
		}
	}
	if last != -1 {
		ret = append(ret, []rune{last})
	}
	return ret, nil
}
