import m from "mithril";
import { Input } from "./input";

interface TextAttrs {
  value?: string;
  style?: Record<string, any>;
  readonly?: boolean;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Text implements m.ClassComponent<TextAttrs> {
  view({ attrs }: m.CVnode<TextAttrs>) {
    const { style, ...rest } = attrs;
    const inner = <Input {...rest} />;
    return style ? <div style={style}>{inner}</div> : inner;
  }
}

export function DemoText() {
  return <Text value={"Hello world"} />
}
