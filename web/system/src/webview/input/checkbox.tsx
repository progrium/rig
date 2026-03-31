import m from "mithril";
import * as atom from "../elem/atom/mod";

interface CheckboxAttrs {
  value?: boolean;
  label?: string;
  class?: string;
  stateless?: boolean;
  onchange?: (e: any) => void;
  onclick?: (e: any) => void;
}

interface CheckboxState {
  checked?: boolean;
}

// TODO: consider not being under input?
export class Checkbox implements m.ClassComponent<CheckboxAttrs> {
  view(vnode: m.CVnode<CheckboxAttrs>) {
    const { attrs } = vnode;
    const state = vnode.state as CheckboxState;
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
      onclick = (e: any) => {
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

