// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Input} from "./input.tsx";

interface NumberAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Number extends ui.Element {
  onrender({attrs}: {attrs: NumberAttrs}) {
    const oldfocus = attrs.onfocus;
    attrs.onfocus = (e) => {
      document.onwheel = (ee) => {
          let sign = (ee.deltaY > 0) ? 1 : -1;
          let inc = (sign*10) + (sign*0.01); // TODO: use step?
          e.target.value = e.target.valueAsNumber + inc;
      }
      oldfocus(e);
    };
    return (
      <Input type="number" {...attrs} />
    )
  }
}

export function DemoNumber() {
  return <Number value={"100"} />
}

/*
input[type=number]::-webkit-inner-spin-button {
  margin-top: -4px;
  -webkit-appearance: none;
}

*/
