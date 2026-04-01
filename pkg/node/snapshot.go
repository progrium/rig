package node

import "reflect"

// Snapshot returns a Raw-shaped copy of n suitable for serialization or export.
// When n is a *Raw, it returns a shallow copy of the struct with Value duplicated via dupVal
// (not a deep copy). Otherwise it synthesizes a Raw from Node APIs: ids of object and
// component subnodes, AttrMap, parent id, name, kind-related fields, and dupVal(Value(n)).
func Snapshot(n Node) Raw {
	if r, ok := Unwrap[*Raw](n); ok {
		v := *r
		v.Value = dupVal(v.Value)
		return v
	}
	var objIDs []string
	var comIDs []string
	for _, child := range Subnodes(n, TypeObject) {
		objIDs = append(objIDs, child.NodeID())
	}
	for _, child := range Subnodes(n, TypeComponent) {
		comIDs = append(comIDs, child.NodeID())
	}
	return Raw{
		ID:         n.NodeID(),
		Name:       Name(n),
		Value:      dupVal(Value(n)),
		Attrs:      AttrMap(n),
		Parent:     ParentID(n),
		Component:  ComponentType(n),
		Children:   objIDs,
		Components: comIDs,
	}
}

// dupVal returns a duplicate of v for pointers and interfaces by allocating a new pointer
// and copying the element; other kinds are returned as-is. It is not a deep copy.
func dupVal(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		el := rv.Elem()
		nv := reflect.New(el.Type())
		nv.Elem().Set(el)
		return nv.Interface()
	}
	return rv.Interface()
}
