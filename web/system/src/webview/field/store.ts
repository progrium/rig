import type { Field } from "./field";

export interface Store {
  setValue(field: Field, value: any): void;
  appendValue(field: Field, value: any): void;
  removeKey(field: Field, key: string): void;
  setKey(field: Field, key: string, value: any): void;
}

export class Backend implements Store {
  setValue(_field: Field, _value: any): void {
    //T.execute("backplane.SetValue", {Type: field.TypeName, Selector: field.ID, Value: value });
  }

  appendValue(_field: Field, _value: any): void {
    //T.execute("backplane.AppendValue", {Type: field.TypeName, Selector: `${field.ID}`, Value: value});
  }

  removeKey(_field: Field, _key: string): void {
    //T.execute("backplane.UnsetValue", {Type: field.TypeName, Selector: `${field.ID}/${key}`});
  }

  setKey(_field: Field, _key: string, _value: any): void {
    //T.execute("backplane.SetValue", {Type: field.TypeName, Selector: `${field.ID}/${key}`, Value: value });
  }
}

export class Buffer implements Store {
  fields: Field[];

  constructor(fields: Field[]) {
    this.fields = fields;
  }

  setValue(field: Field, value: any): void {
    const f = this.fields.find((x) => x.Name === field.Name);
    if (f) {
      f.Value = value;
    }
  }

  getValue(): Record<string, any> {
    return Object.fromEntries(this.fields.map((f) => [f.Name, f.Value]));
  }

  appendValue(_field: Field, _value: any): void {
    throw new Error("todo");
  }

  removeKey(_field: Field, _key: string): void {
    throw new Error("todo");
  }

  setKey(_field: Field, _key: string, _value: any): void {
    throw new Error("todo");
  }
}
