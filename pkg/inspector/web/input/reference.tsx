// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Box} from "./box.tsx";
// @ts-ignore
import * as atom from "/-/inspector/elem/atom/mod.ts";

interface ReferenceAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}


export class Reference extends ui.Element {
  onrender({attrs, state}: {attrs: ReferenceAttrs}) {
    const onreset = attrs.onreset;

    const onchange = attrs.onchange;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;
    const readonly = attrs.readonly || false;
    let value = attrs.value;
    
    //state.now = state.now || undefined;    

    if (value === "00:00") {
      value = undefined;
    }
    // if (state.now) {
    //   value = state.now;
    //   state.now = undefined;
    // }
    
    const onkeydown = (e) => {
      if (e.key === "Backspace" && e.shiftKey) {
        e.preventDefault();
        if (onreset) {
          onreset(e);
        }
        if (onchange) {
          onchange(e, "00:00");
        }
      }
    }
    const ondblclick = (e) => {
      // let n = new Date();
      // state.now = `${("0" + n.getHours()).slice(-2)}:${("0" + n.getMinutes()).slice(-2)}`;
      // if (onchange) {
      //   onchange(e, state.now);
      // }
    }
    
    return (
      <Box>
        <input type="text" 
            readonly={readonly}
            onkeydown={onkeydown} 
            onchange={onchange}
            oninput={oninput}
            onfocus={onfocus}
            onfocusout={onfocusout}
            style={{ 
              width: "100%",
              background: "transparent",
              outline: "none", 
              border: "0",
            }}
            placeholder="--:-- --"
            value={value} />
        {/* <atom.Icon ondblclick={ondblclick} name="far fa-clock"></atom.Icon> */}
      </Box>
    )
  }
}


function ReferenceInput({attrs,style,state}) {
  var icon = attrs.icon || "";
  var onchange = attrs.onchange || undefined;
  var onunset = attrs.onunset || undefined;
  var value = attrs.value || "";
  var placeholder = attrs.placeholder || "";
  var context = attrs.context || {};
  var contextMenu = attrs.contextMenu || "";
  var bind = (attrs.bind!==undefined)?attrs.bind:true;
  var hover_ = attrs.hover || false;
  var buffer = attrs.buffer || undefined;

  state.hover = (state.hover === undefined) ? hover_ : state.hover;

  let displayIcon = "fa-asterisk";
  if (icon) {
      displayIcon = icon;
  }
  if (state.hover && onunset) {
      displayIcon = "fa-times-circle";
  }

  const mouseover = (e) => state.hover = true;
  const mouseout = (e) => state.hover = false;
  const onclick = (e) => (onunset) ? onunset(e) : undefined;

  let inputAttrs = {
      onchange: onchange,
      value: value,
  }

  if (buffer) {
      inputAttrs = Object.assign(inputAttrs, buffer.attrs);
      inputAttrs["onkeydown"] = (e) => {
          if (e.key === "Enter") {
              buffer.commit(true);
              buffer.cancel();
              e.target.blur();
          }
      }
  }

  style.add("menu-context", () => contextMenu);
  return (
      <InputBox 
          data-context-menu={contextMenu} 
          data-context={JSON.stringify(context)}
          onmouseover={mouseover} 
          onmouseout={mouseout}>
          <input {...inputAttrs}
              data-bind={JSON.stringify(bind)}
              title={placeholder}
              placeholder={placeholder}
              class="w-full"
              type="text"
          />
          <atom.Icon onclick={onclick} style={{marginTop: "2px", display: (buffer && buffer.editing)?"none":"block"}} fa={`fas ${displayIcon}`}></atom.Icon>
      </InputBox>
  )
}