import m from "mithril";

interface BoxAttrs {
  noborder?: boolean;
  transparent?: boolean;
  onmouseout?: (e: any) => void;
  onmouseover?: (e: any) => void;
}

export class Box implements m.ClassComponent<BoxAttrs> {
  view({ attrs, children }: m.CVnode<BoxAttrs>) {
    const noborder = attrs.noborder || false;
    const transparent = attrs.transparent || false;
    const onmouseout = attrs.onmouseout || undefined;
    const onmouseover = attrs.onmouseover || undefined;

    const style = {
      height: "1.25rem",
      width: "100%",
    };
    if (transparent) {
      (style as any).backgroundColor = "transparent";
    }
  
    const inner = {
      height: "1.25rem",
      // backgroundColor: "#222",
      lineHeight: "1.25rem",
      display: "flex",
      paddingLeft: "0.25rem",
      paddingRight: "0.25rem",
    };
    if (transparent) {
      (inner as any).backgroundColor = "transparent";
    }
    if (!noborder) {
      (inner as any).boxShadow = "inset 1px 1px 3px var(--box-shadow)";
      (inner as any).border = "1px solid var(--box-border)";
    }

    return (
      <div style={style}>
        <form style={inner}
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