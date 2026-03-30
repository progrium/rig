// @ts-ignore
import * as field from "./field.ts";

export interface Store {
  setValue(field, value)
  appendValue(field, value)
  removeKey(field, key)
  setKey(field, key, value)
}

export class Backend {
  setValue(field, value) {
    T.execute("backplane.SetValue", {Type: field.TypeName, Selector: field.ID, Value: value });
  }

  appendValue(field, value) {
    T.execute("backplane.AppendValue", {Type: field.TypeName, Selector: `${field.ID}`, Value: value});
  }
  
  removeKey(field, key) {
    T.execute("backplane.UnsetValue", {Type: field.TypeName, Selector: `${field.ID}/${key}`});
  }
  
  setKey(field, key, value) {
    T.execute("backplane.SetValue", {Type: field.TypeName, Selector: `${field.ID}/${key}`, Value: value });
  }
}

export class Buffer {
  fields: field.Field[]

  constructor(fields: field.Field[]) {
    this.fields = fields
  }

  setValue(field, value) {
    this.fields.filter(f => f.Name === field.Name)[0].Value = value
  }

  getValue() {
    return Object.fromEntries(this.fields.map(f => [f.Name, f.Value]))
  }

  appendValue(field, value) {
    throw "todo"
    let v = this.fields[field.ID||field.Name]
    if (!v) {
      v = []
      this.fields[field.ID||field.Name] = v
    }
    v.push(value)
  }
  
  removeKey(field, key) {
    throw "todo"
    delete this.fields[`${field.ID||field.Name}/${key}`]
  }
  
  setKey(field, key, value) {
    throw "todo"
    this.fields[`${field.ID||field.Name}/${key}`] = value
  }
}