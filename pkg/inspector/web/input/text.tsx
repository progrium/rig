// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Input} from "./input.tsx";

interface TextAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Text extends ui.Element {
  onrender({attrs}: {attrs: TextAttrs}) {
    return (
      <Input {...attrs} />
    )
  }
}

export function DemoText() {
  return <Text value={"Hello world"} />
}
