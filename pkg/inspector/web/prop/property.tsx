// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface PropertyAttrs {
  label: string;
  labelFmt?: (label: string) => any;
  ondragvalue?: (v: any) => any;
}

export class Property extends ui.Element {
  onrender({attrs, children}: {attrs: PropertyAttrs, children: any[]}) {
    // split up camel case labels into separate words
    const label = attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g).join(" ");
    const labelFmt = attrs.labelFmt || ((l) => l);
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
      // flexGrow: "1",
      width: "65%",
    };

    const startDrag = (e) => {
      let origin = e.pageX
      if (!ondragvalue) {
        return
      }
      const inputs = document.querySelectorAll("input")
      Array.from(inputs).forEach(input => {
        input["origPointerEvents"] = input.style.pointerEvents
        input.style.pointerEvents = "none"
      })
      const emit = (e) => ondragvalue(e.pageX-origin)
      document.addEventListener('mousemove', emit)
      document.addEventListener('mouseup', (e) => {
        Array.from(inputs).forEach(input => {
          input.style.pointerEvents = input["origPointerEvents"]
        })
        emit(e)
        document.removeEventListener('mousemove', emit)
      }, {once: true})
    }

    return (
      <div>
      <atom.HStack>
        <div title={label} style={left} onmousedown={startDrag}>{labelFmt(label)}</div>
        <div style={right}>{children}</div>
      </atom.HStack>
      </div>
    )
  }
}

export function DemoProperty() {
  return <Property label="Foo">Bar</Property>
}

export function DemoProperty_Alt() {
  return <Property label="Baz">Qux</Property>
}
