import m from "mithril";
import * as atom from "../elem/atom/mod";

interface PropertyAttrs {
  label: string;
  labelFmt?: (label: string) => any;
  ondragvalue?: (v: any) => any;
  style?: Record<string, any>;
}

export class Property implements m.ClassComponent<PropertyAttrs> {
  view(vnode: m.CVnode<PropertyAttrs>) {
    const { attrs, children } = vnode;
    const labelParts = attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g);
    const label = (labelParts || []).join(" ");
    const labelFmt = attrs.labelFmt || ((l: string) => l);
    const ondragvalue = attrs.ondragvalue;

    const left = {
      overflow: "hidden",
      textOverflow: "ellipsis",
      whiteSpace: "nowrap",
      marginRight: ".25rem",
      marginLeft: "0",
      width: "35%",
      cursor: "default",
    };
    const right = {
      display: "flex",
      width: "65%",
    };

    const startDrag = (e: any) => {
      const origin = e.pageX;
      if (!ondragvalue) {
        return;
      }
      const inputs = document.querySelectorAll("input");
      Array.from(inputs).forEach((input) => {
        (input as HTMLInputElement & { origPointerEvents?: string }).origPointerEvents =
          input.style.pointerEvents;
        input.style.pointerEvents = "none";
      });
      const emit = (me: any) => ondragvalue(me.pageX - origin);
      document.addEventListener("mousemove", emit);
      document.addEventListener(
        "mouseup",
        (me: any) => {
          Array.from(inputs).forEach((input) => {
            const inp = input as HTMLInputElement & { origPointerEvents?: string };
            input.style.pointerEvents = inp.origPointerEvents ?? "";
          });
          emit(me);
          document.removeEventListener("mousemove", emit);
        },
        { once: true }
      );
    };

    return (
      <div style={attrs.style}>
        <atom.HStack>
          <div title={label} style={left} onmousedown={startDrag}>
            {labelFmt(label)}
          </div>
          <div style={right}>{children}</div>
        </atom.HStack>
      </div>
    );
  }
}

export function DemoProperty() {
  return <Property label="Foo">Bar</Property>;
}

export function DemoProperty_Alt() {
  return <Property label="Baz">Qux</Property>;
}
