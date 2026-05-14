import * as field from "./field/mod";
import * as prop from "./prop/mod";
import * as atom from "./elem/atom/mod";


export const Inspector = {
  view: ({attrs}) => {
    
    let subject = window.nodes.resolve(attrs.selectedID);
    if (!subject) {
      return <div>no selection</div>
    }

    // let status = undefined
    // console.log(subject.Value[0])
    // try {
    //   let s = subject.Value[0].Value.S
    //   if (s < 0) {
    //     status = <span style={{float: "right", color: "red"}}>{ui.trust("&#x2B24;")}</span>
    //   } else if (s >= 0 && s < 100) {
    //     status = <span style={{float: "right", color: "gray"}}>{ui.trust("&#x2B24;")}</span>
    //   } else if (s >= 100 && s < 200) {
    //     status = <span style={{float: "right"}}><img src="/-/backplane/loading.gif" style="width: 20px; height: 20px;" /></span>
    //   } else if (s >= 200 && s < 300) {
    //     status = <span style={{float: "right", color: "green"}}>{ui.trust("&#x2B24;")}</span>
    //   } else if (s >= 300 && s < 400) {
    //     status = <span style={{float: "right", color: "orange"}}>{ui.trust("&#x2B24;")}</span>
    //   }
    // } catch {}
    
    const oncheck = (e) => {
      if (e.target.checked) {
        window.vscode.postMessage({action: "manifold.activate", args: [e.target.name]}, "*");
      } else {
        window.vscode.postMessage({action: "manifold.deactivate", args: [e.target.name]}, "*");
      }
    }
    const isEnabled = (id) => {
      const node = window.nodes.resolve(id);
      if (!node) {
        console.warn("unable to find:", id);
        return false;
      }
      return node.raw.Attrs["enabled"]==="true";
    };
    return (
      <div style={{margin: "10px", marginTop: "0px", marginLeft: "20px"}}>
        <div class="id">{subject.id}</div>
        <div class="name">{subject.name}</div>
        <hr />
        <div class="value">
        {(subject.isComponent)
          ?(attrs.fields||[]).map((f) => <field.Inspector field={f} store={attrs.store} />)
          :(attrs.fields||[]).map((f,idx) => 
            <prop.Nested 
                data-id={f.ID} 
                key={f.ID} 
                pre={<input 
                  type="checkbox" 
                  name={f.ID} 
                  oninput={oncheck} 
                  checked={isEnabled(f.ID)} />}
                label={f.Name} 
                extra={<div data-vscode-context={`{"webviewSection": "more", "preventDefaultContextMenuItems": true, "id": "${f.ID}"}`}><atom.Icon name="far fa-ellipsis-h" 
                  onclick={(e) => {
                  e.preventDefault();
                  e.target.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, clientX: e.clientX, clientY: e.clientY }));
                  e.stopPropagation();
                  //window.vscode.postMessage({menu: "menu-more", params: {id: f.ID, clientX: e.clientX, clientY: e.clientY}}, "*") 
                  }} 
                /></div>}>
              {(f.Fields||[]).map((f) => <field.Inspector field={f} store={attrs.store} />)}
            </prop.Nested>
          )
        }
        </div>
        <hr />
        {!subject.isComponent && <button onclick={(e) => {
          window.vscode.postMessage({action: "manifold.addComponent", args: [subject.id]}, "*");
        }}><atom.Icon name="far fa-plus" /></button>}
        &nbsp;
        {!subject.isComponent && <button onclick={(e) => {
          window.vscode.postMessage({action: "manifold.newComponent", args: [subject.id]}, "*");
        }}><atom.Icon name="far fa-file-plus" /></button>}
      </div>
    )
  }
}