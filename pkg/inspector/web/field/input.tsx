// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as field from "./field.ts";
// @ts-ignore
import * as input from "/-/inspector/input/mod.ts";


// const committer = (v, field) => {
//   cmd.Exec("object.value.set", {Type: field.Type, Path: field.Path, Value: v });
// };

function fieldToBufferType(name: string): string {
  return {
    "String": "string",
    "Integer": "number",
    "Float": "number",
    //"Password": "password",
    //"Color": "color",
    "Time": "time",
    //"Date": "date",
    "Boolean": "boolean",
    "Pointer": "pointer",
    "Interface": "pointer"
  }[name]
}

interface InputAttrs {
  field: field.Field;
  value: any;
  committer: input.Committer;
}

interface InputState {
  buffer: input.ValueBuffer;
}

export class Input extends ui.Element {
  onrender({attrs, state}: {attrs: InputAttrs, state: InputState}) {
    const field = attrs.field;
    const committer = attrs.committer || undefined;
    const fieldType = fieldToBufferType(field.TypeName);
    let value = attrs.value || field.Value;

    state.buffer = state.buffer || new input.ValueBuffer(value, field, fieldType, committer, true, (field.TypeName==="Float")?0:500);
    state.buffer.set(value, field);

    if (field.Enum) {
      const selected = (opt, value) => {
        if (opt !== value) return {};
        return {"selected": "selected"};
      };
      let options = field.Enum.map((opt) => <option {...selected(opt, state.buffer.value)}>{opt}</option>)
      if (fieldType === "number") {
        options = field.Enum.map((opt, idx) => <option value={(field.Min||0) + idx} {...selected((field.Min||0) + idx, state.buffer.value)}>{opt}</option>);
      }
      return (
        <input.Select {...state.buffer.attrs}>
          {state.buffer.value === "" && <option></option>}
          {options}
        </input.Select>
      )
    }

    switch (fieldType) {
    case "pointer":
      if (field.Annots.Obj) {
        value = field.Annots.Obj.Name;
      }
      return <input.Reference readonly={true} value={value} oncontextmenu={(e) => {
        if (!field.Annots["pkgpath"]) {
          return;
        }
        window.parent.postMessage({menu: "menu-ref", params: {
          iface: `${field.Annots["pkgpath"]}.${field.ElemType}`, 
          fieldID: field.ID, 
          clientX: e.clientX, 
          clientY: e.clientY
        }}, "*");
      }} />
    case "string":
      return <input.Text {...state.buffer.attrs} />
    case "password":
      return <input.Password {...state.buffer.attrs} />
    case "number":
      let attrs = state.buffer.attrs
      if (field.TypeName==="Float") {
        attrs["step"] = "0.01"
      }
      return <input.Number {...attrs} />
    case "color":
      return <input.Color {...state.buffer.attrs} />
    case "time":
      return <input.Time {...state.buffer.attrs} />
    case "date":
      return <input.Date {...state.buffer.attrs} />
    case "boolean":
      return <input.Checkbox stateless={true} value={state.buffer.value} onchange={(e) => {
        console.log(e, state.buffer.value)
        state.buffer.change(e, true)
      }} />
    default:
      console.log(fieldType)
      return <input.Box>TODO: {fieldType}</input.Box>
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