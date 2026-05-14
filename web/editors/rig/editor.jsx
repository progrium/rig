
export const Header = null;

export const Editor = {
    view: ({attrs, state}) => {
        const {vscode,peer,realm,nodeID} = attrs;
        const node = realm.resolve(nodeID);
        if (node === null) {
            return <div>No node</div>;
        }
        if (!state.workbench) {
          state.workbench = new Workbench(new Path(node), peer, realm);
        }
        return <div class="flex flex-col">
          hello
            <div class="title-node" oncontextmenu={(e) => state.workbench.showMenu(e, { node, path })} data-menu="node">
                <NodeEditor workbench={state.workbench} path={state.workbench.context.path} disallowEmpty={true} />
            </div>
            <OutlineEditor path={state.workbench.context.path.sub()} workbench={state.workbench} />
        </div>;
    }
}

const commands = {
  "expand": (wb, ctx) => {
    if (!ctx.node) return;
    wb.setExpanded(ctx.path.head, ctx.node, true);
    m.redraw();
  },
  "collapse": (wb, ctx) => {
    if (!ctx.node) return;
    wb.setExpanded(ctx.path.head, ctx.node, false);
    m.redraw();
  },
  "zoom": (wb, ctx) => {
    wb.context.path = ctx.path.append(ctx.node);
    wb.root = workbench.context.path;
    m.redraw();
  },
  "insert": async (wb, ctx, name = "") => {
    if (!ctx.node) return;
    // if (ctx.path.previous && objectManaged(ctx.path.previous)) return;
    const node = await wb.createNode(name);
    node.parent = ctx.node.parent;
    node.siblingIndex = ctx.node.siblingIndex + 1;
    m.redraw.sync();
    const p = ctx.path.clone();
    p.pop();
    wb.focus(p.append(node));
  },
  "insert-before": async (wb, ctx) => {
    if (!ctx.node) return;
    // if (ctx.path.previous && objectManaged(ctx.path.previous)) return;
    const node = await wb.createNode("");
    node.parent = ctx.node.parent;
    node.siblingIndex = ctx.node.siblingIndex;
    m.redraw.sync();
    const p = ctx.path.clone();
    p.pop();
    wb.focus(p.append(node));
  },
  "insert-child": async (wb, ctx, name = "", siblingIndex = undefined) => {
    if (!ctx.node) return;
    // if (objectManaged(ctx.node)) return;
    const node = await wb.createNode(name);
    node.parent = ctx.node;
    if (siblingIndex !== undefined) {
      node.siblingIndex = siblingIndex;
    }
    wb.setExpanded(ctx.path.head, ctx.node, true);
    m.redraw.sync();
    wb.focus(ctx.path.append(node), name.length);
  },
  "delete": async (wb, ctx) => {
    if (!ctx.node) return;
    // if (ctx.node.id.startsWith("@")) return;
    // if (ctx.path.previous && objectManaged(ctx.path.previous)) return; // should probably provide feedback or disable delete
    const above = wb.findAbove(ctx.path);
    ctx.node.destroy();
    m.redraw.sync();
    if (above) {
      let pos = 0;
      if (ctx.event && ctx.event.key === "Backspace") {
        if (above.node.value) {
          pos = above.node.value.length;
        } else {
          pos = above.node.name.length;
        }
      }
      if (above.node.childCount === 0) {
        // TODO: use subCount
        wb.setExpanded(ctx.path.head, above.node, false);
      }
      wb.focus(above, pos);
    }
  },
}

class Workbench {
    constructor(path, peer, realm) {
        this.context = {
          path: path,
          node: path.node,
        };
        this.root = path;
        this.clipboard = {};
        this.expanded = {};
        this.peer = peer;
        this.realm = realm;
    }

    async createNode(name) {
        return await this.realm.create(name);
    }

    showMenu() {
        // Implementation here
        console.log("TODO");
    }

