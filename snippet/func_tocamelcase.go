package snippet

const ToCamelCaseFunc = `func toCamelCase(s string) string {
	var sb strings.Builder
	shouldUpper := true
	for _, r := range s {
		if r == '_' { // 下划线分隔符，将下一个字符转换为大写字母
			shouldUpper = true
		} else {
			if shouldUpper && uni.IsLetter(r) { // 需要转换为大写字母
				sb.WriteRune(uni.ToUpper(r))
				shouldUpper = false
			} else {
				sb.WriteRune(uni.ToLower(r))
			}
		}
	}
	return sb.String()
}`
