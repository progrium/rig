/* @jsx h */
import m from "/-/inspector/mithril/mithril-2.0.4.mjs"

// m.render(element, vnodes, redraw)
export const render = m.render

// m.mount(element, Component)
export const mount = m.mount

// vnode = m.trust(html)
export const trust = m.trust

// vnode = m.fragment(attrs, children)
export const fragment = m.fragment

// m.redraw()
export const redraw = () => {
  if (isPaused) return
  m.redraw()
}

let isPaused = false
export const pause = () => {
  isPaused = true
}
export const resume = () => {
  isPaused = false
  redraw()
}

export const classes = (names, conditionals={}) => {
  return [names, 
    Object.keys(conditionals)
      .filter(key => conditionals[key])
      .join(" ")
  ].join(" ")
}

// vnode = m(selector, attrs, children)
export function h() { 
  let vnode = m.m.apply(this, arguments)
  if (typeof vnode.tag === 'object' && vnode.tag.view && !vnode.tag.oldview) {
    // we wrap the component view function
    // to extend/modify the model slightly.
    // not convinced this should be here and
    // not an opt-in API like the legacy Element
    // class, but trying it here again.
    vnode.tag.oldview = vnode.tag.view
    vnode.tag.view = (input) => { 
      // 1. add self reference for destructuring
      let output = vnode.tag.oldview(Object.assign(input, {vnode}))
      if (!output) {
        return
      }
      // 2. allow common element attributes to pass through
      output.attrs = mergeElemAttrs(input.attrs||{}, output.attrs||{}, 
        vnode.tag.passthrough ? vnode.tag.passthrough() : undefined)
      // 3. allow all data- attributes to pass through
      output.attrs = mergeDataAttrs(input.attrs||{}, output.attrs||{})
      return output
    }
  }
  return vnode
}

function mergeElemAttrs(input, output, allowed=undefined) {
  // use this to define which attributes are allowed, and
  // how they are merged/chosen between inner/outer element
  const elemAttrs = {
    "id":       (i,o) => ["id", i["id"]],
    "title":    (i,o) => ["title", i["title"]],
    "class":    (i,o) => ["className", [i["class"], o["className"]].join(" ").trim()],
    "style":    (i,o) => ["style", Object.assign(o["style"]||{}, i["style"]||{})],
    
    "onclick":      (i,o) => ["onclick", i["onclick"]],
    "onmouseover":  (i,o) => ["onmouseover", i["onmouseover"]],
    "onmouseout":   (i,o) => ["onmouseout", i["onmouseout"]],
    "onmousedown":   (i,o) => ["onmousedown", i["onmousedown"]],
  }
  return Object.assign(output, Object.fromEntries(
    Object.keys(elemAttrs)
      .filter(attr => allowed ? attr in allowed : true)
      .filter(attr => attr in input)
      .map(attr => elemAttrs[attr](input, output))
  ))
}

function mergeDataAttrs(input, output) {
  return Object.assign(output, Object.fromEntries(
    Object.keys(input)
      .filter(attr => attr.startsWith("data-"))
      .map(attr => [attr, input[attr]])
  ))
}

//
// BELOW IS LEGACY VERSION OF ABOVE AND IS DEPRECATED
//

// attributes allowed to be implicitly passed from custom
// component/element attributes to the outer element rendered
// by the component. 
const allowedAttrs = [
  "id", 
  "class", 
  "style",
  "title",
  "onclick", 
  "ondblclick",
  "oncontextmenu",
  "onmouseout",
  "onmouseover",
  "onmousedown",
  "onmouseup",
];

// Element is a mithril component base class that allows Elements to
// implicitly pass attributes from allowedAttrs and all "data-" attributes
// to the vnode returned by onrender, unless already set by onrender. 
// Allowed event attributes starting with "on" are only passed if not used
// as an attribute in onrender, as detected by attrProxy.
export class Element {
  view(outer) {
    outer.attrs = attrProxy(outer.attrs);
    outer.vnode = outer;

    let inner = this.onrender(outer);
    //let before = Object.assign({}, inner);
    
    inner.attrs = inner.attrs || {}; 
    allowedAttrs.forEach((attr) => {
      if (!outer.attrs.hasOwnProperty(attr)) return;
      
      if (attr.startsWith("on") && outer.attrs._used.has(attr)) return;
      
      if (attr === "class") {
        inner.attrs["className"] = [inner.attrs["className"], outer.attrs["class"]].join(" "); 
        return;
      }
      if (attr === "style") {
        inner.attrs["style"] = Object.assign(inner.attrs["style"]||{}, outer.attrs["style"]); 
        return;
      }
      
      if (inner.attrs.hasOwnProperty(attr)) return;
      
      inner.attrs[attr] = outer.attrs[attr]; 
    })

    for (const attr in outer.attrs) {
      if (attr.startsWith("data-")) {
        inner.attrs[attr] = outer.attrs[attr];
      }
    }

    //console.log(Object.assign({}, outer), before, Object.assign({}, inner));

    return inner;
  }
}

// finds out if allowed attrs starting with "on" are used.
function attrProxy(attrs) {
  return new Proxy(attrs, {
      get: function (target, prop, receiver) {
          if (!this.used) {
              this.used = new Set();
          }
          if (prop === "_used") {
              return this.used;
          }
          if (prop.startsWith("on") && allowedAttrs.includes(prop)) {
              this.used.add(prop);
          }
          return Reflect.get(...arguments);
      },
  })
}