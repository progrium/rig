
export const Label = {
  view({attrs}) {
    var text = attrs.text;

    return (
      <label title={text} style={{
        zIndex: "1",
        textOverflow: "ellipsis", 
        whiteSpace: "nowrap",
      }}>
        {text}
      </label>
    )
  }
}