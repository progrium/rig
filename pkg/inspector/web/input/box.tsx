// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as styling from "/-/inspector/util/styling.ts";

interface BoxAttrs {
  noborder?: boolean;
  transparent?: boolean;
  onmouseout?: (e) => void;
  onmouseover?: (e) => void;
}

export class Box extends ui.Element {
  onrender({attrs, children}: {attrs: BoxAttrs, children: any}) {
    const noborder = attrs.noborder || false;
    const transparent = attrs.transparent || false;
    const onmouseout = attrs.onmouseout || undefined;
    const onmouseover = attrs.onmouseover || undefined;

    const style = styling.from({
      height: "1.25rem",
      width: "100%",
    });
    style.add({background: "transparent"}, () => transparent);
  
    const inner = styling.from({
      height: "1.25rem",
      // backgroundColor: "#222",
      lineHeight: "1.25rem",
      display: "flex",
      paddingLeft: "0.25rem",
      paddingRight: "0.25rem",
    });
    inner.add({background: "transparent"}, () => transparent);
    inner.add({
        boxShadow: "inset 1px 1px 3px var(--box-shadow)",
        border: "1px solid var(--box-border)",
    }, () => !noborder);

    return (
      <div {...style.attrs()}>
        <form style={inner.style()}
              spellcheck=""
              onsubmit={() => false}              
              onmouseout={onmouseout}
              onmouseover={onmouseover}>
          {children}
        </form>
      </div>
    )
  }
}

export function DemoBox() {
  return <Box>Hello world</Box>
}

/*
.InputBox {
  display: flex;
  background-color: #404040;

  width: 100%;
  height: 1.75rem;
}
.InputBox form {
  display: flex;
  width: 100%;
  height: 1.75rem;

  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.InputBox input, .InputBox select {
  background: transparent;
  height: 1.75rem;
  margin-right: auto;
}
.InputBox i {
  margin-top: 0.4rem;
}
*/