    // findAbove(path: Path): Path | null {
    findAbove(path /*: Path*/) /*: Path | null */ {
        if (path.node.id === path.head.id) {
          return null;
        }
        const p = path.clone();
        p.pop(); // pop to parent
        let prev = path.node.prevSibling;
        if (!prev) {
          // if not a field and parent has fields, return last field
        //   const fieldCount = path.previous.getLinked("Fields").length;
        //   if (path.node.raw.Rel !== "Fields" && fieldCount > 0) {
        //     return p.append(path.previous.getLinked("Fields")[fieldCount - 1]);
        //   }
          // if no prev sibling, and no fields, return parent
          return p;
        }
        // const lastSubIfExpanded = (p: Path): Path => {
        const lastSubIfExpanded = (p /*: Path*/) /*: Path */ => {
          const expanded = this.getExpanded(path.head, p.node);
          if (!expanded) {
            // if not expanded, return input path
            return p;
          }
        //   const fieldCount = p.node.getLinked("Fields").length;
        //   if (p.node.childCount === 0 && fieldCount > 0) {
        //     const lastField = p.node.getLinked("Fields")[fieldCount - 1];
        //     // if expanded, no children, has fields, return last field or its last sub if expanded
        //     return lastSubIfExpanded(p.append(lastField));
        //   }
          if (p.node.childCount === 0) {
            // expanded, no fields, no children
            return p;
          }
          const lastChild = p.node.children[p.node.childCount - 1];
          // return last child or its last sub if expanded
          return lastSubIfExpanded(p.append(lastChild));
        }
        // return prev sibling or its last child if expanded
        return lastSubIfExpanded(p.append(prev));
    }

    //   findBelow(path: Path): Path | null {
    findBelow(path /*: Path*/) /*: Path | null */ {
      // TODO: find a way to indicate pseudo "new" node for expanded leaf nodes
      const p = path.clone();
      // if (this.getExpanded(path.head, path.node) && path.node.getLinked("Fields").length > 0) {
      //   // if expanded and fields, return first field
      //   return p.append(path.node.getLinked("Fields")[0]);
      // }
      if (this.getExpanded(path.head, path.node) && path.node.childCount > 0) {
        // if expanded and children, return first child
        return p.append(path.node.children[0]);
      }
      // const nextSiblingOrParentNextSibling = (p: Path): Path | null => {
      const nextSiblingOrParentNextSibling = (p /*: Path*/) /*: Path | null */ => {
        const next = p.node.nextSibling;
        if (next) {
          p.pop(); // pop to parent
          // if next sibling, return that
          return p.append(next);
        }
        const parent = p.previous;
        if (!parent) {
          // if no parent, return null
          return null;
        }
      //   if (p.node.raw.Rel === "Fields" && parent.childCount > 0) {
      //     p.pop(); // pop to parent
      //     // if field and parent has children, return first child
      //     return p.append(parent.children[0]);
      //   }
        p.pop(); // pop to parent
        // return parents next sibling or parents parents next sibling
        return nextSiblingOrParentNextSibling(p);
      }
      // return next sibling or parents next sibling
      return nextSiblingOrParentNextSibling(p);
    }

    getExpanded(head, n) {
        if (!this.expanded[head.id]) {
            this.expanded[head.id] = {};
        }
        let expanded = this.expanded[head.id][n.id];
        if (expanded === undefined) {
            expanded = false;
        }
        return expanded;
    }

    setExpanded(head, n, b) {
        if (!this.expanded[head.id]) {
          this.expanded[head.id] = {};
        }
        this.expanded[head.id][n.id] = b;
        // this.save();
    }

    executeCommand(command, ...args) {
      const fn = commands[command];

      if (!fn) {
        throw new Error(`Unknown command: "${command}"`);
      }

      console.log(command);
      return fn(this, ...args);
    }

    defocus() {
      const input = this.getInput(this.context.path);
      if (input) {
        input.blur();
      }
      this.context.node = null;
      this.context.path = null;
    }

  //focus(path: Path, pos: number = 0) {
  focus(path, pos = 0) {
    const input = this.getInput(path);
    if (input) {
      this.context.path = path;
      input.focus();
      if (pos !== undefined) {
        input.setSelectionRange(pos,pos);
      }
    } else {
      console.warn("unable to find input for", path);
    }
  }

  //getInput(path: Path): any {
  getInput(path) /*: any */ {
    let id = `input-${path.id}-${path.node.id}`;
    // kludge:
    // if (path.node.raw.Rel === "Fields") {
    //   if (path.node.name !== "") {
    //     id = id+"-value"; 
    //   }
    // }
    const el = document.getElementById(id);
    if (el?.editor) {
      return el.editor;
    }
    return el;
  }
}

