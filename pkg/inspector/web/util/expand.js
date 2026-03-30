// expand module provides utility functions for keeping
// state of expandable elements like accordians or treeviews

// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as long from "./long.js";

export function initExpanded(v, identAttr) {
  let suffix = v.attrs[identAttr];
  if (v.attrs.data && v.attrs.data[identAttr]) {
    suffix = v.attrs.data[identAttr];
  }
  v.state.ident = long.uniqueIdent(v, suffix);
  v.state.expanded = long.get(v.state, v.attrs.expanded)
  ui.redraw();
}

export function isExpanded(state, attrs) {
  if (state.expanded === undefined) {
    return attrs.expanded;
  }
  return state.expanded;
}

export function toggler(state, ontoggle) {
  return (e) => {
    if (state.expanded) {
      state.expanded = false;
    } else {
      state.expanded = true;
    }
    e.expanded = state.expanded;
    long.set(state, state.expanded);
    if (ontoggle) ontoggle(e);
  };
}
