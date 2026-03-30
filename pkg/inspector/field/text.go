package field

var TypeText = "Text"

type Text struct {
	String
}

func TextFrom(v string, fi FieldInfo) Text {
	return Text{StringFrom(v, fi)}
}

func (f Text) TypeName() string { return TypeText }
