// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import {Input} from "./input.tsx";

interface PasswordAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e) => void;
  oninput?: (e) => void;
  onfocus?: (e) => void;
  onfocusout?: (e) => void;
}

export class Password extends ui.Element {
  onrender({attrs}: {attrs: PasswordAttrs}) {
    return (
      <Input type="password" {...attrs} />
    )
  }
}

export function DemoPassword() {
  return <Password value={"secret"} />
}
