
interface ViewData {
  ID: string;
  Name: string;
  Attrs: {};
  Annot: {};
  Value: any;
  Children: ViewData[];
}

export class View {
  ID: string;
  Name: string;
  Attrs: {};
  Annot: {};
  Value: any;
  Parent: View|undefined;
  Children: View[];
  Components: {};

  constructor(data: ViewData, annots: {}, parent: View = undefined) {
    this.ID = data.ID;
    this.Name = data.Name;
    this.Attrs = data.Attrs;
    this.Value = data.Value;
    this.Parent = parent;

    // merge annotations
    data.Annot = data.Annot || {};
    if (!annots[this.ID]) {
      annots[this.ID] = {};
    }
    for (const k in data.Annot) {
      annots[this.ID][k] = data.Annot[k];
    }
    this.Annot = annots[this.ID];

    // TODO: replace with proxy wrapper that accepts string props to lookup by name?
    this.Children = (data.Children || []).map((d) => new View(d, annots, this));

    this.Components = new Proxy(this.Value, {
      get: (target: any, prop: string, receiver: any) => {
        if (Number.isInteger(prop)) {
          return Reflect.get( target, prop, receiver );
        }
        const v = this.componentValue(prop);
        if (v) return v;
        const c = this.componentsByName(prop);
        if (c) return c.map((n) => n.Value);
      }
    })
  }

  // TODO: replace with values proxy property?
  componentValue(key: string): any|null {
    if (!Array.isArray(this.Value)) {
      return null;
    }
    for (const com of this.Value) {
      let v = com.Value[key]
      if (v) {
        return v;
      }
    }
    return null;
  }

  hasComponent(name: string): boolean {
    if (!Array.isArray(this.Value)) {
      return false;
    }
    for (const com of this.Value) {
      if (com.Name === name) {
        return true;
      }
    }
    return false;
  }

  childrenWithComponent(name: string): View[] {
    const children = [];
    for (const child of this.Children) {
      if (child.hasComponent(name)) {
        children.push(child);
      }
    }
    return children;
  }

  descendantsWithComponent(name: string): View[] {
    const descendants = [];
    this.walk((n) => {
      descendants.push(...n.childrenWithComponent(name));
      return true;
    })
    return descendants;
  }

  componentsByName(name: string): any[]|null {
    if (!Array.isArray(this.Value)) {
      return null;
    }
    return this.Value.filter((com) => com.Name === name);
  }

  childByName(name: string): View|null {
    for (const child of this.Children) {
      if (child.Name === name) {
        return child;
      }
    }
    return null;
  }

  select(...keys: string[]): View|null {
    let n: View = this;
    for (const k of keys) {
      let parts = k.split("/");
      for (const part of parts) {
        if (part === ".") {
          continue;
        }
        let child: View = n.childByName(part);
        if (child === null) {
          return null;
        }
        n = child;
      }
    }
    return n;
  } 

  walk(fn: (View) => boolean): boolean {
    if (!fn(this)) {
      return false;
    }
    for (const child of this.Children) {
      if (!child.walk(fn)) {
        return false;
      }
    }
    return true;
  }

  findID(id: string): View|null {
    let vv: View|null = null;
    this.walk((v: View) => {
      if (v.ID === id) {
        vv = v;
        return false;
      }
      return true;
    })
    return vv;
  }
}

// sketch of potential properties
//
// children - child nodes by name or index. "with" prop proxy for children with component? (obj.children.with["main.Attachment"])
// parents - ancestors by name or index. and "with"?
// components - components by name or index
// fields - aggregate of fields across components by name
// methods - aggregate of methods across components by name
// value - aggregate of component field values by name
