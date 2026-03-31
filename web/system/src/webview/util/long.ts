// long module provides wrapper for localstorage to keep
// component state 'long' term, like expandable elements
// that should retain their open/close state across runs.
// (see expand.ts)

export function uniqueIdent(vnode: any, suffix: any): string {
  return [dirname(getPathTo(vnode.dom)) + "|" + vnode.tag.name, "?" + vnode.key, "#" + suffix]
    .join(" ")
    .replace(/ /g, "-");
}

export function get(state: any, initial: any, suffix?: any): any {
  const raw = localStorage.getItem(`--tractor-${state.ident}${suffix ? "-" + suffix : ""}`);
  if (raw == null) return initial;
  return JSON.parse(raw) || initial;
}

export function set(state: any, value: any, suffix?: any): void {
  localStorage.setItem(`--tractor-${state.ident}${suffix ? "-" + suffix : ""}`, JSON.stringify(value));
}

function dirname(path: any): any {
  return path.match(/.*\//) || "";
}

function getPathTo(element: any): any {
  if (!element) return "";
  if (element.id !== "") return "*[@id=" + element.id + "]";

  if (element === document.body) return element.tagName.toLowerCase();

  let ix = 0;
  const siblings = element.parentNode.childNodes;
  for (let i = 0; i < siblings.length; i++) {
    const sibling = siblings[i];

    if (sibling === element)
      return getPathTo(element.parentNode) + "/" + element.tagName.toLowerCase() + "[" + (ix + 1) + "]";

    if (sibling.nodeType === 1 && sibling.tagName === element.tagName) {
      ix++;
    }
  }
}
