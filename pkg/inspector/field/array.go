package field

var TypeArray = "Array"

type Array struct {
	Slice
}

func ArrayFrom(v interface{}, fi FieldInfo) Array {
	return Array{SliceFrom(v, fi)}
}

func (f Array) TypeName() string { return TypeArray }
