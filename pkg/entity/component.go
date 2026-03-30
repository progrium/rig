package entity

// should this be in node?
func ComponentEnabled(v any) bool {
	return Attr(v, "enabled") == "true"
}

type ComponentEntity interface {
	GetComponentType() string
}

func ComponentType(v any) string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(ComponentEntity); ok {
			return ee.GetComponentType()
		}
	}
	return ""
}
