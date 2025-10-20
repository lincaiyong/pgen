package snippet

const CreationHookVar = `var creationHook = func(Node) {}

func SetCreationHook(h func(Node)) {
	creationHook = h
}`
