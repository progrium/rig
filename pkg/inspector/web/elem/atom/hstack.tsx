
export const HStack = {
  view({attrs, children}) {
    var align = attrs.align || "stretch";
  
    return (
      <div class="HStack" style={{
        display: "flex",
        flexDirection: "row",
        alignItems: align,
      }}>
        {children}
      </div>
    )
  }
}