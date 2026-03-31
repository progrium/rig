import m from "mithril";
import type { Field } from "./field";
import * as input from "../input/mod";

function fieldToBufferType(name: string): input.Type | "pointer" | undefined {
  const map: Record<string, input.Type | "pointer"> = {
    String: "string",
    Integer: "number",
    Float: "number",
    Time: "time",
    Boolean: "boolean",
    Pointer: "pointer",
    Interface: "pointer",
  };
  return map[name];
}

function valueBufferType(fieldKind: input.Type | "pointer" | undefined): input.Type {
  if (fieldKind === undefined || fieldKind === "pointer") {
    return "string";
  }
  return fieldKind;
}

interface InputAttrs {
  field: Field;
  value?: any;
  committer?: input.Committer;
  style?: Record<string, any>;
}

export class Input implements m.ClassComponent<InputAttrs> {
  buffer?: input.ValueBuffer;

  view(vnode: m.CVnode<InputAttrs>) {
    const { attrs } = vnode;
    const state = vnode.state as Input;
    const field = attrs.field;
    const committer = attrs.committer || undefined;
    const fieldKind = fieldToBufferType(field.TypeName);
    let value = attrs.value ?? field.Value;

    const bufferType = valueBufferType(fieldKind);
    const commitFn: input.Committer =
      committer ?? ((_v: any, _extra: any, _buf: input.ValueBuffer) => {});
    state.buffer =
      state.buffer ||
      new input.ValueBuffer(
        value,
        field,
        bufferType,
        commitFn,
        true,
        field.TypeName === "Float" ? 0 : 500
      );
    state.buffer.set(value, field);

    const wrapStyle = (node: m.Children): m.Children =>
      attrs.style ? <div style={attrs.style}>{node}</div> : node;

    if (field.Enum) {
      const selected = (opt: any, val: any) => {
        if (opt !== val) return {};
        return { selected: "selected" };
      };
      let options = field.Enum.map((opt) => (
        <option {...selected(opt, state.buffer!.value)}>{opt}</option>
      ));
      if (fieldKind === "number") {
        options = field.Enum.map((opt, idx) => (
          <option
            value={(field.Min || 0) + idx}
            {...selected((field.Min || 0) + idx, state.buffer!.value)}
          >
            {opt}
          </option>
        ));
      }
      return wrapStyle(
        <input.Select {...state.buffer!.attrs}>
          {state.buffer!.value === "" && <option></option>}
          {options}
        </input.Select>
      );
    }

    switch (fieldKind) {
      case "pointer":
        if (field.Annots?.Obj) {
          value = field.Annots.Obj.Name;
        }
        return wrapStyle(
          <input.Reference
            readonly={true}
            value={value}
            oncontextmenu={(e: any) => {
              if (!field.Annots?.["pkgpath"]) {
                return;
              }
              window.parent.postMessage(
                {
                  menu: "menu-ref",
                  params: {
                    iface: `${field.Annots["pkgpath"]}.${field.ElemType}`,
                    fieldID: field.ID,
                    clientX: e.clientX,
                    clientY: e.clientY,
                  },
                },
                "*"
              );
            }}
          />
        );
      case "string":
        return wrapStyle(<input.Text {...state.buffer!.attrs} />);
      case "password":
        return wrapStyle(<input.Password {...state.buffer!.attrs} />);
      case "number": {
        const numAttrs: Record<string, any> = { ...state.buffer!.attrs };
        if (field.TypeName === "Float") {
          numAttrs.step = "0.01";
        }
        return wrapStyle(<input.Number {...numAttrs} />);
      }
      case "color":
        return wrapStyle(<input.Color {...state.buffer!.attrs} />);
      case "time":
        return wrapStyle(<input.Time {...state.buffer!.attrs} />);
      case "date":
        return wrapStyle(<input.Date {...state.buffer!.attrs} />);
      case "boolean":
        return wrapStyle(
          <input.Checkbox
            stateless={true}
            value={state.buffer!.value}
            onchange={(e: any) => {
              console.log(e, state.buffer!.value);
              state.buffer!.change(e, true);
            }}
          />
        );
      default:
        console.log(fieldKind);
        return wrapStyle(<input.Box>TODO: {String(fieldKind)}</input.Box>);
    }
  }
}

/*
// const bind = (attrs.bind!==undefined)?attrs.bind:true;


if (!field.Type.startsWith("reference:")) {
          console.log("unknown field type:", field.Type);
          return <pre>Unknown field type: {field.Type}</pre>;
      }

      let icon = undefined;
      if (field.Value) {
          let objPath = field.Value.match(/(.*)\//)[1];
          let obj = manifold.lookup(objPath);
          if (obj) {
              icon = obj.Icon;
          }
      }

      let refType = field.Type.split(":")[1];
      let context = {Type: field.Type, Path: field.Path, RefType: refType };

      const onunset = () => {
          cmd.Exec("object.value.set", {Type: field.Type, Path: field.Path, Value: null });
      }
      state.buffer.committer = (v, field) => {
          if (!v) return onunset();
          cmd.Exec("object.value.set", {Type: field.Type, Path: field.Path, RefValue: `${v}/${refType}` });
      }

      state.buffer.formatter = (v, field) => {
          // show only object base path + component
          let parts = v.split("/");
          if (parts[parts.length-1].endsWith(".Main")) {
              parts[parts.length-1] = "object.Main";
          }
          return parts.splice(parts.length-2,2).join("/");
      };
      
      return <input.ReferenceInput 
          bind={bind}
          contextMenu="component/reference" 
          icon={icon}
          context={context}
          placeholder={refType} 
          buffer={state.buffer}
          onunset={onunset}
          />   

*/
