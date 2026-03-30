package telepath

type Selector interface {
	Select(path ...string) Cursor
}

func Select(v any, path ...string) Cursor {
	if s, ok := v.(Selector); ok {
		return s.Select(path...)
	}
	return New(v).Select(path...)
}
