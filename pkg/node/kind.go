package node

const (
	Object    string = "obj"
	Component string = "com"
)

func IsComponent(n any) bool {
	return ComponentType(n) != ""
}

func IsObject(n any) bool {
	return ComponentType(n) == ""
}
