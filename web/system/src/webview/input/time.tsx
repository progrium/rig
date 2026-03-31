import m from "mithril";
import { Box } from "./box";
import * as atom from "../elem/atom/mod";

interface TimeAttrs {
  value?: string;
  onchange?: (e: any, v: string) => void;
  onreset?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

interface TimeState {
  now?: string;
}

export class Time implements m.ClassComponent<TimeAttrs> {
  view(vnode: m.CVnode<TimeAttrs>) {
    const { attrs } = vnode;
    const state = vnode.state as TimeState;
    const onreset = attrs.onreset;
    const onchange = attrs.onchange;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    let value = attrs.value;
    
    state.now = state.now || undefined;

    if (value === "00:00") {
      value = undefined;
    }
    if (state.now) {
      value = state.now;
      state.now = undefined;
    }
    
    const onkeydown = (e: any) => {
      if (e.key === "Backspace" && e.shiftKey) {
        e.preventDefault();
        if (onreset) {
          onreset(e);
        }
        if (onchange) {
          onchange(e, "00:00");
        }
      }
    }
    const ondblclick = (e: any) => {
      let n = new Date();
      state.now = `${("0" + n.getHours()).slice(-2)}:${("0" + n.getMinutes()).slice(-2)}`;
      if (onchange) {
        onchange(e, state.now);
      }
    }
    
    return (
      <Box>
        <input type="time" 
            onkeydown={onkeydown} 
            onchange={onchange}
            oninput={oninput}
            onfocus={onfocus}
            onfocusout={onfocusout}
            style={{ 
              width: "100%",
              background: "transparent",
              outline: "none", 
              border: "0",
           }}
            placeholder="--:-- --"
            value={value} />
        <atom.Icon ondblclick={ondblclick} name="far fa-clock"></atom.Icon>
      </Box>
    )
  }
}

export function DemoTime() {
  return <Time />
}
