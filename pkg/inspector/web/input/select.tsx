// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Box} from "./box.tsx";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface SelectAttrs {
  value: string;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Select extends ui.Element {
  onrender({attrs, children}: {attrs: SelectAttrs, children: any[]}) {
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    const value = attrs.value || "";
    
    const style = {
      outline: "none",
      background: "transparent", 
      width: "100%", 
      border: "0",
      "-webkit-appearance": "none"
    }

    let selectAttrs = {
      style: style,
      onchange: onchange,
      oninput: oninput,
      onfocus: onfocus,
      onfocusout: onfocusout,
    }


    // convert value to selected option
    let childrenSel = children.map((el) => {
      let v = el.text;
      if (el.attrs && el.attrs.value !== undefined) {
        v = el.attrs.value;
      }
      if (v == value.toString()) {
        if (!el.attrs) {
            el.attrs = {};
        }
        el.attrs.selected = true;
      }
      return el;
    });

    return (
      <Box>
        <select {...selectAttrs}>
          {childrenSel}
        </select>
        <atom.Icon style={{marginRight:"3px", marginTop: "1px"}} name="fas fa-sort" />
      </Box>
    )
  }
}

export function DemoSelect() {
  return (
    <Select value="Bar">
      <option>Foo</option>
      <option>Bar</option>
      <option>Baz</option>
      <option>Qux</option>
    </Select>
  )
}
