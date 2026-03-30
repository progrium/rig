// @ts-ignore
import * as ui from "/-/inspector/ui.js";

interface KnobAttrs {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  sensitivity?: number;
  stateless?: boolean;
  highlight?: string;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

interface KnobState {
  color: string;
  value: number;
}

export class Knob extends ui.Element {
  onrender({attrs, state}: {attrs: KnobAttrs, state: KnobState}) {
    const value = attrs.value || 0;
    const min = attrs.min || 0;
    const max = attrs.max || 100;
    const step = attrs.step || 1;
    const sensitivity = attrs.sensitivity || 1;
    const stateless = attrs.stateless || false;
    const highlight = attrs.highlight || "white";
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;

    state.color = (state.color === undefined)? "white" : state.color;
    state.value = (state.value === undefined)? value : state.value;
    if (stateless) {
      state.value = value;
    }
    const rot = (state.value * 300 / (max - min)) + 30;

    const onmousedown = (e) => {
      let lastX = e.pageX;
      let lastY = e.pageY;

      state.color = highlight;

      const onmousemove = (e) => {
        const offsetX = e.pageX - lastX;
        const offsetY = e.pageY - lastY;
        const diff = (offsetX - offsetY) * sensitivity;
        state.value = Math.min(max, Math.max(min, state.value + diff));
        ui.redraw();
        lastX = e.pageX;
        lastY = e.pageY;
      };

      const onmouseup = (e) => {
        state.color = "white";
        ui.redraw();

        document.body.removeEventListener("mouseup", onmouseup);
        document.body.removeEventListener("mousemove", onmousemove);
        return false;
      }

      document.body.addEventListener("mouseup", onmouseup);
      document.body.addEventListener("mousemove", onmousemove);

      return false;
    }
    
    const style = {
      width: "1.5rem",
      height: "1.5rem",
      background: "#484848",
      borderRadius: "50%",
      position: "relative",
      top: "0.25rem",
      boxShadow: "inset -1px -2px 2px #222",
      border: "1px solid #050505",
    }

    const indicator = {
      height: "0.8rem",
      width: "0.2rem",
      position: "relative",
      left: "50%",
      top: "50%",
      marginLeft: "-2px",
      borderRadius: "2px",
      transform: `rotate(${rot}deg)`,
      transformOrigin: "50% 0%",
      backgroundColor: state.color,
    }

    return (
      <div onmousedown={onmousedown} style={style}>
        <div style={indicator}></div>
        <input type="range" 
               style={{display: "none"}} 
               onchange={onchange} 
               oninput={oninput}
               onfocus={onfocus}
               onfocusout={onfocusout}
               min={min} 
               max={max} 
               value={state.value} 
               step={step} />
      </div>
    )
  }
}

export function DemoKnobAtom() {
  return <Knob value={0} highlight="yellow" />
}

