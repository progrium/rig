import m from "mithril";
import * as atom from "../elem/atom/mod";
import { Box } from "./box";

interface ColorAttrs {
  value?: string;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

interface ColorState {
  color?: string;
}

export class Color implements m.ClassComponent<ColorAttrs> {
  view(vnode: m.CVnode<ColorAttrs>) {
    const { attrs } = vnode;
    const state = vnode.state as ColorState;
    const color = attrs.value || "#ffffff";
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
  
    
    state.color = state.color ?? color;
    // if (buffer) {
    //     state.color = buffer.value;
    // }

    const getRoot = () => (vnode as any).dom as Element | undefined;

    const select = () => {
      getRoot()?.querySelector<HTMLInputElement>("input[type=color]")?.click();
    };
    const updateFromPicker = (e: any) => {
      state.color =
        getRoot()?.querySelector<HTMLInputElement>("input[type=color]")?.value ??
        state.color ??
        color;
      if (onchange) onchange(e);
    };
    const updateFromTextbox = (e: any) => {
      state.color =
        getRoot()?.querySelector<HTMLInputElement>("input[type=text]")?.value ??
        state.color ??
        color;
      if (oninput) oninput(e);
    };
  
    const inputAttrs = {
      onchange: onchange,
      oninput: updateFromTextbox,
      onfocus: onfocus,
      onfocusout: onfocusout,
      value: state.color,
      style: {
        outline: "none", 
        background: "transparent",
        width: "100%",
        border: "0",
      }
    };

    const wellStyle = {
      width: "2.25rem",
      height: "100%",
      boxShadow: "inset 1px 1px 3px #111",
      marginLeft: "-0.5rem",
      marginRight: "0.5rem",
      backgroundColor: state.color,
    };

    return (
      <Box>
          <div style={wellStyle} onclick={select}>&nbsp;</div>
          <input type="color" onchange={updateFromPicker} value={state.color} style={{
            width: "0px", 
            height: "0px",
            visibility: "hidden",
          }} />
          <input type="text" {...inputAttrs} />
          <atom.Icon onclick={select} name="fas fa-eye-dropper" />
      </Box>
    )
  }
}

export function DemoColor() {
  return <Color value={"#0f0"} />
}

