import * as qtalk from "/-/vnd/qtalk/qtalk.min.js";

var popover = undefined;

export function show(event, config) {
  // allow closer to work on current popover
  setTimeout(() => {
    if (config.x && config.y) {
      popover = config;
      h.redraw();
      return;
    }
    let el = event.target;
    let pos = {
      x: el.offsetLeft, 
      y: el.offsetTop+el.offsetHeight,
    };
    if (config.mouseOffset) {
      pos = {
        x: event.clientX+config.mouseOffset.x, 
        y: event.clientY+config.mouseOffset.y,
      }
    }
    popover = Object.assign(pos, config);
    h.redraw();
    // wait for elements to be rendered
    setTimeout(() => {
      let el = document.querySelector("#popover input");
      if (el) el.focus();
    }, 20);
  }, 1);
}

function popoverCloser(fn) {
  return (e) => {
    if (fn) fn(e);
    popover=undefined;
  }
}

export class Overlay {

  setup() {
    const frame = document.getElementById("popover");
    const peer = qtalk.open("popover", new qtalk.JSONCodec(), {
      ready: async (r, c) => {
        const resp = await peer.call("render", popover);
        frame.width = resp.reply.w;
        frame.height = resp.reply.h;
      },
      click: async (r, c) => {
        if (!popover) return;
        const label = await c.receive();
        for (const item of popover.menu) {
          if (item.label === label && item.onclick) {
            const ret = item.onclick();
            if (ret !== true) {
              popover=undefined;
              h.redraw();
            }
            return;
          }
        }
      }
    });
  }

  view(v) {
    const overlayStyle = {
      zIndex: "10",
      backgroundColor: "transparent",
      position: "fixed",
      top: "0px",
      left: "0px",
      width: "100%",
      height: "100%",
    };
    return (!popover) ? 
      null :
      <div style={overlayStyle} onclick={popoverCloser()}>
        <iframe id="popover"
                onload={() => this.setup()}
                style={{
                  position: "absolute", 
                  border: "4px solid black", 
                  left: `${popover.x}px`,
                  top: `${popover.y}px`
                }} 
                src="/-/gadget/popover.html"></iframe>
      </div>
  }
}