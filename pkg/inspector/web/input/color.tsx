// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";
// @ts-ignore
import {Box} from "./box.tsx";

interface ColorAttrs {
  value: string;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

interface ColorState {
  color: string;
}

export class Color extends ui.Element {
  onrender({attrs, state, vnode}: {attrs: ColorAttrs, state: ColorState, vnode: any}) {
    const color = attrs.value || "#ffffff";
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
  
    
    state.color = (state.color === undefined) ? color : state.color;
    // if (buffer) {
    //     state.color = buffer.value;
    // }

    const select = (e) => {
      vnode.dom.querySelector("input[type=color]").click();
    }
    const updateFromPicker = (e) => {
        state.color = vnode.dom.querySelector("input[type=color]").value;
        if (onchange) onchange(e);
    }
    const updateFromTextbox = (e) => {
      state.color = vnode.dom.querySelector("input[type=text]").value;
      if (oninput) oninput(e);
    }
  
    let inputAttrs = {
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

