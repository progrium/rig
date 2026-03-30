package node

import "github.com/progrium/rig/pkg/entity"

const (
	Object    string = "obj"
	Component string = "com"
)

func IsComponent(n any) bool {
	return entity.ComponentType(n) != ""
}

func IsObject(n any) bool {
	return entity.ComponentType(n) == ""
}
