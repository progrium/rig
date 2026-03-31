import m from "mithril";
import * as atom from "../elem/atom/mod";
import * as molecule from "../elem/expander";
import * as expand from "../util/expand";

interface CollectionAttrs {
  label: string;
  expanded: boolean;
  items: any[];
  input: any;
  renderer: (item: any, idx: number) => any;
  onadd?: (e: any, el: NodeListOf<HTMLInputElement>) => void;
}

export class Collection implements m.ClassComponent<CollectionAttrs> {
  ident?: string;
  expanded?: boolean;
  adding?: boolean;

  oncreate(vnode: m.CVnodeDOM<CollectionAttrs>) {
    expand.initExpanded(vnode, "label");
  }

  view(vnode: m.CVnode<CollectionAttrs>) {
    const { attrs, children } = vnode;
    const state = vnode.state as Collection;
    const childArr = (children || []) as any[];

    const labelParts = attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g);
    const label = (labelParts || []).join(" ");
    const items = attrs.items || [];
    const input = attrs.input || <input type="text" />;
    const onadd = attrs.onadd || undefined;
    const renderer = attrs.renderer || ((item: any) => <div>{item}</div>);
    const expanded = expand.isExpanded(state, attrs);

    const addingInit = childArr.length === 0 && expanded;
    state.adding = state.adding === undefined ? addingInit : state.adding;

    const toggle = expand.toggler(state);
    const toggleAdd = (_e: any) => {
      state.adding = state.adding ? false : true;
    };

    const vstackStyle = {
      WebkitUserSelect: "none",
      userSelect: "none" as const,
    };

    const add = (e: any) => {
      const el = (e.target as HTMLElement).closest(".HStack");
      if (onadd && el) {
        onadd(e, el.querySelectorAll("input"));
      }
      state.adding = false;
    };

    return (
      <atom.VStack style={vstackStyle}>
        <molecule.Expander expanded={expanded} onclick={toggle}>
          <atom.Label style={{ flexGrow: "1" }} onclick={toggle} text={label} />
          <span style={{ color: "#888" }}>{items.length} items</span>
          <atom.Icon
            name="fas fa-plus-circle"
            onclick={toggleAdd}
            style={{ marginLeft: "0.5rem" }}
          />
        </molecule.Expander>
        {expanded ? (
          <div style={{ marginLeft: "1.5rem" }}>
            {state.adding ? (
              <atom.HStack>
                <atom.HStack style={{ flexGrow: "1" }}>{input}</atom.HStack>
                <button onclick={add}>Add</button>
              </atom.HStack>
            ) : null}
            {items.map((item, idx) => renderer(item, idx))}
          </div>
        ) : null}
      </atom.VStack>
    );
  }
}
