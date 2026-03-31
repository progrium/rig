import m from "mithril";

import type { Field } from "./field";
import { Input as FieldInput } from "./input";
import * as store from "./store";
import * as text from "../input/text";
import * as atom from "../elem/atom/mod";
import * as prop from "../prop/mod";
import * as collection from "../prop/collection";
import type { Committer } from "../input/buffer";

interface InspectorAttrs {
  field: Field;
  label?: string | false;
  expanded: boolean;
  store?: store.Store;
}

function throttle<T extends (...args: any[]) => void>(
  fn: T,
  threshold = 50,
  scope?: any
): (...args: Parameters<T>) => void {
  let last: number | undefined;
  let deferTimer: ReturnType<typeof setTimeout> | undefined;
  return function (this: any, ...args: Parameters<T>) {
    const context = scope ?? this;
    const now = +new Date();
    if (last !== undefined && now < last + threshold) {
      if (deferTimer !== undefined) clearTimeout(deferTimer);
      deferTimer = setTimeout(() => {
        last = now;
        fn.apply(context, args);
      }, threshold);
    } else {
      last = now;
      fn.apply(context, args);
    }
  };
}

export class Inspector implements m.ClassComponent<InspectorAttrs> {
  view(vnode: m.CVnode<InspectorAttrs>) {
    const { attrs } = vnode;
    const field = attrs.field;
    const rawLabel: string | false = attrs.label === undefined ? field.Name : attrs.label;
    const expanded = attrs.expanded;
    const subfields = field.Fields ?? [];
    const fieldstore = attrs.store ?? new store.Backend();

    const committer: Committer = (v: any, f: any, _buf) => {
      fieldstore.setValue(f, v);
    };

    const stringLabel = typeof rawLabel === "string" ? rawLabel : field.Name;

    if (rawLabel === "Component" && field.TypeName === "Struct") {
      return [];
    }

    switch (field.TypeName) {
      case "Struct":
        return (
          <prop.Nested key={vnode.key} expanded={expanded} label={stringLabel}>
            {subfields.map((f: Field) => (
              <Inspector key={f.Name} field={f} store={fieldstore} expanded={expanded} />
            ))}
          </prop.Nested>
        );
      case "Map":
        // console.log("map", field);
        return (
          <collection.Collection
            expanded={expanded}
            label={stringLabel}
            onadd={(e: any, el: NodeListOf<HTMLInputElement>) =>
              fieldstore.setKey(field, el[0].value, el[1].value)
            }
            input={[<text.Text style={{ width: "30%" }} />, <text.Text />]}
            renderer={(entry: [string, any], idx: number) => (
              <atom.HStack key={idx} style={{ marginBottom: "0.25rem", marginTop: "0.25rem" }}>
                <text.Text style={{ width: "30%" }} value={entry[0]} />
                <FieldInput
                  field={{
                    ID: `foo.${idx}`,
                    Name: "",
                    Flags: [],
                    TypeName: field.ElemType ?? "String",
                    Value: entry[1],
                  }}
                />
                <atom.Icon
                  name="fas fa-times-circle"
                  onclick={() => fieldstore.removeKey(field, entry[0])}
                  style={{
                    marginLeft: "0.5rem",
                    marginTop: "auto",
                    marginBottom: "auto",
                  }}
                />
              </atom.HStack>
            )}
            items={Object.entries((field.Value as object) ?? {})}
          />
        );
      case "Slice":
      case "Array": {
        const buf = new store.Buffer(field.ElemFields ?? []);
        return (
          <collection.Collection
            expanded={expanded}
            label={stringLabel}
            onadd={(e: any, el: NodeListOf<HTMLInputElement>) =>
              fieldstore.appendValue(
                field,
                field.ElemType === "Struct" ? buf.getValue() : el[0].value
              )
            }
            input={
              field.ElemType === "Struct" ? (
                <div>
                  {(field.ElemFields ?? []).map((f: Field) => (
                    <Inspector key={f.Name} field={f} store={buf} expanded={expanded} />
                  ))}
                </div>
              ) : (
                <text.Text />
              )
            }
            renderer={(item: any, idx: number) => (
              <atom.HStack key={idx} style={{ marginBottom: "0.25rem", marginTop: "0.25rem" }}>
                {field.ElemType === "Struct" ? (
                  <prop.Nested expanded={true} label={field.ElemType ?? ""}>
                    {(item.Fields as Field[] | undefined)?.map((f: Field) => (
                      <Inspector key={f.Name} field={f} expanded={expanded} />
                    ))}
                  </prop.Nested>
                ) : (
                  <FieldInput
                    field={{
                      ID: `foo.${idx}`,
                      Name: "",
                      Flags: [],
                      TypeName: field.ElemType ?? "String",
                      Value: item.Value,
                    }}
                  />
                )}
                <atom.Icon
                  name="fas fa-times-circle"
                  onclick={() => fieldstore.removeKey(field, String(idx))}
                  style={{
                    marginLeft: "0.5rem",
                    marginTop: "auto",
                    marginBottom: "auto",
                  }}
                />
              </atom.HStack>
            )}
            items={field.Fields ?? []}
          />
        );
      }
      default: {
        const style = {
          lineHeight: "1.25rem",
          marginTop: "0.25rem",
          marginBottom: "0.25rem",
          marginRight: "0.25rem",
        };
        if (rawLabel === false) {
          return (
            <FieldInput key={vnode.key} field={field} style={style} committer={committer} />
          );
        }
        return (
          <prop.Property
            key={vnode.key}
            label={stringLabel}
            style={style}
            {...(field.TypeName === "Float"
              ? {
                  ondragvalue: throttle((v: any) => {
                    fieldstore.setValue(field, field.Value + v);
                  }),
                }
              : {})}
          >
            <FieldInput field={field} committer={committer} />
          </prop.Property>
        );
      }
    }
  }
}
