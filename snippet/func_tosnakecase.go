package snippet

const ToSnakeCaseFunc = `func toSnakeCase(camelCaseString string) string {
	var sb strings.Builder
	for i, char := range camelCaseString {
		if uni.IsUpper(char) && i != 0 {
			sb.WriteRune('_')
		}
		sb.WriteRune(uni.ToLower(char))
	}
	return sb.String()
}`
