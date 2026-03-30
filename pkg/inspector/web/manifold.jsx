
import * as layout from "/-/backplane/layout.tsx"
import outline from "/-/vnd/heroicons/outline.jsx"

import * as draggable from "/-/util/draggable.ts"
import * as atom from "/-/elem/atom/mod.ts"
import * as manifold from "/-/manifold/mod.ts"
import * as ui from "/-/ui.js"
import * as input from "/-/backplane/input/mod.ts"

var clipboard=undefined

function popupMenu(items, ctxEl) {
  var id = 0
  var callbacks = {}
  const itemize = (o) => {
    id++
    let item = {}
    if (o.title) {
      item.ID = id
      item.Title = o.title
    }
    if (o.separator) {
      item.Separator = o.separator
    }
    if (o.disabled) {
      item.Disabled = true
    }
    if (o.onclick) {
      callbacks[id] = () => o.onclick(o, ctxEl)
    }
    if (o.submenu) {
      item.SubMenu = o.submenu.map(itemize)
    }
    return item
  }
  T.call("menu.popup", [items.map(itemize)]).then((resp) => {
    let cb = callbacks[resp.reply]
    if (cb) cb()
  })
}


// TODO: replace with data from library struct
function relatedComponents(id) {
  let mv = new manifold.View(ObjectProxy(T.state.Manifold), {})
  let obj = mv.findID(id)
  if (!obj) return []
  if (!Array.isArray(obj.Value)) {
    return []
  }
  let relatedNames = obj.Value
    .filter(c => c)
    .map(com => T.state.Components.filter(c => c.Name === com.Name)[0] )
    .filter(c => c)
    .map(c => c.Related || [])
    .flat()
  return [...new Set(relatedNames)]
    .map(name => T.state.Components.filter(c => c.Name === name)[0] )
    .filter(c => c)
}

function treeNode(node) {
  return {
    "id": node.ID,
    "label": node.Name || "(unnamed)",
    "icon": node.Icon||"",
    "children": (node.Children||[]).map((c) => treeNode(c))
  }
}


