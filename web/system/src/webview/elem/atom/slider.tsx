interface SliderAttrs {
  value: Number;
  min?: Number;
  max?: Number;
  step?: Number;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

import m from "mithril";

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

    return (
      <input type="range"
             onchange={onchange}
             oninput={oninput}
             onfocus={onfocus}
             onfocusout={onfocusout}
             style={{
              background: "transparent",
             }}
             min={min}
             max={max}
             value={value}
             step={step} />
    );
  }
}

export function DemoSlider() {
  return <Slider value={25} />
}


/*


input[type=range] {
  -webkit-appearance: none;
}
input[type=range]::-webkit-slider-runnable-track {
  width: 100%;
  height: 5px;
  cursor: pointer;
  background: #404040;
  border-radius: 2px;
}
input[type=range]::-moz-range-track {
  width: 100%;
  height: 5px;
  cursor: pointer;
  background: #404040;
  border-radius: 1px;
}
input[type=range]::-webkit-slider-thumb {
  height: 18px;
  width: 18px;
  border-radius: 25px;
  background: white;
  border: 1px solid black;
  cursor: pointer;
  -webkit-appearance: none;
  margin-top: -7px;
}
input[type=range]::-moz-range-thumb {
  height: 18px;
  width: 18px;
  border-radius: 25px;
  background: white;
  cursor: pointer;
}

*/