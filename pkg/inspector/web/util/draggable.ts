// newer version of "sortable"

// @ts-ignore
import * as ui from "/-/inspector/ui.js";
// @ts-ignore
import * as restorable from "./restorable.js";

interface Options {
  axis: "x"|"y"
  parentClass: string
  cursor: string
  cursorFactory: (ctx) => Element
  placeholderFactory: (ctx) => Element
  indicatorFactory: (ctx) => Element
  oncomplete: (el: Element, idx: number) => void
}

export const emptyPlaceholder = (ctx) => {
  let rect = ctx.target.getBoundingClientRect()
  let el = document.createElement('div')
  el.classList.add('placeholder')
  el.style.display = "inline-block"
  el.style.width = `${rect.width}px`
  el.style.height = `${rect.height}px`
  return el
}

export const linePlaceholder = (ctx) => {
  let rect = ctx.target.getBoundingClientRect()
  let el = document.createElement('div')
  el.classList.add('placeholder')
  el.style.display = "inline-block"
  el.style.width = `${rect.width}px`
  el.style.height = `0px`
  return el
}

export const lineIndicator = (ctx) => {
  let rect = ctx.target.getBoundingClientRect()
  let indicator = document.createElement('div')
  indicator.classList.add('indicator')
  indicator.style.display = "inline-block"
  indicator.style.width = `${rect.width}px`
  indicator.style.height = `0px`
  indicator.style.left = `${rect.left}px`
  indicator.style.top = `${rect.top}px`
  indicator.style.position = "absolute"
  indicator.style.border = "1px dashed gray"
  indicator.style.opacity = "0.5"
  indicator.style.transition = "top 0.1s, left 0.1s, width 0.1s, height 0.1s"
  return indicator
}

// use return fn as onmousedown handler.
export function start(opts: Options) {
  const parentClass = opts.parentClass || "draggable"
  const cursor = opts.cursor || undefined
  const cursorFactory = opts.cursorFactory || undefined
  const placeholderFactory = opts.placeholderFactory || undefined
  const indicatorFactory = opts.indicatorFactory || undefined

  // onmousedown
  return (e) => {
    const target = e.target.closest(`.${parentClass} > *`)
    if (!target) return

    const rect = target.getBoundingClientRect()
    const ctx = {
      target: target,
      mouseOrigin: {
        x: e.pageX,
        y: e.pageY
      },
      targetOffset: {
        x: e.pageX - rect.left,
        y: e.pageY - rect.top,
      },
      moved: false,
      placeholder: undefined,
      indicator: undefined,
      cursor: undefined,
      restorables: undefined,
    }

    function dragging(e) {
      // Set position for dragging element
      if (ctx.cursor) {
        ctx.cursor.style.top = `${e.pageY}px`
        ctx.cursor.style.left = `${e.pageX}px`
      }

      // The current order:
      // prev
      // ctx.target
      // ctx.placeholder
      // next
      let prev = ctx.target.previousElementSibling
      if (ctx.cursor !== ctx.target) {
        prev = ctx.placeholder.previousElementSibling
      }
      let next = ctx.placeholder.nextElementSibling
      
      if (prev && isOver(opts.axis, ctx.cursor, prev)) {
        // The current order    -> The new order
        // prev                 -> ctx.target
        // ctx.target               -> ctx.placeholder
        // ctx.placeholder        -> prev
        if (ctx.cursor === ctx.target) {
          swap(prev, ctx.target)
        }
        swap(ctx.placeholder, prev)
        if (ctx.indicator) {
          ctx.indicator.style.left = `${ctx.placeholder.offsetLeft}px`
          ctx.indicator.style.top = `${ctx.placeholder.offsetTop}px`
        }
        ctx.moved = true
      }

      if (next && isOver(opts.axis, next, ctx.cursor)) {
        // The current order    -> The new order
        // ctx.target               -> next
        // ctx.placeholder        -> ctx.target
        // next                 -> ctx.placeholder
        swap(next, ctx.placeholder)
        if (ctx.cursor === ctx.target) {
          swap(next, ctx.target)
        }
        if (ctx.indicator) {
          ctx.indicator.style.left = `${ctx.placeholder.offsetLeft}px`;
          ctx.indicator.style.top = `${ctx.placeholder.offsetTop}px`;  
        }
        ctx.moved = true
      }

    }

    function startDrag() {
      ui.pause()

      if (placeholderFactory) {
        ctx.placeholder = placeholderFactory(ctx)
      } else {
        ctx.placeholder = ctx.target.cloneNode()
      }
      ctx.target.parentNode.insertBefore(ctx.placeholder, ctx.target.nextSibling)

      if (cursorFactory) {
        ctx.cursor = cursorFactory(ctx)
        document.body.appendChild(ctx.cursor)
      } else {
        ctx.cursor = ctx.target
      }

      if (indicatorFactory) {
        ctx.indicator = indicatorFactory(ctx)
        document.body.appendChild(ctx.indicator)
      }

      ctx.restorables = [
        restorable.disableIframePointerEvents(),
        restorable.setCursor(ctx.cursor, cursor),
        restorable.setStyle(ctx.cursor, {
          position: "absolute",
          left: "0",
          top: "0",
        }),
        restorable.addClass(ctx.cursor, "dragging"),
      ]

      function completeDrag() {
        document.removeEventListener('mousemove', dragging)

        ctx.restorables.forEach(fn => fn())

        // if cursor was created, move the target
        // since it was not being moved during drag
        if (ctx.cursor !== ctx.target) {
          ctx.placeholder.parentNode.insertBefore(ctx.target, ctx.placeholder)
        }

        if (ctx.placeholder) {
          ctx.placeholder.parentNode.removeChild(ctx.placeholder)
        }
      
        if (ctx.indicator) {
          ctx.indicator.parentNode.removeChild(ctx.indicator)
        }

        // only remove cursor if it wasnt the target
        if (ctx.cursor !== ctx.target) {
          ctx.cursor.parentNode.removeChild(ctx.cursor)
        }
        
        ui.resume()

        if (ctx.moved) {
          opts.oncomplete(ctx.target, siblingIndex(ctx.target))
        }
      }

      document.addEventListener('mousemove', dragging)
      document.addEventListener('mouseup', completeDrag, {once: true})
    }

    function detectDrag(e) { 
      if (e.buttons !== 1) return
      if (distance(ctx.mouseOrigin, {
        x: e.pageX,
        y: e.pageY
      }) > 1) {
        document.removeEventListener("mousemove", detectDrag)
        startDrag()
      }
    }
    document.addEventListener("mousemove", detectDrag)
    document.addEventListener("mouseup", () => {
      document.removeEventListener("mousemove", detectDrag)
    })
  }
}