export const TreeView = {
  onupdate: ({state,dom}) => {
    if (state.edit) {
      dom.querySelector("input").value = state.edit
      dom.querySelector("input").select()
      state.edit = undefined
    }
  },
  view: ({attrs, state, vnode}) => {
    var data = treeNode(attrs.data || {ID:""})

    state.expanded = state.expanded || JSON.parse(localStorage.getItem("expanded")) || {}
    state.selected = state.selected || []
    state.editing = state.editing || undefined


    const isExpanded = (id) => {
      return state.expanded[id] !== undefined ? state.expanded[id] : false
    }

    const walk = (node, cb, level=0, sibIdx=0, parent=undefined, visible=true) => {
      cb(node, level, sibIdx, parent, visible)
      if (node.children && node.children.length > 0) {
        node.children.forEach((n,idx) => walk(n, cb, level+1, idx, node, visible && isExpanded(node.id)))
      }
    }

    const nodeList = []
    const nodes = {}

    walk(data, (n, lvl, idx, parent, visible) => {
      let node = {
        id: n.id,
        label: n.label,
        parent: parent,
        index: idx,
        level: lvl,
        children: n.children && n.children.length > 0,
        expanded: isExpanded(n.id),
        selected: state.selected.includes(n.id),
      }
      nodes[n.id] = node
      if (visible) {
        nodeList.push(node)
      }
    })


    const toggle = (e) => {
      e.stopPropagation()
      let id = e.target.closest(".node").id
      state.expanded[id] = !isExpanded(id)
      localStorage.setItem("expanded", JSON.stringify(state.expanded))
    }

    const select = (e) => {
      let id = e.target.closest(".node").id
      selectID(id)
    }

    const selectID = (id) => {
      if (!id) {
        state.selected = []
        T.execute("selection.Clear", {Args: []})
        return
      }
      console.log(relatedComponents(id))
      state.selected = [id]
      T.execute("selection.Select", {Args: [id]})
    }

    const contextMenu = (items) => {
      return (e) => {
        e.stopPropagation()
        select(e)
        popupMenu(items, e.target.closest("[data-id]")) 
        return false
      }
    }

    return (
      <div class="TreeView" style={{margin: "10px"}}>
        {nodeList.map(node => 
          <div class={ui.classes("node", {"selected": node.selected})} 
                key={node.id}
                id={node.id} 
                data-id={node.id}
                onclick={select}
                oncontextmenu={contextMenu([
                  {title: "New", submenu: [
                    {title: "Empty Object", onclick: async (item, el) => {
                      let id = await T.execute("backplane.Create", {Name: "EmptyObject", ParentID: el.dataset["id"]})
                      state.expanded[el.dataset["id"]] = true
                      selectID(id)
                      state.editing = id
                      state.edit = "Empty Object"
                      ui.redraw()
                    }},
                    {separator: true},
                    ...relatedComponents(node.id).map(com => {
                      return {
                        title: com.Name,
                        onclick: async (item, el) => {
                          let id = await T.execute("backplane.Create", {
                            Name: "EmptyObject", 
                            ParentID: el.dataset["id"], 
                            Components: [com.Name]
                          })
                          state.expanded[el.dataset["id"]] = true
                          selectID(id)
                          state.editing = id
                          state.edit = com.Name.split(".")[1]
                          ui.redraw()
                        }
                      }
                    }),
                  ]},
                  {separator: true},
                  {title: "Cut", onclick: (item, el) => {
                    clipboard = {cut: el.dataset["id"]}
                  }},
                  {title: "Copy", onclick: (item, el) => {
                    clipboard = {copy: el.dataset["id"]}
                  }},
                  {title: "Paste", onclick: async (item, el) => {
                    if (clipboard && clipboard.copy) {
                      let id = await T.execute("backplane.Duplicate", {ID: clipboard.copy, ParentID: el.dataset["id"]})
                      state.expanded[el.dataset["id"]] = true
                      selectID(id)
                      return
                    }
                    if (clipboard && clipboard.cut) {
                      let id = T.execute("backplane.Move", {ID: clipboard.cut, ParentID: el.dataset["id"]})
                      state.expanded[el.dataset["id"]] = true
                      selectID(id)
                      clipboard = undefined
                    }
                  }, disabled: (clipboard === undefined)},
                  {separator: true},
                  {title: "Duplicate", onclick: async (item, el) => {
                    let id = await T.execute("backplane.Duplicate", {ID: el.dataset["id"]})
                    selectID(id)
                  }},
                  {title: "Rename", onclick: (item, el) => {
                    state.editing = el.dataset["id"]
                    state.edit = nodes[el.dataset["id"]].label
                    ui.redraw()
                  }},
                  {title: "Delete", onclick: (item, el) => { 
                    T.execute("backplane.Delete", {ID: el.dataset["id"]})
                    selectID(null)
                  }},
                ])}
                onmousedown={draggable.start({
                  axis: "y",
                  parentClass: "TreeView",
                  indicatorFactory: draggable.lineIndicator,
                  placeholderFactory: draggable.linePlaceholder,
                  cursorFactory: (ctx) => {
                    const el = ctx.target.cloneNode(true)
                    el.style.paddingLeft = "0"
                    return el
                  },
                  oncomplete: (el, idx) => {
                    const node = nodes[el.id]
                    const nextNode = nodes[el.nextElementSibling.id]
                    let reparented = false
                    if (nextNode) {
                      if (nextNode.parent.id !== node.parent.id) {
                        reparented = true
                        node.parent = nextNode.parent
                        el.style.paddingLeft = `${0.75*nextNode.level}rem`
                      }
                      node.index = nextNode.index
                    }
                    //console.log(node.id, node.index, reparented ? node.parent.id : null)
                    T.execute("backplane.Move", {ID: node.id, Index: node.index, ParentID: reparented ? node.parent.id : ""})
                    selectID(node.id)
                  }
                })}
                style={{paddingLeft: `${0.75*node.level}rem`}}>
            <atom.HStack style={{paddingLeft: "1rem"}} title={node.id}>
              <atom.Icon 
                name={`fas fa-caret-${(node.expanded) ? 'down' : 'right'}`} 
                onclick={toggle}
                style={{
                  marginLeft: "-1rem",
                  position: "absolute",
                  display: node.children ? "block" : "none",
                }}/>
              {state.editing === node.id
                ? <input type="text"
                    onkeydown={(e) => {
                      if (e.key === "Enter") {
                        T.execute("backplane.Rename", {ID: node.id, Name: e.target.value})
                        state.editing = undefined
                      }
                      if (e.key === "Escape") {
                        state.editing = undefined
                      }
                    }}
                    onfocusout={(e) => {
                      T.execute("backplane.Rename", {ID: node.id, Name: e.target.value})
                      state.editing = undefined
                    }} />
                : node.label}
            </atom.HStack>
          </div>
        )}
      </div>
    )
  }
}