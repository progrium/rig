import m from "mithril";
import { Input } from "./input";

interface NumberAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Number implements m.ClassComponent<NumberAttrs> {
  view({ attrs }: m.CVnode<NumberAttrs>) {
    const oldfocus = attrs.onfocus;
    const onfocus = (e: any) => {
      document.onwheel = (ee) => {
          let sign = (ee.deltaY > 0) ? 1 : -1;
          let inc = (sign*10) + (sign*0.01); // TODO: use step?
          e.target.value = e.target.valueAsNumber + inc;
      }
      if (oldfocus) oldfocus(e);
    };
    return (
      <Input type="number" {...attrs} onfocus={onfocus} />
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
