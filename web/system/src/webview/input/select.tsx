import m from "mithril";
import { Box } from "./box";
import * as atom from "../elem/atom/mod";

interface SelectAttrs {
  value: string;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Select implements m.ClassComponent<SelectAttrs> {
  view({ attrs, children }: m.CVnode<SelectAttrs>) {
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

    const selectAttrs = {
      style: style,
      onchange: onchange,
      oninput: oninput,
      onfocus: onfocus,
      onfocusout: onfocusout,
    }


    // convert value to selected option
    // keep the old behavior: mark the selected <option> by mutating vnode attrs
    const kids = (children || []) as any[];
    const childrenSel = kids.map((el) => {
      let v = el?.text;
      if (el?.attrs && el.attrs.value !== undefined) v = el.attrs.value;
      if (v == value.toString()) {
        el.attrs = el.attrs || {};
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