export const NewNode = {
  view({attrs: {workbench, path}}) {
    // const node = path.node;
    const keydown = (e) => {
      if (e.key === "Tab") {
        e.stopPropagation();
        e.preventDefault();
        if (node.childCount > 0) {
          const lastchild = path.node.children[path.node.childCount-1];
          workbench.executeCommand("insert-child", {node: lastchild, path});
        }
      } else {
        workbench.executeCommand("insert-child", {node: path.node, path}, e.target.value);
      }
    }
    return (
      <div class="NewNode flex flex-row items-center">
        <svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16">
          <circle cx="8" cy="7" r="7" />
          <path style={{transform: "translate(0px, -1px)"}} d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
        </svg>
        <div class="flex grow">
          <input class="grow"
            type="text"
            onkeydown={keydown}
            value={""}
          />
        </div>
      </div>
    )
  }
}

export const NodeEditor = {
  view ({attrs: {workbench, path, onkeydown, oninput, disallowEmpty, editValue, placeholder}, state}) {
    const node = path.node;
    let prop = (editValue) ? "value" : "name";
    
    const display = () => {
    //   if (prop === "name") {
    //     return objectHas(node, "displayName") ? objectCall(node, "displayName", node) : node.name;
    //   }
      return node[prop] || "";
    }
    const onfocus = () => {
      state.initialValue = node[prop];
      workbench.context.node = node;
      workbench.context.path = path;
    }
    const getter = () => {
      return node[prop];
    }
    const setter = (v, finished) => {
      if (!node.isDestroyed) {
        if (disallowEmpty && v.length === 0) {
          node[prop] = state.initialValue;
        } else {
          node[prop] = v;
        }
      }
      if (finished) {
        workbench.context.node = null;
      }
    }

    // if (node.raw.Rel === "Fields") {
    //   placeholder = (editValue) ? "Value" : "Field";
    // }
    
    let id = `input-${path.id}-${node.id}`;
    if (prop === "value") {
      id = id+"-value";
    }
    let editor = TextAreaEditor;
    // if (node.parent && node.parent.hasComponent(Document) && window.Editor) {
    //   editor = CodeMirrorEditor;
    // }
    let desc = undefined;
    // if (node.hasComponent(Description)) {
    //   desc = node.getComponent(Description);
    // }
    return (
      <div class="NodeEditor flex flex-col">
        {m(editor, {id, getter, setter, display, onkeydown, onfocus, oninput, placeholder, workbench, path})}    
        {(desc) ? m(desc.editor(), {node}) :null}
      </div>
    )
  }
}


export const TextAreaEditor = {
  oncreate({dom,attrs}) {
    const textarea = dom.querySelector("textarea");
    const initialHeight = textarea.offsetHeight;
    const span = dom.querySelector("span");
    this.updateHeight = () => {
      span.style.width = `${Math.max(textarea.offsetWidth, 100)}px`;
      span.innerHTML = textarea.value.replace("\n", "<br/>");
      let height = span.offsetHeight;
      if (height === 0 && initialHeight > 0) {
        height = initialHeight;
      }
      textarea.style.height = (height > 0) ? `${height}px` : `var(--body-line-height)`;
    }
    textarea.addEventListener("input", () => this.updateHeight());
    textarea.addEventListener("blur", () => span.innerHTML = "");
    setTimeout(() => this.updateHeight(), 50);
    if (attrs.onmount) attrs.onmount(textarea);
  },
  onupdate() {
    this.updateHeight();
  },
  view ({attrs: {id, onkeydown, onfocus, onblur, oninput, getter, setter, display, placeholder, path, workbench}, state}) {
    const value = (state.editing) 
      ? state.buffer 
      : (display) ? display() : getter();
    
    const defaultKeydown = (e) => {
      if (e.key === "Enter") {
        e.preventDefault();
        e.stopPropagation();
      }
    }
    const startEdit = (e) => {
      if (onfocus) onfocus(e);
      state.editing = true;
      state.buffer = getter();
    }
    const finishEdit = (e) => {
      // safari can trigger blur more than once
      // for a given element, namely when clicking
      // into devtools. this prevents the second 
      // blur setting node name to undefined/empty.
      if (state.editing) {
        state.editing = false;
        setter(state.buffer, true);
        state.buffer = undefined;
      }
      if (onblur) onblur(e);
    }
    const edit = (e) => {
      state.buffer = e.target.value;
      setter(state.buffer, false);
      if (oninput) {
        oninput(e);
      }
    }
    // const handlePaste = (e) => {
    //   const textData = e.clipboardData.getData('Text');
    //   if (textData.length > 0) {
    //     e.preventDefault();
    //     e.stopPropagation();

    //     const lines = textData.split('\n').map(line => line.trim()).filter(line => line.length > 0);
    //     state.buffer = lines.shift();
    //     setter(state.buffer, true);

    //     let node = path.node;
    //     for (const line of lines) {
    //       const newNode = workbench.workspace.new(line);
    //       newNode.parent = node.parent;
    //       newNode.siblingIndex = node.siblingIndex + 1;
    //       m.redraw.sync();
    //       const p = path.clone();
    //       p.pop();
    //       workbench.focus(p.append(newNode));
    //       node = newNode;
    //     }
    //   }
    // }
    
    return (
      <div class="TextAreaEditor">
        <textarea
          id={id}
          rows="1"
          onfocus={startEdit}
          onblur={finishEdit}
          oninput={edit}
        //   onpaste={handlePaste}
          placeholder={placeholder}
          onkeydown={onkeydown||defaultKeydown}
          value={value}>{value}</textarea>
        <span style={{visibility: "hidden", position: "fixed"}}></span>
      </div>
    )
  }
}

