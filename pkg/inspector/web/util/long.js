// long module provides wrapper for localstorage to keep
// component state 'long' term, like expandable elements
// that should retain their open/close state across runs.
// (see expand.js)

export function uniqueIdent(vnode, suffix) {
  return [dirname(getPathTo(vnode.dom)) + "|"+ vnode.tag.name, "?"+vnode.key, "#"+suffix].join(" ").replace(/ /g, "-");
}

export function get(state, initial, suffix) {
  return JSON.parse(localStorage.getItem(`--tractor-${state.ident}${(suffix)?"-"+suffix:""}`)) || initial;
}

export function set(state, value, suffix) {
  localStorage.setItem(`--tractor-${state.ident}${(suffix)?"-"+suffix:""}`, JSON.stringify(value));
}

function dirname(path) {
  return path.match(/.*\//) || "";
}

function getPathTo(element) {
  if (!element) return "";
  if (element.id !== '')
      return "*[@id=" + element.id + "]";

  if (element === document.body)
      return element.tagName.toLowerCase();

  var ix = 0;
  var siblings = element.parentNode.childNodes;
  for (var i = 0; i < siblings.length; i++) {
      var sibling = siblings[i];

      if (sibling === element) return getPathTo(element.parentNode) + '/' + element.tagName.toLowerCase() + '[' + (ix + 1) + ']';

      if (sibling.nodeType === 1 && sibling.tagName === element.tagName) {
          ix++;
      }
  }
}