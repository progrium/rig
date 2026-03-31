import m from "mithril";

interface DotsAttrs {
  color: string;
  size: number;
  cols: number;
  rows: number;
}

export class Dots implements m.ClassComponent<DotsAttrs> {
  view({ attrs }: m.CVnode<DotsAttrs>) {
    const color = attrs.color || "#222";
    const size = attrs.size || 4;
    const cols = attrs.cols;
    const rows = attrs.rows;

    const style: Partial<CSSStyleDeclaration> & { [key: string]: string } = {
      backgroundImage: `radial-gradient(${color} 50%, transparent 50%)`,
      backgroundColor: "transparent",
      backgroundRepeat: "repeat",
      backgroundSize: `${size}px ${size}px`,
      opacity: "70%",
      width: "100%",
    };
    if (cols) {
      style["width"] = `${size * cols}px`;
    }
    if (rows) {
      style["height"] = `${size * rows}px`;
    }
    return <div style={style} />;
  }
}