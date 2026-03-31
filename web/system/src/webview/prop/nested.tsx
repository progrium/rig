import m from "mithril";
import * as atom from "../elem/atom/mod";
import * as molecule from "../elem/expander";
import * as expand from "../util/expand";

interface NestedAttrs {
  label: string;
  expanded: boolean;
  icon?: string;
  extra?: any;
  pre?: any;
  ongrip?: (e: any) => void;
}

export class Nested implements m.ClassComponent<NestedAttrs> {
  ident?: string;
  expanded?: boolean;

  oncreate(vnode: m.CVnodeDOM<NestedAttrs>) {
    expand.initExpanded(vnode, "label");
  }

  view(vnode: m.CVnode<NestedAttrs>) {
    const { attrs, children } = vnode;
    const state = vnode.state as Nested;
    const ongrip = attrs.ongrip || undefined;
    const icon = attrs.icon || undefined;
    const pre = attrs.pre || undefined;
    const extra = attrs.extra || undefined;

    const label = attrs.label.includes(".")
      ? attrs.label
      : (attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g) || [attrs.label]).join(" ");
    const expanded = expand.isExpanded(state, attrs);

    const toggle = expand.toggler(state);

    const vstackStyle = {
      WebkitUserSelect: "none",
      userSelect: "none" as const,
    };

    return (
      <atom.VStack style={vstackStyle}>
        <molecule.Expander expanded={expanded} onclick={toggle}>
          {pre && <atom.HStack>{pre}</atom.HStack>}
          <atom.HStack title={label} align="center" style={{ flexGrow: "1" }}>
            {icon && (
              <atom.Icon
                style={{ marginRight: ".5rem", marginLeft: ".25rem" }}
                name={icon}
              />
            )}
            <atom.Label style={{ flexGrow: "1" }} onclick={toggle} text={label} />
            {ongrip && <atom.Dots style={{ flexShrink: "1" }} rows={3} />}
          </atom.HStack>
          {extra && <atom.HStack>{extra}</atom.HStack>}
        </molecule.Expander>
        <div style={{ marginLeft: ".75rem", marginBottom: "0.50rem" }}>
          {expanded && children}
        </div>
      </atom.VStack>
    );
  }
}
