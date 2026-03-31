
import m from "mithril";
import * as atom from "./atom/mod";

interface ExpanderAttrs {
  expanded: boolean;
  selected: boolean;
  noexpand: boolean;
  highlight: boolean;
  onclick?: (e: any) => void;
}

interface ExpanderState {
  hover: boolean;
}

// css classes: Expander, highlight.hover, highlight.selected
export class Expander implements m.ClassComponent<ExpanderAttrs> {
  view({ attrs, children }: m.CVnode<ExpanderAttrs>) {
    const expanded = attrs.expanded;
    const noexpand = attrs.noexpand;
    const onclick = attrs.onclick;

    const icon = `fas fa-caret-${expanded ? "down" : "right"}`;
    return (
      <atom.HStack class="Expander" align="center">
        {!noexpand && (
          <atom.Icon
            name={icon}
            {...{ onclick }}
            style={{
              marginLeft: "-1rem",
              position: "absolute",
            }}
          />
        )}
        {children}
      </atom.HStack>
    );
  }
}
