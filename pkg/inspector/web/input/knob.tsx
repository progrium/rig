// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Box} from "./box.tsx";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface KnobAttrs {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  sensitivity?: number;
  readonly?: boolean;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

// TODO: consider not being under input? 
export class Knob extends ui.Element {
  onrender({attrs}: {attrs: KnobAttrs}) {
    const value = attrs.value || 0;
    const min = attrs.min || 0;
    const max = attrs.max || 100;
    const step = attrs.step || 1;
    const sensitivity = attrs.sensitivity || 1;
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    const readonly = attrs.readonly || false;

    let inputAttrs = {
      readonly: (readonly)?"readonly":undefined,
      onchange: onchange,
      oninput: oninput,
      onfocus: onfocus,
      onfocusout: onfocusout,
      value: value,
      min: min,
      max: max,
      step: step,
      sensitivity: sensitivity
    };

    return (
      <Box transparent noborder>
        <atom.Knob {...inputAttrs} />
      </Box>
    )
  }
}

export function DemoKnob() {
  return <Knob value={50} />
}

