// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as field from "./field.ts";
// @ts-ignore
import * as input from "./input.tsx";
// @ts-ignore
import * as store from "./store.ts";
// @ts-ignore
import * as text from "/-/inspector/input/text.tsx";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";
// @ts-ignore
import * as prop from "/-/inspector/prop/mod.ts";
// @ts-ignore
import * as collection from "/-/inspector/prop/collection.tsx";



interface InspectorAttrs {
  field: field.Field;
  label?: string|false;
  expanded: boolean;
  store?: store.Store;
  // placeholder
  // nolabel
}

export class Inspector extends ui.Element {
  onrender({attrs, vnode}: {attrs: InspectorAttrs, vnode: any}) {
    const field = attrs.field || {};
    const label = attrs.label || field.Name;
    const expanded = attrs.expanded;
    const subfields = field.Fields || [];
    const fieldstore = attrs.store || new store.Backend();

    const committer = (v, field) => fieldstore.setValue(field, v)

    if (label === "Component" && field.TypeName === "Struct") {
      return [];
    }
    switch (field.TypeName) {
    case "Struct":
      return (
        <prop.Nested key={vnode.key} expanded={expanded} label={label}>
          {subfields.map((f) => <Inspector key={f.Name} field={f} store={fieldstore} />)}
        </prop.Nested>
      )
    // case "checkboxes":
    //   return (
    //     <Checkboxes 
    //       key={vnode.key} 
    //       expanded={expanded} 
    //       label={label} 
    //       field={field} />
    //   )
    case "Map":
      console.log("map", field)
      return <collection.Collection 
                expanded={expanded} 
                label={label} 
                onadd={(e, el) => fieldstore.setKey(field, el[0].value, el[1].value)}
                input={[<text.Text style={{width: "30%"}} />, <text.Text />]}
                renderer={(entry,idx) => 
                  <atom.HStack key={idx} style={{marginBottom: "0.25rem", marginTop: "0.25rem"}}>
                    <text.Text style={{width: "30%"}} value={entry[0]} />
                    <input.Input field={{ID: `foo.${idx}`, TypeName: field.ElemType, Value: entry[1]}} />
                    <atom.Icon name="fas fa-times-circle" 
                      onclick={() => fieldstore.removeKey(field, entry[0])}
                      style={{
                        marginLeft: "0.5rem", 
                        marginTop: "auto", 
                        marginBottom: "auto"
                      }} />
                  </atom.HStack>}
                items={Object.entries(field.Value||[])} />
    case "Slice":
    case "Array":
      const buf = new store.Buffer(field.ElemFields)
      return <collection.Collection 
                expanded={expanded} 
                label={label} 
                onadd={(e, el) => fieldstore.appendValue(field, (field.ElemType==="Struct")?buf.getValue():el[0].value)}
                input={(field.ElemType==="Struct")?<div>{field.ElemFields.map((f) => <Inspector key={f.Name} field={f} store={buf} />)}</div>:<text.Text />}
                renderer={(item,idx) => 
                  <atom.HStack key={idx} style={{marginBottom: "0.25rem", marginTop: "0.25rem"}}>
                    {(field.ElemType==="Struct")
                      ?<prop.Nested expanded={true} label={field.ElemType}>
                        {item.Fields.map((f) => <Inspector key={f.Name} field={f} />)}
                      </prop.Nested>
                      :<input.Input field={{ID: `foo.${idx}`, TypeName: field.ElemType, Value: item.Value}} />}
                    <atom.Icon name="fas fa-times-circle" 
                      onclick={() => fieldstore.removeKey(field, idx)}
                      style={{
                        marginLeft: "0.5rem", 
                        marginTop: "auto", 
                        marginBottom: "auto"
                      }} />
                  </atom.HStack>}
                items={field.Fields||[]} />
    default:
      const style = {
        lineHeight: "1.25rem",
        marginTop: "0.25rem", 
        marginBottom: "0.25rem",
        marginRight: "0.25rem",
      }
      if (label === false) {
        return <input.Input key={vnode.key} field={field} style={style} committer={committer} />
      }
      return (
        <prop.Property key={vnode.key} label={label} style={style} {...(field.TypeName==="Float")?{ondragvalue: throttle((v) => {
          committer(field.Value+v, field)
        })}:{}}>
          <input.Input field={field} committer={committer} />
        </prop.Property>
      )
    }
  }
}


function throttle(fn, threshhold=undefined, scope=undefined) {
  threshhold || (threshhold = 50);
  var last,
      deferTimer;
  return function () {
    var context = scope || this;

    var now = +new Date,
        args = arguments;
    if (last && now < last + threshhold) {
      // hold on to it
      clearTimeout(deferTimer);
      deferTimer = setTimeout(function () {
        last = now;
        fn.apply(context, args);
      }, threshhold);
    } else {
      last = now;
      fn.apply(context, args);
    }
  };
}