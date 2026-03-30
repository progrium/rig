// @ts-ignore
import * as ui from "/-/ui.js";
// @ts-ignore
import * as restorable from "/-/util/restorable.js";

interface ResizerAttrs {
  dir: "col"|"row";
}

export const Resizer = {
  view({attrs}) {
    const style = {
      display: "flex",
      flex: "0 0 6px",
      alignItems: "center",
      flexDirection: {col: "column", row: "row"}[attrs.dir],
      cursor: {col: "ns-resize", row: "ew-resize"}[attrs.dir],
    }
    return (
      <div style={style} onmousedown={onmousedown}>
        {{
          col: <svg xmlns='http://www.w3.org/2000/svg' style='pointer-events: none;' width='30' height='6'><path d='M0 3 h30 ' fill='none' stroke='gray'/></svg>,
          row: <svg xmlns='http://www.w3.org/2000/svg' style='pointer-events: none;' width='6' height='30'><path d='M3 0 v30' fill='none' stroke='gray'/></svg>
        }[attrs.dir]}
      </div>
    )
  }
}

function onmousedown(e) {
  let cursor, sizeProp, posProp
  let col = e.target.parentNode.classList.contains("col")
  if (col) {
    cursor = "row-resize"
    sizeProp = "offsetHeight"
    posProp = "pageY"
  }
  let row = e.target.parentNode.classList.contains("row")
  if (row) {
    cursor = "col-resize"
    sizeProp = "offsetWidth"
    posProp = "pageX"
  }


  const prev = e.target.previousElementSibling;
	const next = e.target.nextElementSibling;
	if (!prev || !next) {
		return;
	}

  const restoreIframes = restorable.disableIframePointerEvents()
  const restoreCursor = restorable.setCursor(e.target, cursor)

	e.preventDefault();

	let prevSize = prev[sizeProp];
	let nextSize = next[sizeProp];
	let sumSize = prevSize + nextSize;
	let prevGrow = Number(prev.style.flexGrow);
	let nextGrow = Number(next.style.flexGrow);
	let sumGrow = prevGrow + nextGrow;
	let lastPos = e[posProp];

	function onmousemove(e) {
		let pos = e[posProp];
		let d = pos - lastPos;
		prevSize += d;
		nextSize -= d;
		if (prevSize < 0) {
			nextSize += prevSize;
			pos -= prevSize;
			prevSize = 0;
		}
		if (nextSize < 0) {
			prevSize += nextSize;
			pos += nextSize;
			nextSize = 0;
		}

		prev.style.flexGrow = sumGrow * (prevSize / sumSize);
		next.style.flexGrow = sumGrow * (nextSize / sumSize);

		lastPos = pos;
	}

	function onmouseup(e) {
		restoreCursor()
    restoreIframes()
    
    if (window["onlayoutchange"]) {
      let obj = {};
      obj[prev.id] = Number(prev.style.flexGrow);
      obj[next.id] = Number(next.style.flexGrow);
      window["onlayoutchange"](obj);
    }

    window.removeEventListener("mousemove", onmousemove);
		window.removeEventListener("mouseup", onmouseup);
	}

	window.addEventListener("mousemove", onmousemove);
	window.addEventListener("mouseup", onmouseup);
};