const distance = (a, b) => { 
	let xd = a.x - b.x
  let yd = a.y - b.y
	return Math.sqrt(xd * xd + yd * yd)
}

const swap = function(nodeA, nodeB) {
  const parentA = nodeA.parentNode;
  const siblingA = nodeA.nextSibling === nodeB ? nodeA : nodeA.nextSibling;

  // Move `nodeA` to before the `nodeB`
  nodeB.parentNode.insertBefore(nodeA, nodeB);

  // Move `nodeB` to before the sibling of `nodeA`
  parentA.insertBefore(nodeB, siblingA);
};

const isOver = function(axis, nodeA, nodeB) {
  // Get the bounding rectangle of nodes
  const rectA = nodeA.getBoundingClientRect();
  const rectB = nodeB.getBoundingClientRect();

  switch (axis) {
  case "x":
    return (rectA.left + rectA.width / 2 < rectB.left + rectB.width / 2);
  case "y":
    return (rectA.top + rectA.height / 2 < rectB.top + rectB.height / 2);
  }
  
};

const siblingIndex = function(el) {
  let children = el.parentNode.childNodes;
  for (let i = 0; i < children.length; i++) {
    if (children[i] === el) {
      return i;
    }
  }
}


// const mouseWithin = (el, ev) => {
//   const rect = el.getBoundingClientRect()
//   return (
//     ev.pageX > rect.left && 
//     ev.pageX < rect.left+rect.width &&
//     ev.pageY > rect.top && 
//     ev.pageY < rect.top+rect.height)
// }

// const findOver = (els, ev) => {
//   let found = []
//   for (let i=0; i<els.length; i++) {
//     if (mouseWithin(els[i], ev)) {
//       found.push(els[i])
//     }
//   }
//   return found[found.length-1]
// }
