// @ts-ignore
import * as buffer from "/-/inspector/input/buffer.ts";

interface Type {
  TypeName: string;
}

interface ValueType extends Type {
  Default: string;
  Enum?: string[];
  Range?: Range;
}

interface Range {
	Min: number;
  Step: number;
	Max: number;
}

interface StructType extends Type {
  Fields: Field[];
}

interface CollectionType extends Type {
  IdxType: string;
	ElemType: string;
}

interface Field extends Type {
  Name: string;
  Flags: string[];
  ID?: string;
  Value?: any;
}


// type BasicType interface {
// 	Parse(s string) (Value, error)
// 	Format() string
// }
