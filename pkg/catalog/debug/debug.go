package debug

import "log"

type Debug struct {
	String    string
	Integer   int
	Float     float64
	Bool      bool
	Struct    SubDebug
	PtrString *string
	PtrStruct *Debug
}

func (*Debug) DebugMethod() {
	log.Println("DebugMethod called!")
}

type SubDebug struct {
	Sub       string
	SubStruct SubSubDebug
	PtrStruct *Debug
}

func (SubDebug) SubMethod(int) bool { return false }

type SubSubDebug struct {
	SubSub string
	Leaf   LeafDebug
}

type LeafDebug struct {
	Leaf string
}
