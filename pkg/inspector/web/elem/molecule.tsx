
// @ts-ignore
import * as atom from "./atom/mod.ts";

interface ExpanderAttrs {
  expanded: boolean;
  selected: boolean;
  noexpand: boolean;
  highlight: boolean;
  onclick?: (e) => void;
}

interface ExpanderState {
  hover: boolean;
}

// css classes: Expander, highlight.hover, highlight.selected
export const Expander = {
  passthrough: () => ["class", "style"],
  view: ({attrs, children}) => {
    var expanded = attrs.expanded;
    var noexpand = attrs.noexpand;
    var onclick = attrs.onclick;

    const icon = `fas fa-caret-${(expanded) ? 'down' : 'right'}`
    return (
      <atom.HStack class="Expander" align="center">
        {!noexpand && <atom.Icon name={icon} {...{onclick}} el-style={{
          marginLeft: "-1rem",
          position: "absolute"
        }}/>}
        {children}
      </atom.HStack>
    )
  }
}
