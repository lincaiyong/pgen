package snippet

const DumpNodeFunc = `func DumpNode(n Node, hook func(Node, map[string]string) string) string {
	return CustomDumpNode(n, hook)
}

func DumpNodeIndent(node Node) string {
	result := SimpleDumpNode(node)
	var v any
	err := json.Unmarshal([]byte(result), &v)
	if err != nil {
		panic(err)
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func CustomDumpNode(node Node, hook func(Node, map[string]string) string) string {
	if node.IsDummy() {
		return "null"
	}
	itemMap := node.Dump(hook)
	ret := hook(node, itemMap)
	if ret != "" {
		return ret
	}
	items := make([]string, 0)
	for k, v := range itemMap {
		if k == "kind" {
			continue
		}
		items = append(items, fmt.Sprintf("\"%s\": %s", k, v))
	}
	sort.Strings(items)
	items = append([]string{fmt.Sprintf("\"kind\": %s", itemMap["kind"])}, items...)
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}

func SimpleDumpNode(node Node) string {
	return CustomDumpNode(node, func(n Node, m map[string]string) string {
		return ""
	})
}`
