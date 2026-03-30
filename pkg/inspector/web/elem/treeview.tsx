// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as expand from "/-/inspector/util/expand.js";
// @ts-ignore
import * as atom from "./atom/mod.ts";
// @ts-ignore
import * as molecule from "./molecule.tsx";

interface NodeData {
  children: NodeData[];
  id?: string;
  label?: string;
  icon?: string;
  selected?: boolean;
  hasChildren?: boolean;
  onclick?: (e) => void;
}

interface NodeAttrs {
  onselect?: (e, data) => void;
  onexpand?: (e) => void;
  expanded?: boolean;
  selected?: string[];
  icon?: string;
  label?: string
  data?: NodeData;
}

interface NodeState {
  hover: boolean;
  expanded: boolean;
}

export const Node = {
  oncreate(v) {
    expand.initExpanded(v, "label");
  },
  onupdate() {
    
  },
  view({attrs, state, children}) {
    var data = attrs.data || {children: [], id: ""};
    var icon = attrs.icon || attrs.data.icon;
    var label = attrs.label || attrs.data.label;
    var selected = attrs.selected || [];
    var onexpand = attrs.onexpand;
    var onselect = attrs.onselect;
    var ondrag = attrs.ondrag;

    const expanded = expand.isExpanded(state, attrs);

    if (data.children) {
      children = data.children.map((data: NodeData) => 
        <Node key={data.id||data.label} {...{data, onselect, selected, onexpand, ondrag}} />);
    }
    
    const hasChildren: boolean = (data.hasChildren) ? data.hasChildren : data.children?.length > 0

    const toggle = expand.toggler(state, (e) => {
      if (attrs.data.onclick) attrs.data.onclick(e);
      if (onselect) onselect(e, attrs.data);
      if (attrs.onexpand && e.expanded) {
        attrs.onexpand(attrs.data);
      }
      e.stopPropagation()
    })

    return (
      <atom.VStack class={ui.classes("Node", {"open": expanded})} data-id={data.id} style={{"-webkit-user-select": "none"}}>
          <molecule.Expander expanded={expanded}
                             selected={selected.indexOf(data.id) > -1}
                             noexpand={!hasChildren}
                             oncontextmenu={oncontextmenu}
                             highlight={true}
                             onclick={toggle}
                             onmousedown={ondrag}>  
            {icon && <atom.Icon name={icon} />}
            <atom.Label text={label} />
          </molecule.Expander>
          <div class="children" style={{marginLeft: ".75rem"}}>
              {expanded && children}
          </div>
      </atom.VStack>
    ) 
  }
}

// const setupEditHandler = (v) => {
//   let el = v.dom.querySelector("div.label");
//   if (!el) return;
//   el.addEventListener("edit", (e) => {
//       state.editvalue = label;
//       state.onchange = (ee) => {
//           e.onchange(ee);
//           state.editvalue = undefined;
//           h.redraw();
//       };
//       setTimeout(() => {
//           v.dom.querySelector("input").select();
//       }, 50);
//       h.redraw();
//   });
// };
