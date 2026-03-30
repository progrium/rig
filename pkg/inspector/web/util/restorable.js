// collection of helper operations that return
// a function to restore previous state

// prevents iframes from breaking drag operations
export const disableIframePointerEvents = () => {
  const iframes = document.querySelectorAll("iframe")
  if (iframes.length === 0) return () => {}
  Array.from(iframes).forEach(iframe => {
    iframe.origPointerEvents = iframe.style.pointerEvents
    iframe.style.pointerEvents = "none"
  })
  return () => {
    Array.from(iframes).forEach(iframe => {
      iframe.style.pointerEvents = iframe.origPointerEvents
    })
  }
}

// set element cursor at html level as well
// to avoid cursor flickering
export const setCursor = (el, cursor) => {
  if (!cursor) return () => {}
  const html = document.querySelector('html')
  const origElCursor = el.style.cursor
  const origHtmlCursor = html.style.cursor
  el.style.cursor = cursor
  html.style.cursor = cursor
  return () => {
    el.style.cursor = origElCursor
    html.style.cursor = origHtmlCursor
  }
}

export const setStyle = (el, style) => {
  const orig = Object.keys(style).map(key => [key, el.style[key]])
  Object.keys(style).forEach(key => el.style[key] = style[key])
  return () => {
    orig.forEach(entry => el.style[entry[0]] = entry[1])
  }
}

export const addClass = (el, className) => {
  el.classList.add(className)
  return () => {
    el.classList.remove(className)
  }
}