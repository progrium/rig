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
    _realm: Realm;
  
    constructor(realm: Realm, id: string) {
      this._realm = realm;
      this._id = id;
    }
  
    get id(): string {
      return this._id;
    }
  
    get realm(): Realm {
      return this._realm;
    }
  
    get raw(): RawNode {
      const raw = this._realm.nodes[this.id];
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
      if (!this._realm.nodes[this.raw.Parent]) return null;
      return new Node(this._realm, this.raw.Parent);
    }
  
    get children(): Node[] {
      let children: Node[] = [];
      if (this.raw.Children) {
        children = this.raw.Children.map(id => new Node(this._realm, id));
      };
      return children;
    }

    get childCount(): number {
      return this.raw.Children?.length ?? 0;
    }

    get siblingIndex(): number {
      if (!this.raw.Parent) return -1;
      if (!this._realm.nodes[this.raw.Parent]) return -1;
      const parent = new Node(this._realm, this.raw.Parent);
      return parent.children.indexOf(this);
    }


    get nextSibling(): Node|null {
      if (!this.raw.Parent) return null;
      if (!this._realm.nodes[this.raw.Parent]) return null;
      const parent = new Node(this._realm, this.raw.Parent);
      const index = parent.children.indexOf(this);
      if (index === -1 || index === parent.children.length - 1) return null;
      return parent.children[index + 1];
    }

    get prevSibling(): Node|null {
      if (!this.raw.Parent) return null;
      if (!this._realm.nodes[this.raw.Parent]) return null;
      const parent = new Node(this._realm, this.raw.Parent);
      const index = parent.children.indexOf(this);
      if (index === -1 || index === 0) return null;
      return parent.children[index - 1];
    }
  
    get components(): Node[] {
      let coms: Node[] = [];
      if (this.raw.Components) {
        for (const id of this.raw.Components) {
          if (this._realm.nodes[id]) {
            coms.push(new Node(this._realm, id));
          }
        }
      };
      return coms;
    }

    get componentCount(): number {
      return this.raw.Components?.length ?? 0;
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
      return this.realm.resolve(id)
    }

    async destroy(): Promise<void> {
      await this.realm.peer.call("callMethod", {Selector: `${this.id}/Destroy`});
      // delete this.realm.nodes[this.id];
      // this.realm.dispatchEvent(new CustomEvent("remove", {detail: this}));
    }

    async setName(name: string): Promise<void> {
      await this.realm.peer.call("callMethod", {Selector: `${this.id}/SetName`, Args: [name]});
    }

    set name(name: string) {
      this.setName(name);
      this.raw.Name = name;
    }

    async setSiblingIndex(index: number): Promise<void> {
      await this.realm.peer.call("callMethod", {Selector: `${this.id}/SetSiblingIndex`, Args: [index]});
    }

    set siblingIndex(index: number) {
      if (!this.raw.Parent) return;
      if (!this._realm.nodes[this.raw.Parent]) return;
      this.setSiblingIndex(index);

      // now we do it locally. a lot of work just to be made eventually consistent.
      const parentRaw = this._realm.nodes[this.raw.Parent];
      let children: string[] = [];
      if (this.raw.Kind === "obj") {
        children = parentRaw.Children;
      } else {
        children = parentRaw.Components;
      }
      if (!children) return;

      const currentIdx = children.indexOf(this.id);
      if (currentIdx === -1 || index === currentIdx) return;

      // Remove at current position
      children.splice(currentIdx, 1);

      // Bound the new index between 0 and children.length
      let newIndex = index;
      if (newIndex < 0) newIndex = 0;
      if (newIndex > children.length) newIndex = children.length;

      // Insert this.id at the new index
      children.splice(newIndex, 0, this.id);
    }

    async setParent(parent: Node): Promise<void> {
      await this.realm.peer.call("callMethod", {Selector: `${this.id}/SetParentID`, Args: [parent.id]});
      await this.realm.peer.call("callMethod", {Selector: `${parent.id}/Node/AppendSubnode`, Args: ["obj", this.id]});
      // direct node ops need to manually be signalled
      await this.realm.peer.call("callMethod", {Selector: `${parent.id}/Send`, Args: ["AppendSubnode", ["obj", this.id]]});
    }

    set parent(parent: Node) {
      this.setParent(parent);
      this.raw.Parent = parent.id;
      // we'll let the subnode backrefs update eventually.
    }
  }
  
  export class Realm extends EventTarget {
    nodes: Record<string,RawNode>;
    ready: Promise<void>;
    peer: any;

    _firstUpdate: (() => void) | null;
  
    constructor(peer: any) {
      super();
      this.nodes = {};
      this._firstUpdate = null;
      this.ready = new Promise((resolve) => {
        this._firstUpdate = resolve;
      });
      this.peer = peer;
    }

    async create(name: string, parent: Node|null): Promise<Node> {
      const resp = await this.peer.call("addObject", {Value: name, Selector: parent?.id ?? "" });
      const id = resp.value;
      // temp node for now, well be eventually consistent
      this.nodes[id] = {
        ID: id,
        Kind: "obj",
        Bus: "",
        Name: name,
        Value: null,
        Component: "obj",
        Parent: parent?.id ?? "",
        Attrs: {},
        Children: [], 
        Components: []
      }
      return new Node(this, id);
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
          if (this.nodes[id]) {
            removed.push(id);
            delete this.nodes[id];
          }
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
  