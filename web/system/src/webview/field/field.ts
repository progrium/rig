export interface Field {
  TypeName: string;
  Name: string;
  Flags: string[];
  ID?: string;
  Value?: any;
  Enum?: string[];
  Min?: number;
  Annots?: {
    Obj?: { Name?: string };
    pkgpath?: string;
    [key: string]: any;
  };
  ElemType?: string;
  /** Struct subfields, or Array/Slice element field descriptors */
  Fields?: Field[];
  ElemFields?: Field[];
}