// handles: expanded state, node menu+handle, children
// <OutlineNode path={node:{id:"", childCount: 0, name: "", children: []}, append: () => null} workbench={{showMenu: () => {}}} />
export const OutlineNode = {
    view ({attrs, state, children}) {
      let {path, workbench} = attrs;
      let node = path.node;
  
      let isRef = false;
      let handleNode = node;
      if (node.refTo) {
        isRef = true;
        node = handleNode.refTo;
      }
  
      let isCut = false;
    //   if (workbench.clipboard && workbench.clipboard.op === "cut") {
    //     if (workbench.clipboard.node.id === node.id) {
    //       isCut = true;
    //     }
    //   }
  
      const expanded = workbench.getExpanded(path.head, handleNode);
      const placeholder = ''; //objectHas(node, "handlePlaceholder") ? objectCall(node, "handlePlaceholder") : '';
  
      const hover = (e) => {
        state.hover = true;
        e.stopPropagation();
      }
      
      const unhover = (e) => {
        state.hover = false;
        e.stopPropagation();
      }
          
  
      const cancelTagPopover = () => {
        // if (state.tagPopover) {
        //   workbench.closePopover();
        //   state.tagPopover = undefined;
        // }
      }
  
      const oninput = (e) => {
        // if (state.tagPopover) {
        //   state.tagPopover.oninput(e);
        //   if (!e.target.value.includes("#")) {
        //     cancelTagPopover();
        //   }
        // } else {
        //   if (e.target.value.includes("#")) {
        //     state.tagPopover = {};
        //     // Don't love that we're hard depending on Tag
        //     Tag.showPopover(workbench, path, node, (onkeydown, oninput) => {
        //       state.tagPopover = {onkeydown, oninput};
        //     }, cancelTagPopover);
        //   }
        // }
      }
  
      const onkeydown = (e) => {
        if (state.tagPopover) {
          if (e.key === "Escape") {
            cancelTagPopover();
            return;
          }
          if (state.tagPopover.onkeydown(e) === false) {
            e.stopPropagation();
            return false;
          }
        }
        const anyModifiers = e.shiftKey || e.metaKey || e.altKey || e.ctrlKey;
        switch (e.key) {
        case "ArrowUp":
          if (e.target.selectionStart !== 0 && !anyModifiers) {
            e.stopPropagation()
          }
          break;
        case "ArrowDown":
          if (e.target.selectionStart !== e.target.value.length && e.target.selectionStart !== 0 && !anyModifiers) {
            e.stopPropagation()
          }
          break;
        case "Backspace":
          // cursor at beginning of empty text
          if (e.target.value === "") {
            e.preventDefault();
            e.stopPropagation();
            if (node.childCount > 0) {
              return;
            }
            workbench.executeCommand("delete", {node, path, event: e});
            return;
          }
          // cursor at beginning of non-empty text
          if (e.target.value !== "" && e.target.selectionStart === 0 && e.target.selectionEnd === 0) {
            e.preventDefault();
            e.stopPropagation();
            if (node.childCount > 0) {
              return;
            }
            
            // TODO: make this work as a command?
            const above = workbench.findAbove(path);
            if (!above) {
              return;
            }
            const oldName = above.node.name;
            above.node.name = oldName+e.target.value;
            node.destroy();
            m.redraw.sync();
            workbench.focus(above, oldName.length);
            
            return;
          }
          break;
        case "Enter":
          e.preventDefault();
          if (e.ctrlKey || e.shiftKey || e.metaKey || e.altKey) return;
          
          // first check if node should become a code block
          // todo: this should be a hook or some loose coupled system
        //   if (e.target.value.startsWith("```") && !node.hasComponent(CodeBlock)) {
        //     const lang = e.target.value.slice(3);
        //     if (lang) {
        //       workbench.executeCommand("make-code-block", {node, path}, lang);
        //       e.stopPropagation();
        //       return;
        //     }
        //   }
  
          // cursor at end of text
          if (e.target.selectionStart === e.target.value.length) {
            if (node.childCount > 0 && workbench.getExpanded(path.head, node)) {
              workbench.executeCommand("insert-child", {node, path}, "", 0);
            } else {
              workbench.executeCommand("insert", {node, path});
            }
            e.stopPropagation();
            return;
          }
          // cursor at beginning of text
          if (e.target.selectionStart === 0) {
            workbench.executeCommand("insert-before", {node, path});
            e.stopPropagation();
            return;
          }
          // cursor in middle of text
          if (e.target.selectionStart > 0 && e.target.selectionStart < e.target.value.length) {
            workbench.executeCommand("insert", {node, path}, e.target.value.slice(e.target.selectionStart)).then(() => {
              node.name = e.target.value.slice(0, e.target.selectionStart);
            });
            e.stopPropagation();
            return;
          }
          break;
        }
      }
  
      const open = (e) => {
        e.preventDefault();
        e.stopPropagation();
        
        workbench.executeCommand("zoom", {node, path});
        
        // clear text selection that happens after from double click
        if (document.selection && document.selection.empty) {
          document.selection.empty();
        } else if (window.getSelection) {
          window.getSelection().removeAllRanges();
        }
      }
  
      const toggle = (e) => {
        // TODO: hook or something so to not hardcode
        // if (node.hasComponent(Document)) {
        //   open(e);
        //   return;
        // }
        if (expanded) {
          workbench.executeCommand("collapse", {node: handleNode, path});
        } else {
          workbench.executeCommand("expand", {node: handleNode, path});
        }
        e.stopPropagation();
      }
  
      const subCount = (n) => {
        return n.childCount;
      }
  
      const showHandle = () => {
        if (node.id === workbench.context?.node?.id || state.hover) {
          return true;
        }
        if (node.name.length > 0) return true;
        if (placeholder.length > 0) return true;
      }
  
      return (
        <div onmouseover={hover} onmouseout={unhover} id={`node-${handleNode.id}`} class={`OutlineNode ${isCut ? "cut-node" : ""}`}>
          <div class="node-row-outer-wrapper flex flex-row items-start">
            <svg class="node-menu shrink-0" xmlns="http://www.w3.org/2000/svg"
                onclick={(e) => workbench.showMenu(e, {node: handleNode, path})}
                oncontextmenu={(e) => workbench.showMenu(e, {node: handleNode, path})} 
                data-menu="node"
                viewBox="0 0 16 16">
              {state.hover && <path style={{transform: "translateY(-1px)"}} fill="currentColor" fill-rule="evenodd" d="M2.5 12a.5.5 0 0 1 .5-.5h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5zm0-4a.5.5 0 0 1 .5-.5h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5zm0-4a.5.5 0 0 1 .5-.5h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5z" />}
            </svg>
            <div class="node-handle shrink-0" onclick={toggle} ondblclick={open} oncontextmenu={(e) => workbench.showMenu(e, {node: handleNode, path})} data-menu="node" style={{ display: showHandle() ? 'block' : 'none' }}>
              {/* {(objectHas(node, "handleIcon"))
                ? objectCall(node, "handleIcon", subCount(node) > 0 && !expanded)
                :  }*/}
                <svg class="node-bullet" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
                  {(subCount(node) > 0 && !expanded)?<circle id="node-collapsed-handle" cx="8" cy="8" r="8" />:null}
                  <circle cx="8" cy="8" r="3" fill="currentColor" />,
                  {(isRef)?<circle id="node-reference-handle" cx="8" cy="8" r="7" fill="none" stroke-width="1" stroke="currentColor" stroke-dasharray="3,3" />:null}
                </svg>
            </div>
            {/* {(node.raw.Rel === "Fields") 
              ? <div class="flex grow items-start flex-row">
                  <div>
                    <NodeEditor workbench={workbench} path={path} onkeydown={onkeydown} oninput={oninput} />
                  </div>
                  <NodeEditor editValue={true} workbench={workbench} path={path} onkeydown={onkeydown} oninput={oninput} />
                </div>
              : } */}
              <div class="flex grow items-start flex-row" style={{gap: "0.5rem"}}>
                  {/* {objectHas(node, "beforeEditor") && componentsWith(node, "beforeEditor").map(component => m(component.beforeEditor(), {node, component}))} */}
                  <NodeEditor workbench={workbench} path={path} onkeydown={onkeydown} oninput={oninput} placeholder={placeholder} />
                  {/* {objectHas(node, "afterEditor") && componentsWith(node, "afterEditor").map(component => m(component.afterEditor(), {node, component}))} */}
                </div>
          </div>
          {/* {objectHas(node, "belowEditor") && componentsWith(node, "belowEditor").map(component => m(component.belowEditor(), {node, component, expanded}))} */}
          {(expanded === true) &&
            <div class="expanded-node flex flex-row">
              <div class="indent flex" onclick={toggle}></div>
              <div class="view grow">
                {/* {objectHas(node, "childrenView")
                  ? m(componentsWith(node, "childrenView")[0].childrenView(), {workbench, path})
                  : } */}
                  {m(getNodeView(node), {workbench, path})}
              </div>
            </div>
          }
        </div>
      )
    }
  };

  function getNodeView(node) {
    return listView; //views[node.getAttr("view") || "list"] || empty;
  }

  const listView = {
    view({attrs: {workbench, path, alwaysShowNew}}) {
      let node = path.node;
      if (path.node.refTo) {
        node = path.node.refTo;
      }
      let showNew = false;
      if ((node.childCount === 0) || alwaysShowNew) {
        showNew = true;
      }
      // TODO: find some way to not hardcode this rule
    //   if (node.hasComponent(SmartNode)) {
    //     showNew = false;
    //   }
      return (
        <div class="list-view">
          <div class="children">
            {(node.childCount > 0) && node.children.map(n => <OutlineNode key={n.id} workbench={workbench} path={path.append(n)} />)}
            {showNew && m(NewNode, {workbench, path})}
          </div>
        </div>
      )
    }
  }

