package snippet

const TypeNameOfFunc = `func TypeNameOf(node Node) string {
	structName := reflect.ValueOf(node).Elem().Type().Name()
	name := structName[:len(structName)-4]
	return toSnakeCase(name)
}`
