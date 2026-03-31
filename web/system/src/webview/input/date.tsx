import m from "mithril";
import { Box } from "./box";
import * as atom from "../elem/atom/mod";

interface DateAttrs {
  value?: string;
  onchange?: (e: any, v: string) => void;
  onreset?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Date implements m.ClassComponent<DateAttrs> {
  view({ attrs }: m.CVnode<DateAttrs>) {
    let value = attrs.value || undefined;
    const onreset = attrs.onreset || undefined;
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;

    const onchangeWrap = (e: any) => {
      if (!onchange) return;
      onchange(e, e?.target?.value);
    };
    
    const onkeydown = (e: any) => {
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
               onchange={onchangeWrap}
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

