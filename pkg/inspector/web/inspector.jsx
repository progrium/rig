import * as manifold from "/-/inspector/manifoldold/mod.ts"
import * as field from "/-/inspector/field/mod.ts"
import * as prop from "/-/inspector/prop/mod.ts"
import * as atom from "/-/inspector/elem/atom/mod.ts"


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
        window.parent.postMessage({action: "manifold.activate", args: [e.target.name]}, "*");
      } else {
        window.parent.postMessage({action: "manifold.deactivate", args: [e.target.name]}, "*");
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
      <div style={{margin: "10px", marginTop: "5px"}}>
        <div><h2>{subject.name}</h2> {/*status*/}</div>
        <hr />
        <div>
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
                extra={<atom.Icon name="far fa-ellipsis-h" onclick={(e) => window.parent.postMessage({menu: "menu-more", params: {id: f.ID, clientX: e.clientX, clientY: e.clientY}}, "*") } />}>
              {(f.Fields||[]).map((f) => <field.Inspector field={f} store={attrs.store} />)}
            </prop.Nested>
          )
        }
        </div>
        <hr />
        {!subject.isComponent && <button onclick={(e) => {
          window.parent.postMessage({action: "manifold.addComponent", args: [subject.id]}, "*");
        }}><atom.Icon name="far fa-plus" /></button>}
        &nbsp;
        {!subject.isComponent && <button onclick={(e) => {
          window.parent.postMessage({action: "manifold.newComponent", args: [subject.id]}, "*");
        }}><atom.Icon name="far fa-file-plus" /></button>}
      </div>
    )
  }
}