export const OutlineEditor = {
    view ({attrs: {workbench, path, alwaysShowNew}}) {
    //   return objectHas(path.node, "childrenView")
    //     ? m(componentsWith(path.node, "childrenView")[0].childrenView(), {workbench, path})
    //     : 
        return m(getNodeView(path.node), {workbench, path, alwaysShowNew});
    }
}

/**
 * Path is a stack of nodes. It can be used as a history stack
 * so you can "zoom" into subnodes and return back to previous nodes.
 * It is also used to identify nodes in the UI more specifically than
 * the node ID since a node can be shown more than once (references, panels, etc).
 */
export class Path {
  constructor(head, name) {
    if (name) {
      this.name = name;
    } else {
      this.name = Math.random().toString(36).substring(2);
    }
    if (head) {
      this.nodes = [head];
    } else {
      this.nodes = [];
    }
  }

  push(node) {
    this.nodes.push(node);
  }

  pop() {
    return this.nodes.pop() || null;
  }

  // chroot?
  sub() {
    return new Path(this.node, this.name);
  }

  clone() {
    const p = new Path();
    p.name = this.name;
    p.nodes = [...this.nodes];
    return p;
  }

  append(node) {
    const p = this.clone();
    p.push(node);
    return p;
  }

  get length() {
    return this.nodes.length;
  }

  get id() {
    return [this.name, ...this.nodes.map(n => n.id)].join(":");
  }

  get node() {
    return this.nodes[this.nodes.length-1];
  }

  get previous() {
    if (this.nodes.length < 2) return null;
    return this.nodes[this.nodes.length-2];
  }

  get head() {
    return this.nodes[0];
  }
}