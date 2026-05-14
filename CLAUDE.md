- satisfy user requests with the rig command unless it involves editing source code
- user requests are often transcribed, so consider possible mis-transcriptions
- "Maine" is a common mis-transcription of "Main"
- use `rig tree --id` to get node ids used for arguments that take NODE
- nodes are components or objects. components are shown in parens in `rig tree` output
- unless specified, "Main" refers to the Main object, not the main.Main component
- when talking about values, unless otherwise specified, this implies components because typically components have values (so "setting a value on Main" would probably mean the main.Main component)
- nodes (but mainly components) have a value you can see using `rig get`
- component values can be manipulated using `rig value` commands
- the `field` argument of `rig value` commands is `<node>/<path to field of value>`
- so if component `abc123` has a struct value with a field `Foo` that is a struct with field `Bar` this can be addressed with `abc123/Foo/Bar`
- new components can be defined by adding to the go source code in `/src`. create a go source file for each component and define a struct in it. make sure files have g+rw permissions. changes will be picked up once written.
- "create a component with so and so fields" means create a component struct type with those fields. if types are not specified, guess.

### editors

"editors" are new webview tabs for interacting with objects. editors dont need 
a hierarchy/browser or inspector because these are already provided in the UI.

important: presently they are readonly.

new editors can be created at the user's request to view and/or manipulate nodes. to create a new editor, create
a directory in `/go/src/github.com/progrium/rig/web/editors`. this will be the editor name. make sure the
directory and files have g+rw permissions.
an editor requires at least one file `editor.jsx` which must export `Editor`, which is a 
mithril.js component. these are objects with a single `view` method that takes a `vnode`
argument that has `args`, `state`, `dom`. you can see an example in the existing `rig` editor.
typically jsx is used for the `view` return value.

attrs passed to the `Editor` component include:
- `nodeID` the id of the node being edited
- `realm` an object with access to nodes via `resolve(id: string): null|Node`
- the node API is defined in `/go/src/github.com/progrium/rig/web/system/src/webview/manifold/manifold.ts`

typically you will want to subscribe to the add/remove/change events on `realm` to re-render the editor
on changes to the model.

you can also see full mithril.js docs at: https://mithril.js.org/#components

### systems

when asked for a new "system" it usually means components to model the system,
 an editor made to interact with them, and an example object tree. 

### rig commands
```
Usage:
rig [command]

Rig object model and runtime

Available Commands:
  serve            
  tree             Display a hierarchy of objects
  get              Inspect an object or component
  signal           
  destroy          Remove an object or component
  duplicate        Copy an object or component
  move             Move an object or component
  rename           Rename an object
  attr             Node attribute commands
  value            Node value commands
  catalog          Catalog commands
  add              Add an object
  add-component    Add a component

Flags:
  -v    show version

Use "rig [command] --help" for more information about a command.
```
### rig get
```
Usage:
rig get <node>

Inspect an object or component
```
Example output:
```
# rig get @main
node.Raw{
    ID:         "@main",
    Kind:       "obj",
    Bus:        "",
    Name:       "Main",
    Value:      nil,
    Component:  "",
    Parent:     "@root",
    Attrs:      {},
    Children:   nil,
    Components: {"d7bupasm689c4b2ab54g"},
    Embedded:   {},
    Refs:       {},
    N:          0x1,
    realm:      nil,
    root:       (*node.Raw)(nil),
    mu:         sync.RWMutex{},
    Dispatcher: signal.Dispatcher[github.com/progrium/rig/pkg/node.Node]{},
}
```
### rig add
```
Usage:
rig add <node> <name>

Add an object with NAME to object NODE
```
### rig destroy
```
Usage:
rig destroy <node>

Remove an object or component
```
### rig value
```
Usage:
rig value [command]

Node value commands

Available Commands:
  get              Get a field value
  set              Set a field value
  del              Unset/delete a field value
  meta             Metadata/schema for a field value
  call             Call a method on a field value

Use "rig value [command] --help" for more information about a command.

```
### rig value get
```
Usage:
rig value get <field>

Get a field value
```
### rig value set
```
Usage:
rig value set <field> <value>

Set a field value

```