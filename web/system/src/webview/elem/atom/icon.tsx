import m from "mithril";

interface IconAttrs {
  name: string;
  onclick?: (e: any) => void;
  style?: any;
}

export class Icon implements m.ClassComponent<IconAttrs> {
  view({ attrs }: m.CVnode<IconAttrs>) {
    const { name, onclick, style } = attrs;

    return (
      <div
        onclick={onclick}
        style={{
          zIndex: "1",
          minWidth: ".75rem",
          maxWidth: "1.5rem",
          textAlign: "center",
          ...(style || {}),
        }}
      >
        <i style={{ lineHeight: "inherit" }} class={name}></i>
      </div>
    );
  }
}