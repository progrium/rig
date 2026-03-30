// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface CheckboxAttrs {
  value: boolean;
  label: string;
  class: string;
  stateless: boolean;
  onchange?: (e) => void;
  onclick?: (e) => void;
}

interface CheckboxState {
  checked: boolean;
}

// TODO: consider not being under input? 
export class Checkbox extends ui.Element {
  onrender({attrs, state, vnode}: {attrs: CheckboxAttrs, state: CheckboxState, vnode: any}) {
    const label = attrs.label || "";
    const value = attrs.value || false;
    const class_ = attrs.class || "";
    const stateless = attrs.stateless || false;
    const onchange = attrs.onchange || undefined;
    let onclick = attrs.onclick || undefined;
    
    state.checked = (state.checked === undefined) ? value : state.checked;
    if (stateless) {
        state.checked = value;
    } else {
      onclick = (e) => {
        state.checked = (state.checked) ? false : true;
      }
    }
    

    return (
      <atom.Checkbox
          onclick={onclick}
          label={label}
          onchange={onchange}
          class={class_}
          stateless={true}
          checked={state.checked}
        />
    )
  }
}

export function DemoCheckbox() {
  return <Checkbox value={true} label="Label" />
}

