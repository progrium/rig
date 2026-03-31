import m from "mithril";
import { Box } from "./box";
import * as atom from "../elem/atom/mod";

interface KnobAttrs {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  sensitivity?: number;
  readonly?: boolean;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

// TODO: consider not being under input?
export class Knob implements m.ClassComponent<KnobAttrs> {
  view({ attrs }: m.CVnode<KnobAttrs>) {
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

    const inputAttrs = {
      readonly,
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

