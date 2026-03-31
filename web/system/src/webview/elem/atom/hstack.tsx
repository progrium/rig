import m from "mithril";

interface HStackAttrs {
  align?: string;
  style?: any;
}

export class HStack implements m.ClassComponent<HStackAttrs> {
  view({ attrs, children }: m.CVnode<HStackAttrs>) {
    const align = attrs.align || "stretch";

    return (
      <div class="HStack" style={{
        ...attrs.style,
        display: "flex",
        flexDirection: "row",
        alignItems: align,
      }}>
        {children}
      </div>
    );
  }
}