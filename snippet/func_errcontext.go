package snippet

const ErrorContextFunc = `func errorContext(filePath string, fileContent []rune, offset, lineIdx, charIdx int) string {
	var lineStartOffset int
	for i := offset; i >= 0; i-- {
		if i < len(fileContent) && fileContent[i] == '\n' {
			lineStartOffset = i + 1
			break
		}
	}
	lineText := regexp.MustCompile("[^\\t]").ReplaceAllString(string(fileContent[lineStartOffset:offset]), " ")

	lines := strings.Split(string(fileContent), "\n")
	contextLines := 3
	startLine := lineIdx - contextLines
	if startLine < 0 {
		startLine = 0
	}
	endLine := lineIdx + contextLines
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== error context (%s:%d:%d) ===\n", filePath, lineIdx+1, charIdx+1))
	for i := startLine; i <= endLine; i++ {
		prefix := "   "
		var t string
		if i == lineIdx {
			prefix = ">>>"
			t = fmt.Sprintf("          %s^\n", lineText)
		}
		sb.WriteString(fmt.Sprintf("%s %4d: %s\n", prefix, i+1, lines[i]))
		if t != "" {
			sb.WriteString(t)
		}
	}
	sb.WriteString("=== end of error context ===")
	return sb.String()
}`
