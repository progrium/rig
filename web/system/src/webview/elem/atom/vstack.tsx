import m from "mithril";

interface VStackAttrs {
  align?: string;
  style?: any;
}

export class VStack implements m.ClassComponent<VStackAttrs> {
  view({ attrs, children }: m.CVnode<VStackAttrs>) {
    const align = attrs.align || "stretch";

    return (
      <div class="VStack" style={{
        ...attrs.style,
        display: "flex",
        flexDirection: "column",
        alignItems: align,
      }}>
        {children}
      </div>
    );
  }
}