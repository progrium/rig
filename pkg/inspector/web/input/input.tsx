// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Box} from "./box.tsx";

interface InputAttrs {
  value: string;
  readonly?: boolean;
  type?: string;
  step?: string;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Input extends ui.Element {
  onrender({attrs}: {attrs: InputAttrs}) {
    const type = attrs.type || "text";
    const value =  (attrs.value==="0") ? "0" : attrs.value;
    const readonly = attrs.readonly || false;
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    

    let inputAttrs = {
      type: type,
      readonly: (readonly)?"readonly":undefined,
      onchange: onchange,
      oninput: oninput,
      onfocus: onfocus,
      onfocusout: onfocusout,
      value: value,
    };

    if (type === "password") {
      inputAttrs["autocomplete"] = "password";
    }

    if (attrs.step) {
      inputAttrs["step"] = attrs.step;
    }

    const style = {
      flex: "1 1 auto", 
      background: "transparent", 
      outline: "none", 
      border: "0",
      maxWidth: "100%", 
      minWidth: "20%",
      // width: "0px" 
    }
    inputAttrs["style"] = style;

    return (
      <Box><input {...inputAttrs} /></Box>
    )
  }
}

export function DemoText() {
  return <Input value={"Hello world"} />
}
