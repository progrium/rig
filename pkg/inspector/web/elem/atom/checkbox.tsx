// @ts-ignore
import * as ui from "/-/inspector/ui.js";

interface CheckboxAttrs {
  label: string;
  checked: boolean;
  stateless: boolean;
  onfocus: (e) => void;
  onfocusout: (e) => void;
  oninput: (e) => void;
  onchange: (e) => void;
  onclick: (e) => void;
}

interface CheckboxState {
  checked: boolean;
}

export class Checkbox extends ui.Element {
  onrender({attrs, state}: {attrs: CheckboxAttrs, state: CheckboxState}) {
    const label = attrs.label || "";
    const checked = attrs.checked || false;
    const stateless = attrs.stateless || false;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    const onchange = attrs.onchange || undefined;
    const onclick = attrs.onclick || (() => {
        state.checked = (state.checked) ? false : true;
    });

    state.checked = state.checked || checked;
    if (stateless) {
        state.checked = checked;
    }

    const outer = {
      marginTop: "auto",
      marginBottom: "auto",
      display: "flex"
    };
    const control = {
      // "-webkit-appearance": "none",
      boxShadow: "inset 1px 1px 3px var(--box-border)",
      border: "1px solid var(--box-border)",  
      width: "1rem",
      height: "1rem",
      position: "relative"
    };

    return (
      <div style={outer}>
        <div style={{position: "relative"}}>
          <input type="checkbox"
                style={control}
                onclick={onclick}
                onchange={onchange}
                oninput={oninput}
                onfocus={onfocus}
                onfocusout={onfocusout}
                checked={(state.checked)?"checked":undefined} />
          {/*state.checked ?
            <div style={{
              position: "absolute", 
              pointerEvents: "none",
              top:"0", 
              bottom: "0", 
              lineHeight: "1rem", 
              right: "0", 
              left: "4px", 
            }}>{ui.trust("&check;")}</div>
          :null*/}
        </div>
        {label && <div onclick={onclick} style={{marginLeft: "0.5rem", lineHeight: "inherit"}}>{label}</div>}
      </div>
    )
  }
}

export function DemoCheckbox() {
  return <Checkbox checked={true} label="Checkbox" />
}

/*

input[type="checkbox"]:before, input[type="checkbox"]:checked:before {
  position:absolute;
  top:0px;
  left:1px;
  width:100%;
  height:100%;
  line-height:1rem;
  text-align:center;
  color:#fff;
  content: '';
 }
 input[type="checkbox"]:checked:before {
  content: '✔';
 }

*/