import m from "mithril";
import { Input } from "./input";

interface PasswordAttrs {
  value: string;
  readonly?: boolean;
  onchange?: (e: any) => void;
  oninput?: (e: any) => void;
  onfocus?: (e: any) => void;
  onfocusout?: (e: any) => void;
}

export class Password implements m.ClassComponent<PasswordAttrs> {
  view({ attrs }: m.CVnode<PasswordAttrs>) {
    return (
      <Input type="password" {...attrs} />
    )
  }
}

export function DemoPassword() {
  return <Password value={"secret"} />
}
