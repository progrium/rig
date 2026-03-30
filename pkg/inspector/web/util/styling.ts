// DEPRECATED

// styling module provides a helper for building style
// and class attributes for elements.

type Styling = string | Object | Style | (() => boolean);
type ConditionedStyle = [string, () => boolean];

export function from(...styling: Styling[]): Style {
    return Style.from(...styling);
}

export class Style {
  static propInt(prop: string, el: HTMLElement = document.documentElement): number {
      return parseInt(Style.prop(prop, el), 10);
  }
  
  static prop(prop: string, el: HTMLElement = document.documentElement): string {
      return getComputedStyle(el).getPropertyValue(prop);
  }

  static from(...styling: Styling[]): Style {
      let s = new Style();
      s.add(...styling);
      return s;
  }

  _styling: Set<ConditionedStyle>;

  constructor() {
      this._styling = new Set();
  }

  add(...styling: Styling[]) {
      let condition = () => true;
      if (styling.length > 1 && isFunction(last(styling))) {
          condition = styling.pop() as (() => boolean);
      }
      for (let s of styling) {
          if (s === undefined) {
              continue;
          }
          if (isFunction(s)) {
              s = (s as Function).name;
          }
          this._styling.add([s as string, condition]);
      }
  }

  style(): Object {
      let style = {};
      let styling = filterByConditions(this._styling);
      for (let s of styling) {
          if (isObject(s)) {
              style = Object.assign(style, s);
              continue;
          }
          if (isStyle(s)) {
              style = Object.assign(style, (s as Style).style());
          }
      }
      return style;
  }

  class(): string {
      let classes = new Set();
      let styling = filterByConditions(this._styling);
      for (let s of styling) {
          if (isString(s)) {
              classes.add(s);
              continue;
          }
          if (isStyle(s)) {
              for (let c of (s as Style).class().split(' ')) {
                  classes.add(c);
              }
          }
      } 
      return [...classes].join(' ');
  }

  attrs(attrs: Object = {}): Object {
      return Object.assign(attrs, {
          class: this.class(),
          style: this.style(),
      });
  }
}

function filterByConditions(v: Set<any>): Array<Styling> {
  return [...v].filter(el => el[1]()).map(el => el[0]);
}

function isString(v: any): boolean {
  return v.constructor.name === "String";
}

function isObject(v: any): boolean {
  return v.constructor.name === "Object";
}

function isStyle(v: any): boolean {
  return v.constructor.name === "Style" && v._styling;
}

function isFunction(v: any): boolean {
  return v.constructor.name === "Function";
}

function last(v: Array<any>): any {
  return v[v.length-1]
}
