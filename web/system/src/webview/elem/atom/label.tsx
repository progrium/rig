import m from "mithril";

interface LabelAttrs {
  text: string;
  style?: any;
}

export class Label implements m.ClassComponent<LabelAttrs> {
  view({ attrs }: m.CVnode<LabelAttrs>) {
    const text = attrs.text;

    return (
      <label title={text} style={{
        ...attrs.style,
        zIndex: "1",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
      }}>
        {text}
      </label>
    );
  }
}