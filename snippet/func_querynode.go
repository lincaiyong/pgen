package snippet

const QueryNodeFunc = `func QueryNode(node Node, path string) (any, error) {
	if path == "" {
		return node, nil
	}

	items := strings.Split(path, "/")
	var base any
	base = node
	for _, item := range items {
		var name, nodeType string
		if strings.Contains(item, ":") {
			subs := strings.Split(item, ":")
			name = toCamelCase(subs[0])
			nodeType = subs[1]
		} else {
			name = toCamelCase(item)
		}

		switch base.(type) {
		case Node:
			node = base.(Node)
			if name == "." {
				base = node
			} else if name == ".." {
				base = node.Parent()
				if base == nil {
					return nil, errors.New("query error: node has no parent")
				}
			} else {
				t := reflect.TypeOf(node)
				m, ok := t.MethodByName(name)
				if !ok {
					methods := make([]string, 0)
					for i := 0; i < t.NumMethod(); i++ {
						tmp := t.Method(i).Name
						methods = append(methods, tmp)
					}
					return nil, errors.New(fmt.Sprintf("query error: %v has no method '%s', available: %s", t, name, strings.Join(methods, ", ")))
				}
				result := m.Func.Call([]reflect.Value{
					reflect.ValueOf(node),
				})
				base = result[0].Interface()
			}
		case []Node:
			nodes := base.([]Node)
			index, err := strconv.Atoi(name)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("query error: index should be an integer: '%s'", name))
			}
			if index < 0 || index >= len(nodes) {
				return nil, errors.New("index error")
			}
			base = nodes[index]
		default:
			return nil, errors.New(fmt.Sprintf("query error: neither Node nor []Node: '%s'", name))
		}

		// type assertion
		if nodeType != "" {
			if cast, isNode := base.(Node); isNode {
				t := TypeNameOf(cast)
				if strings.ToLower(t) != nodeType {
					return nil, errors.New(fmt.Sprintf("type assertion error, expect: %s, actual: %s", nodeType, t))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("type assertion error, not node"))
			}
		}
	}
	return base, nil
}`
