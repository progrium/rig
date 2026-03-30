
export const Icon = {
  view({attrs}) {
    var name = attrs.name;

    return (
      <div style={{
        zIndex: "1",
        minWidth: ".75rem",
        maxWidth: "1.5rem",
        textAlign: "center",
      }}>
        <i style={{lineHeight: "inherit"}} class={name}></i>
      </div>
    )
  }
}