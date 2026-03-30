// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Box} from "./box.tsx";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface DateAttrs {
  value: string;
  onchange?: (e, v) => void;
  onreset?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Date extends ui.Element {
  onrender({attrs}: {attrs: DateAttrs}) {
    let value = attrs.value || undefined;
    const onreset = attrs.onreset || undefined;
    const onchange = attrs.onchange;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    
    const onkeydown = (e) => {
      if (e.key === "Backspace" && e.shiftKey) {
        e.preventDefault();
        if (onreset) {
          onreset(e);
        }
        if (onchange) {
          onchange(e, "0001-01-01");
        }
      }
    };

    if (value === "0001-01-01") {
      value = undefined;
    }

    return (
      <Box>
        <input type="text"
               onkeydown={onkeydown}
               placeholder="mm/dd/yyyy"
               style={{ 
                  width: "100%",
                  background: "transparent",
                  outline: "none", 
                  border: "0",
               }}
               onchange={onchange}
               oninput={oninput}
               onfocus={onfocus}
               onfocusout={onfocusout}
               value={value} />
        <atom.Icon name="fas fa-calendar-day"></atom.Icon>
      </Box>
    )
  }
}

export function DemoDate() {
  return <Date />
}

