// @ts-ignore
import * as ui from "/-/inspector/ui.js";

interface DotsAttrs {
  color: string;
  size: number;
  cols: number;
  rows: number;
}

export class Dots extends ui.Element {
  onrender({attrs}: {attrs: DotsAttrs}) {
    const color = attrs.color || "#222";
    const size = attrs.size || 4;
    const cols = attrs.cols;
    const rows = attrs.rows;

    const style = {
      backgroundImage: `radial-gradient(${color} 50%, transparent 50%)`,
      backgroundColor: "transparent",
      backgroundRepeat: "repeat",
      backgroundSize: `${size}px ${size}px`,
      opacity: "70%",
      width: "100%",
    };
    if (cols) {
      style["width"] = `${size*cols}px`;
    }
    if (rows) {
      style["height"] = `${size*rows}px`;
    }
    return <div style={style} />
  }
}