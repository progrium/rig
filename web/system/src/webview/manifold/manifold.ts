interface RawNode {
    ID: string;
    Kind: string;
    Bus: string;
  
    Name: string;
    Value: any;
  
    Component: string;
    Parent: string;
    Attrs: Record<string,string>;
    
    Children: string[];
    Components: string[];
  }
  
  export class Node {
    _id: string;
    _store: Realm;
  
    constructor(store: Realm, id: string) {
      this._store = store;
      this._id = id;
    }
  
    get id(): string {
      return this._id;
    }
  
    get store(): Realm {
      return this._store;
    }
  
    get raw(): RawNode {
      const raw = this._store.nodes[this.id];
      if (!raw) throw `use of non-existent node ${this.id}`;
      return raw;
    }
  
    get kind(): string {
      return this.raw.Kind;
    }

    get componentType(): string {
      return this.raw.Component;
    }
  
    get isComponent(): boolean {
      return this.kind === "com";
    }
  
    get name(): string {
      return this.raw.Name;
    }
  
    get value(): any {
      return this.raw.Value;
    }
  
    get parent(): Node|null {
      if (!this.raw.Parent) return null;
      if (!this._store.nodes[this.raw.Parent]) return null;
      return new Node(this._store, this.raw.Parent);
    }
  
    get children(): Node[] {
      let children: Node[] = [];
      if (this.raw.Children) {
        children = this.raw.Children.map(id => new Node(this._store, id));
      };
      return children;
    }
  
    get components(): Node[] {
      let coms: Node[] = [];
      if (this.raw.Components) {
        for (const id of this.raw.Components) {
          if (this._store.nodes[id]) {
            coms.push(new Node(this._store, id));
          }
        }
      };
      return coms;
    }
  
    // TODO: better api
    hasComponent(name: string): boolean {
      const coms = this.components.filter(n => n.name === name || n.componentType === name);
      if (coms.length > 0) {
        return true;
      }
      return false;
    }
  
    // TODO: better api
    componentValue(key: string): any|null {
      if (this.components.length === 0) {
        return null;
      }
      for (const com of this.components) {
        let v = com.value[key];
        if (v) {
          return v;
        }
      }
      return null;
    }
  
  
    // to port
  
  
    childrenWithComponent(name: string): Node[] {
        const children: Node[] = [];
        for (const child of this.children) {
            if (child.hasComponent(name)) {
                children.push(child);
            }
        }
        return children;
    }
  
    descendantsWithComponent(name: string): Node[] {
        const descendants: Node[] = [];
        this.walk((n) => {
            descendants.push(...n.childrenWithComponent(name));
            return true;
        })
        return descendants;
    }
  
    componentsByName(name: string): any[]|null {
        if (this.components.length === 0) {
            return null;
        }
        return this.components.filter((com) => com.name === name);
    }
  
    childByName(name: string): Node|null {
      for (const child of this.children) {
        if (child.name === name) {
          return child;
        }
      }
      return null;
    }
  
    select(...keys: string[]): Node|null {
      let n: Node = this;
      for (const k of keys) {
        let parts = k.split("/");
        for (const part of parts) {
          if (part === ".") {
            continue;
          }
          let child: Node|null = n.childByName(part);
          if (child === null) {
            return null;
          }
          n = child;
        }
      }
      return n;
    } 
  
    walk(fn: (node: Node) => boolean): boolean {
      if (!fn(this)) {
        return false;
      }
      for (const child of this.children) {
        if (!child.walk(fn)) {
          return false;
        }
      }
      return true;
    }
  
    findID(id: string): Node|null {
      return this.store.resolve(id)
    }
  }
  
  export class Realm extends EventTarget {
    nodes: Record<string,RawNode>;
    ready: Promise<void>;

    _firstUpdate: (() => void) | null;
  
    constructor() {
      super();
      this.nodes = {};
      this._firstUpdate = null;
      this.ready = new Promise((resolve) => {
        this._firstUpdate = resolve;
      });
    }
  
    resolve(id: string): Node|null {
      if (this.nodes[id]) {
        return new Node(this, id);
      }
      return null;
    }
  
    update(update: Record<string,RawNode>) {
      const added: string[] = [];
      const updated: string[] = [];
      const removed: string[] = [];
      for (const id in update) {
        if (update[id] === null) {
            removed.push(id);
            delete this.nodes[id]
            continue;
        }
        if (this.nodes[id]) {
          updated.push(id);
        } else {
          added.push(id);
        }
        this.nodes[id] = update[id];
      }
      for (const id of added) {
        this.dispatchEvent(new CustomEvent("add", {detail: new Node(this, id)}));
      }
      for (const id of updated) {
        this.dispatchEvent(new CustomEvent("update", {detail: new Node(this, id)}));
      }
      for (const id of removed) {
        this.dispatchEvent(new CustomEvent("remove", {detail: new Node(this, id)}));
      }
      this.dispatchEvent(new CustomEvent("change", {detail: [...added, ...updated, ...removed]}));
      if (this._firstUpdate) {
        this._firstUpdate();
        this._firstUpdate = null;
      }
    }
  }
  