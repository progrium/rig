import m from "mithril";

interface KnobAttrs {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  sensitivity?: number;
  stateless?: boolean;
  highlight?: string;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

interface KnobState {
  color: string;
  value: number;
}

export class Knob implements m.ClassComponent<KnobAttrs> {
  color: string = "white";
  value: number = 0;

  view({ attrs }: m.CVnode<KnobAttrs>) {
    const value = attrs.value || 0;
    const min = attrs.min || 0;
    const max = attrs.max || 100;
    const step = attrs.step || 1;
    const sensitivity = attrs.sensitivity || 1;
    const stateless = attrs.stateless || false;
    const highlight = attrs.highlight || "white";
    const onchange = attrs.onchange || undefined;
    const oninput = attrs.oninput || undefined;
    const onfocus = attrs.onfocus || undefined;
    const onfocusout = attrs.onfocusout || undefined;

    this.color = (this.color === undefined) ? "white" : this.color;
    this.value = (this.value === undefined) ? value : this.value;
    if (stateless) {
      this.value = value;
    }
    const rot = (this.value * 300 / (max - min)) + 30;

    const onmousedown = (e: MouseEvent) => {
      let lastX = e.pageX;
      let lastY = e.pageY;

      this.color = highlight;

      const onmousemove = (e: MouseEvent) => {
        const offsetX = e.pageX - lastX;
        const offsetY = e.pageY - lastY;
        const diff = (offsetX - offsetY) * sensitivity;
        this.value = Math.min(max, Math.max(min, this.value + diff));
        m.redraw();
        lastX = e.pageX;
        lastY = e.pageY;
      };

      const onmouseup = (e: MouseEvent) => {
        this.color = "white";
        m.redraw();

        document.body.removeEventListener("mouseup", onmouseup);
        document.body.removeEventListener("mousemove", onmousemove);
        return false;
      };

      document.body.addEventListener("mouseup", onmouseup);
      document.body.addEventListener("mousemove", onmousemove);

      return false;
    };
    
    const style = {
      width: "1.5rem",
      height: "1.5rem",
      background: "#484848",
      borderRadius: "50%",
      position: "relative",
      top: "0.25rem",
      boxShadow: "inset -1px -2px 2px #222",
      border: "1px solid #050505",
    }

    const indicator = {
      height: "0.8rem",
      width: "0.2rem",
      position: "relative",
      left: "50%",
      top: "50%",
      marginLeft: "-2px",
      borderRadius: "2px",
      transform: `rotate(${rot}deg)`,
      transformOrigin: "50% 0%",
      backgroundColor: this.color,
    }

    return (
      <div onmousedown={onmousedown} style={style}>
        <div style={indicator}></div>
        <input type="range" 
               style={{display: "none"}} 
               onchange={onchange} 
               oninput={oninput}
               onfocus={onfocus}
               onfocusout={onfocusout}
               min={min} 
               max={max} 
               value={this.value} 
               step={step} />
      </div>
    )
  }
}

export function DemoKnobAtom() {
  return <Knob value={0} highlight="yellow" />
}

