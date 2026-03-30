// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as styling from "/-/inspector/util/styling.ts";
// @ts-ignore
import * as expand from "/-/inspector/util/expand.js";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";
// @ts-ignore
import * as molecule from "/-/inspector/elem/molecule.tsx";

interface NestedAttrs {
  label: string;
  expanded: boolean;
  icon?: string;
  extra?: any;
  pre?: any;
  ongrip?: (e) => void;
}

interface NestedState {
  expanded: boolean;
}

export class Nested extends ui.Element {
  oncreate(v) {
    expand.initExpanded(v, "label");
  }

  onrender({attrs, state, children}: {attrs: NestedAttrs, state: NestedState, children: any[]}) {
    // split up camel case labels into separate words, if no dots are found
    const label = (attrs.label.includes(".")) ? attrs.label : attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g).join(" ");
    const ongrip = attrs.ongrip || undefined;
    const icon = attrs.icon || undefined;
    const pre = attrs.pre || undefined;
    const extra = attrs.extra || undefined;
    const expanded = expand.isExpanded(state, attrs);

    const toggle = expand.toggler(state);

    const style = styling.from({
      "-webkit-user-select": "none",
    });

    const grip = (e) => {
      if (ongrip) {
        ongrip(e);
      }
    };

    return (
      <atom.VStack {...style.attrs()}>
        <molecule.Expander expanded={expanded}  
                           onclick={toggle}>
          {pre &&
            <atom.HStack>
              {pre}
            </atom.HStack>}
          <atom.HStack title={label} align="center" style={{flexGrow: "1"}}>
            {icon && <atom.Icon style={{marginRight: ".5rem", marginLeft: ".25rem"}} name={icon} />}
            <atom.Label style={{flexGrow: "1"}} onclick={toggle} text={label} />
            {ongrip && <atom.Dots style={{flexShrink: "1"}} rows={3} />}
          </atom.HStack>
          {extra && 
            <atom.HStack>
              {extra}
            </atom.HStack>}
        </molecule.Expander>
        <div style={{marginLeft: ".75rem", marginBottom: "0.50rem"}}>
          {expanded && children}
        </div>
      </atom.VStack>
    )
  }
}
