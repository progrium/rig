// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as styling from "/-/inspector/util/styling.ts";
// @ts-ignore
import * as expand from "/-/inspector/util/expand.js";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";
// @ts-ignore
import * as molecule from "/-/inspector/elem/molecule.tsx";

interface CollectionAttrs {
  label: string;
  expanded: boolean;
  items: [];
  input: any;
  renderer: (item: any, idx: number) => any;
  onadd: (e: any, el: HTMLElement) => void;
}

interface CollectionState {
  expanded: boolean
  adding: boolean
}

export class Collection extends ui.Element {
  oncreate(v) {
    expand.initExpanded(v, "label");
  }

  onrender({attrs, state, children}: {attrs: CollectionAttrs, state: CollectionState, children: any[]}) {
    // split up camel case labels into separate words
    const label = attrs.label.match(/[A-Z]+[^A-Z]*|[^A-Z]+/g).join(" ")
    const items = attrs.items || []
    const input = attrs.input || <input type="text" />
    const onadd = attrs.onadd || undefined
    const renderer = attrs.renderer || ((item) => <div>{item}</div>)
    const expanded = expand.isExpanded(state, attrs)

    let addingInit = (children.length === 0 && expanded) ? true : false;
    state.adding = (state.adding === undefined) ? addingInit: state.adding;

    const toggle = expand.toggler(state);
    const toggleAdd = (e) => {
      state.adding = (state.adding) ? false : true
    };

    const style = styling.from({
      "-webkit-user-select": "none",
    });

    const add = (e) => {
      onadd(e, e.target.closest(".HStack").querySelectorAll("input"))
      state.adding = false
    }

    return (
      <atom.VStack {...style.attrs()}>
        <molecule.Expander expanded={expanded}  
                           onclick={toggle}>
          <atom.Label style={{flexGrow: "1"}} onclick={toggle} text={label} />
          <span style={{color:"#888"}}>{items.length} items</span>
          <atom.Icon name="fas fa-plus-circle" onclick={toggleAdd} style={{marginLeft: "0.5rem"}} />
        </molecule.Expander>
        {expanded ? 
          <div style={{marginLeft: "1.5rem"}}>
            {state.adding ?
              <atom.HStack>
                <atom.HStack style={{flexGrow: "1"}}>{input}</atom.HStack>
                <button onclick={add}>Add</button>
              </atom.HStack>
            :null}
            {items.map((item,idx) => renderer(item,idx))}
          </div>
        :null}
      </atom.VStack>
    )
  }
}


// export function CollectionOld({attrs,state,style,children,hooks,vnode}) {
//   var subtype = attrs.subtype;
//   var label = attrs.label;
//   var placeholder = attrs.placeholder;
//   var onadd = attrs.onadd;
//   var ondelete = attrs.ondelete;
//   var renderer = attrs.renderer || ((f) => <ComponentField key={f.Name} field={f} nolabel={true} />);
//   var items = attrs.items || [];



//   let addingInit = (children.length === 0 && expanded_) ? true : false;
//   state.adding = (state.adding === undefined) ? addingInit: state.adding;
//   state.newValue = state.newValue || placeholder;
  
//   const toggleAdd = (e) => {
//       state.adding = (state.adding) ? false : true;
//   };

//   let expanderStyle = style.new({"marginLeft": "-10px"});
//   expanderStyle.add("mb-0", () => !expanded);
//   expanderStyle.add("mb-2", () => expanded);

//   const changeNewValue = (e, ov) => {
//       if (ov) {
//           state.newValue = ov;
//           return;
//       }
//       state.newValue = e.target.checked || e.target.value;
//       if (e.target.type === "number") {
//           state.newValue = e.target.valueAsNumber;
//       }
//   };
//   const add = (e) => {
//       if (onadd) onadd(state.newValue);
//       state.newValue = undefined;
//       state.adding = false;
//   };

//   style.add("flex flex-col select-none mt-1", {minHeight: "28px"});
//   return (
//       <div>
//           <molecule.Expander 
//               expanded={expanded} 
//               {...expanderStyle.attrs()} 
//               onclick={toggleExpander}>
              
//               <div onclick={toggleExpander} class="label flex-grow h-4">{label}</div>
//               <span class="mr-2 h-4" style={{color:"#888"}}>{items.length} items</span>
//               <atom.Icon class="mr-2 mt-1" fa="fas fa-plus-circle" onclick={toggleAdd} />
//           </molecule.Expander>
//           {state.adding && <CollectionItem class="flex flex-col my-4">
//               <Input field={subtype} onchange={changeNewValue} value={state.newValue} bind={false} />
//               <atom.Button class="mt-2" label="Add" onclick={add} />
//           </CollectionItem>}
//           {expanded && items.map((el, idx) => <CollectionItem idx={idx} ondelete={ondelete} removable draggable>{renderer(el)}</CollectionItem>)}
//       </div>
//   )
// }