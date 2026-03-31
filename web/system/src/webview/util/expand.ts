// expand module provides utility functions for keeping
// state of expandable elements like accordians or treeviews

import m from "mithril";
import * as long from "./long";

export function initExpanded(v: m.CVnodeDOM<any>, identAttr: string): void {
  let suffix = (v.attrs as any)[identAttr];
  if ((v.attrs as any).data && (v.attrs as any).data[identAttr]) {
    suffix = (v.attrs as any).data[identAttr];
  }
  (v.state as any).ident = long.uniqueIdent(v, suffix);
  (v.state as any).expanded = long.get(v.state, (v.attrs as any).expanded);
  m.redraw();
}

export function isExpanded(state: any, attrs: { expanded?: boolean }): boolean {
  if (state.expanded === undefined) {
    return !!attrs.expanded;
  }
  return state.expanded;
}

export function toggler(state: any, ontoggle?: (e: any) => void): (e: any) => void {
  return (e: any) => {
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
