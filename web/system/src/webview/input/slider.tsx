import m from "mithril";
import { Box } from "./box";
import * as atom from "../elem/atom/mod";

interface SliderAttrs {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  readonly?: boolean;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Slider implements m.ClassComponent<SliderAttrs> {
  view({ attrs }: m.CVnode<SliderAttrs>) {
    const value = attrs.value || 0;
    const min = attrs.min || 0;
    const max = attrs.max || 100;
    const step = attrs.step || 1;
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
    };

    return (
      <Box transparent noborder>
          <atom.Slider {...inputAttrs} />
      </Box>
    )
  }
}

export function DemoSlider() {
  return <Slider value={50} />
}

