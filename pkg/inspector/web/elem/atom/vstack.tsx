
export const VStack = {
  view({attrs, children}) {
    var align = attrs.align || "stretch";
    
    return (
      <div class="VStack" style={{
        display: "flex",
        flexDirection: "column",
        alignItems: align,
      }}>
        {children}
      </div>
    )
  }
}