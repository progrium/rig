
export type Type = "string" | "password" | "boolean" | "number" | "time" | "date" | "color";

interface constraints {
  enum?: any[];
  min?: number;
  max?: number;
  minlength?: number;
  maxlength?: number;
}

export type Formatter = (v: any, extra: any) => any;
export type Committer = (v: any, extra: any, buf: ValueBuffer) => void;

export class ValueBuffer {
  _value: any;
  extra: any;
  type: Type;
  autocommit: boolean;
  debounce: number;
  constraints: constraints;
  formatter: Formatter;
  committer: Committer;

  editing: boolean;
  buffer: any;
  debouncer: Function;

  constructor(value: any, extra: any, type: Type, committer: Committer, autocommit: boolean = false, debounce: number = 0, constraints?: constraints, formatter?: Formatter) {
      this.set(value, extra);
      this.type = type;
      this.committer = committer;
      this.formatter = formatter || ((v, extra) => v);
      this.constraints = constraints || {};

      this.editing = false;
      this.buffer = undefined;

      this.autocommit = autocommit;
      this.debounce = debounce;

      let timeout = undefined;
      this.debouncer = (wait, fn, ...args) => {
          clearTimeout(timeout);
          const later = () => {clearTimeout(timeout); fn(...args);}
          timeout = (wait === 0) ? later() : setTimeout(later, wait);
      };
  }

  get value() {
      return (this.editing) ? this.buffer : this.formatter(this._value, this.extra);
  }

  get attrs() {
      // TODO: add back constraints attrs??
      return {
          value: this.value,
          oninput: this.oninput.bind(this),
          onfocus: this.edit.bind(this),
          onfocusout: () => {
              if (this.buffer) {
                  this.commit(true);
              }
              this.cancel();
          },
      };
  }

  set(value: any, extra?: any) {
      this._value = value;
      if (extra) {
          this.extra = extra;
      }
  }

  edit() {
      this.editing = true;
      this.buffer = this._value;
  }

  cancel() {
      this.editing = false;
      this.buffer = undefined;
  }

  commit(immediate: boolean = false) {
      if (this.buffer === this._value) return;
      this.debouncer(
          (immediate) ? 0 : this.debounce, 
          (v) => this.committer(v, this.extra, this), 
          this.buffer
      );
  }

  oninput(event) {
      this.input(normalizeFormValue(event.target, this.type));
  }

  change(event, immediate: boolean = false) {
      let oldDebounce = this.debounce;
      if (immediate) {
          this.debounce = 0;
      }
      this.edit();
      this.oninput(event);
      this.cancel();
      this.debounce = oldDebounce;
  }

  input(value) {
      if (!this.editing) return;
      if (!this.validate(value)) return;

      this.buffer = value;
      
      if (this.autocommit) {
          this.commit();
      }
  }

  validate(value) {
      const c = this.constraints;
      if (value === undefined) return false;
      if (!this.type) return true;
      if (c.enum && c.enum.indexOf(value) === -1) return false;
      switch (this.type) {
      case "color":
      case "password":
      case "string":
          if (typeof value !== "string") return false;
          if (c.minlength !== undefined && value.length < c.minlength) return false;
          if (c.maxlength !== undefined && value.length > c.maxlength) return false;
          return true;
      case "boolean":
          if (typeof value !== "boolean") return false;
          return true;
      //case "range": ??
      case "number":
          if (isNaN(value)) return false;
          if (typeof value !== "number") return false;
          if (c.min !== undefined && value < c.min) return false;
          if (c.max !== undefined && value > c.max) return false;
          return true;
      default:
          console.warn(`unsupported value type: ${this.type}`);
          return false;
      }
  }

}

function normalizeFormValue(element, type) {
  if (["textarea", "select", "input"].indexOf(element.nodeName.toLowerCase()) === -1) {
      console.warn(`unsupported element type: ${element.nodeName}`);
      return undefined;
  }
  const typify = (v) => (type === "number") ? parseInt(v) : v; // todo: floats?
  if (["select-one", "password", "email", "hidden", "url", "textarea", "text"].indexOf(element.type) !== -1) {
      return typify(element.value);
  }
  switch (element.type) {
  case "checkbox":    return typify(element.checked);
  case "number":      return element.valueAsNumber;
  default:
      console.warn(`unsupported input type: ${element.type}`);
      return undefined;
  }
}