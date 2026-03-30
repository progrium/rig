# `rig` System and Package Documentation

This document explains the architecture of the `rig` tree and gives conceptual guidance for working in each package.

## What `rig` Is

`rig` is the core runtime and composition layer for AutoRig. It models an object/component graph, supports dependency-aware activation, emits graph-level signals, exposes graph state via reflective path access, and provides concrete catalogs of components for host integration (web, files, subprocesses, media, etc).

At a high level:

- `node` stores the canonical in-memory graph (`Raw` nodes + `Store`).
- `entity` provides interface-first graph operations and signaling helpers.
- `manifold` is the ergonomics layer over `entity`/`node` (typed Node/List APIs).
- `runtime` hosts and serves the graph to editors and other peers.
- `module` handles module lifecycle and persistence.
- `meta` + `library` provide component registration and discovery metadata.
- `telepath` enables reflective path-based read/write/call operations.
- `catalog/*` holds reusable domain components.

---

## System Overview

### Graph Model

Rig uses a hierarchical graph with two concrete node kinds:

- object nodes (`Kind == "obj"`)
- component nodes (`Kind == "com"`)

Object nodes contain:

- child object IDs (`Children`)
- component IDs (`Components`)
- attributes (`Attrs`) used for runtime/UI state (`view`, `enabled`, `activated`, `error`, etc)
- optional value payloads (`Value`)

Component nodes usually wrap pointers to Go structs that implement behavior (for example, `Activate`, `Deactivate`, `Provides`, `Nodes`, or custom interfaces used by runtime/workbench views).

### Storage and Identity

`node.Store` is the authoritative ID-to-node registry. All graph traversals resolve IDs via store lookups. Root and main IDs are special (`@root`, `@main`), and node IDs are otherwise generated (`xid`).

The store also supports:

- export/import of node snapshots
- reference repair for pointer-like value fields (via `pointers` + `telepath`)
- signal forwarding to watchers

### Lifecycle and Activation

Object activation typically happens through `node.Activate` / `node.Deactivate`:

- components are assembled using `engine`
- activator interfaces are invoked
- object attrs are updated (`busy`, `activated`, `error`)

For more deterministic dependency ordering, components can use `util.ObjectActivator`, which resolves component order through `depgraph` based on `Assemble` and `Provides` method signatures.

Component-level on/off is separate:

- `node.EnableComponent`
- `node.DisableComponent`

These call optional methods on component values and set `enabled`.

### Signal Propagation

Signals are generic (`signal.Signal[T]`) and used broadly. Entity mutation helpers (`entity.SetAttr`, `SetName`, `SetValue`, etc) typically emit signals automatically.

`node.Raw.Signaled` propagates:

- to local watchers
- to store watchers
- to component signal receivers on the same object
- upward to parent receivers

This makes the graph reactive enough for remote UI mirrors and state observers.

### Reflection and Remote Access

`telepath` is the path-based reflective data plane:

- list traversable names
- inspect metadata/schema
- get/set/delete values
- call methods
- insert into slices

It supports local roots and remote RPC roots (`telepath.Remote`), which is central to runtime/editor bridge behavior.

### Runtime Service Model

`runtime.Run` sets up:

- module load/save plumbing (`module`)
- root/super node with service components (bridge, inspector)
- unix socket service for workbench/editor peers
- signal streaming and command handling

`runtime.Workbench` is the orchestration component that powers introspection, tree views, field editing, add/remove operations, and generated component scaffolding flows.

---

## Dataflow: Common Operations

### Creating a Node

1. Call `node.New(...)`.
2. Facets are interpreted as attrs, children, or component values.
3. Components are wrapped as component nodes, attached, and optionally initialized.
4. Components may auto-enable or mark object activation state.
5. Store assignment occurs when object is inserted/stored.

### Activating an Object

1. `node.Activate` sets `busy`.
2. Components are assembled and activated (or delegated to strategy).
3. On success, object is marked `activated=true` when stateful.
4. On error, object attr `error` is set and activation stops.

### Persisting Module State

1. `module.M.Save()` asks store to export snapshots.
2. `node.Store.Export()` includes refs for pointer-linked values.
3. `module.Provider` writes serialized nodes (JSON provider available).
4. On load, import rebuilds values, restores refs, and reattaches where needed.

---

## Package-by-Package Conceptual Overview

## `catalog`

Reusable component catalogs, mostly organized by domain. These packages provide concrete behavior on top of the generic graph runtime.

### `catalog/web`

HTTP/websocket-oriented components:

- listener abstraction (`Listener`)
- `TCPListener` implementation
- `Server` component (serves handler, optional browser launch)
- `FileServer`, `Router`, `Route`, matcher-based router
- `SocketUpgrader` for websocket upgrade dispatch

This package bridges manifold components with standard `net/http` service assembly.

### `catalog/host`

Host OS/process/filesystem integration:

- subprocess lifecycle (`Subprocess`) with process group termination
- host filesystem tree projection (`Directory`/`File` node providers)

Useful for exposing local resources as graph nodes.

### `catalog/tree`

Maps manifold object trees into an `fs.FS`-compatible interface. Nodes can act as virtual files/directories, enabling tooling that expects filesystem semantics.

### `catalog/file`

File-like buffer component (`Buffer`) with read/write methods and editor bridge integration. Selecting it in workbench can open/edit synchronized buffer content.

### `catalog/gfx` and `catalog/gfx/three`

Lightweight graphics-related data/components:

- vector types
- frame update loop (`UpdateLoop`) with updater/flusher interfaces
- three-style scene/geometry/material placeholder component types

### `catalog/ui/quick`

Minimal UI schema components (`Layout`, `Column`, `Text`, `Input`, etc) with handler assembly and related-component hinting.

### `catalog/debug`

Simple debug payload structs with nested and pointer fields, useful for testing inspector and field introspection behavior.

### `catalog/_ssh`

SSH server components and terminal screen integration (`tcell` + `gliderlabs/ssh`).

### `catalog/_obs`, `catalog/_livekit`, `catalog/_vnd/*`

Specialized and/or experimental integrations with external systems:

- OBS integration with subprocess/client/event modeling
- LiveKit server/resource abstractions
- vendor-ish integration helpers (e.g. ngrok listener)

Underscore-prefixed folders suggest internal, experimental, or non-stable catalog surfaces.

## `debouncer`

Simple function debouncer returning `func(func())`. Used to coalesce repeated events (e.g. autosave or event flooding).

## `depgraph`

Reflection-based dependency resolver over component types:

- builds directed graph from method signatures (`Deps`/`Provides` style)
- supports interface dependencies and provider outputs
- resolves topological order and reports missing deps

Central for deterministic activation ordering in `util.ObjectActivator`.

## `entity`

Interface-driven graph access facade:

- core entity interfaces (`E`, `Node`, `StoreEntity`, etc)
- helpers for attrs/value/name/parent/component operations
- mutation helpers with signal emission
- list and hierarchy utilities (parents, siblings, children, counts, indexes)
- destroy/error helpers

This package decouples higher layers from concrete `node.Raw`.

## `inspector`

Live inspector server and field model translation:

- serves embedded inspector web UI (`inspector/web`)
- tracks graph updates via pubsub topics
- streams snapshot deltas to peers
- supports value editing through telepath selectors

### `inspector/field`

Type/field/value model system used to convert reflected Go values into a structured, UI-friendly schema (`Data`) with flags, enums, ranges, nested fields, and pointer metadata.

### `inspector/web`

Embedded static assets for inspector UI (served through `webfs` transform-aware filesystem).

## `library`

Component catalog definitions and helpers:

- generic catalog registration container
- canonical component IDs from Go type info
- component metadata struct (`Prefix`, `Source`, `Icon`, `Desc`)
- source filepath helper

## `manifold`

Developer-facing graph API layer:

- `Node`, `Bus`, `List` abstractions
- wrapper type over entities (`N`)
- convenient object/component list operations
- walker and tree dump helpers
- component embed helper (`manifold.Component`)

This is the package most component code should depend on for graph interaction ergonomics.

## `meta`

Runtime registry metadata:

- known component types map
- known factory functions map
- derived node factories
- related-component lookup expansion

Used heavily by runtime/workbench add-component and creation flows.

## `misc`

Small utility primitives:

- generic `Must`
- URL join helper
- ephemeral listen address helper
- retry utilities
- buffered in-memory pipe for streaming logs/sessions

## `module`

Module lifecycle and persistence boundary:

- `M` wraps store, name, provider, save debounce
- loading from provider or defaults
- save/export through provider
- provider abstraction (`Exists`, `LoadAll`, `SaveAll`, etc)
- JSON provider implementation included

Represents a persisted graph module instance.

## `node`

Core graph implementation:

- raw node struct and ID generation
- creation (`New`, `NewID`, `NewComponent`)
- store implementation with import/export
- activation/deactivation and component enable/disable
- node context utilities
- type-safe component fetch helpers (`Get`, `GetAll`, includes)
- signaling behavior and propagation

This is the foundational runtime package for entity storage and lifecycle.

## `pointers`

Reflective walker to discover pointer/interface-valued fields and map them to telepath paths. Primarily supports persistence reference extraction/repair.

## `pubsub`

Generic in-memory topic with queue + subscribers:

- publish/subscribe/unsubscribe
- dispatch loop
- close semantics
- activate/deactivate convenience methods

Used in inspector and elsewhere for fan-out event streams.

## `resource`

Generic resource modeling helpers for external system objects:

- resource identity/status/sync checks
- list/create/read/delete provider interfaces
- helper for list-backed read
- helper node constructors for “new resource” flows
- list reconciliation helpers that map external resources to child graph nodes

Commonly used by catalog integrations that expose remote platform resources.

## `runtime`

Runtime host and interactive workbench layer:

- process lifecycle and signal handling
- module boot/load and service startup
- unix socket RPC service setup
- bridge integration (editor command channel)
- dynamic tree/fields/methods/editing/add/delete/toggle operations

`Workbench` is effectively the control plane for graph introspection and manipulation.

## `shlex`

Shell-like lexer/token splitter with configurable tokenizer and POSIX/non-POSIX mode support.

## `signal`

Generic observer/signaling primitives:

- signal type and receiver interface
- dispatcher/watcher implementation
- helper send and channel-based receive loop

Widely reused by `entity`, `node`, and runtime services.

## `state`

Small state-machine utility:

- enum state model
- transition validation
- optional transition hook (`Transitioner`)
- embeddable `State` provider

Used by resource/stateful component logic.

## `telepath`

Reflective selector/cursor protocol for object graphs:

- path walking and traversable-name extraction
- cursor abstraction (value/list/meta/set/delete/insert/call)
- local root implementation (`R`)
- remote RPC root implementation (`Remote`)
- type metadata and schema extraction

Critical for inspector editing and runtime RPC-based introspection.

## `util`

Activation strategy helpers:

- `ObjectActivator` dependency-aware activation/deactivation
- deprecated toggler strategy (`ObjectToggler`, older method-based pattern)

This package encapsulates advanced lifecycle strategy logic shared by objects.

## `webfs`

Transforming filesystem wrapper:

- wraps `fs.FS`
- on-the-fly transforms for `.jsx`, `.tsx`, `.ts`, and script tags in `.html`
- powered by esbuild

Used for serving live/transformed frontend assets from embedded or live dirs.

---

## Architectural Boundaries and Responsibilities

When adding new code under `rig`, keep these boundaries in mind:

- put core graph mechanics in `node`/`entity`
- put ergonomic wrappers in `manifold`
- put reflection/path operations in `telepath`
- put cross-cutting small helpers in `misc`/`debouncer`/`signal`
- put domain integrations in `catalog/*`
- put runtime UI/editor orchestration in `runtime`
- put persistence concerns behind `module.Provider`

This keeps the runtime modular and avoids coupling domain integrations to low-level graph internals.

## Extension Patterns

Common extension approaches:

- add a new component type and register metadata (`meta`/catalog registration path)
- implement lifecycle interfaces (`Activate`, `Deactivate`, `Provides`, optional `Assemble`)
- expose derived nodes with `Nodes(parent manifold.Node)` provider patterns
- use attrs for view/runtime status semantics (`view`, `desc`, `deletable`, `enabled`, `activated`)
- use `resource` helpers for external list/create/read/delete-backed entities

## Notes on Stability

- Some areas are explicitly marked TODO/deprecated/experimental in code.
- Underscored catalog packages likely indicate non-final APIs.
- Runtime/workbench behavior is pragmatic and reflective by design; favor integration testing when changing activation, telepath, or store import/export semantics.

## Musings from AI

It’s a live programmable substrate. A few observations:

* The core idea feels like a fusion of ECS-ish composition, dependency injection, and reflective runtime introspection. `node` + `entity` + `manifold` create a graph that is both executable and inspectable.

* `telepath` is unusually significant. Most systems stop at in-process reflection; this one turns reflection into an addressable protocol. That makes remote tooling feel native rather than bolted on.

* The runtime/workbench model suggests the system is designed to be operated while running, not just compiled and deployed. Editing fields, adding components, toggling activation, and generating scaffolds from the UI is a powerful feedback loop.

* Persistence is pragmatic and clever: serialize graph snapshots, then repair references through path-based pointer mapping. That points to a philosophy of preserving live structure, not just config values.

* The catalog pattern is important culturally: it keeps “platform integrations” (web, host, OBS, LiveKit, SSH) as composable graph parts instead of hardcoded subsystems. That’s a strong extensibility story.

* There’s a visible tension between experimentation and hardening (TODOs, deprecated paths, underscore packages). But that tension is healthy in systems like this; it reads like a toolkit growing from real usage rather than top-down architecture diagrams.

If I had to summarize the significance in one line: this looks like infrastructure for building stateful, inspectable, hot-operable software worlds—where runtime structure is as important as source code.

### Why is that cool?

Because it changes the unit of software from “files and functions” to a living graph you can think with.

Why that’s cool in practice:

* You can see and steer behavior live instead of guessing from logs and restarts.

* Introspection isn’t a separate observability stack; it’s built into the model (telepath, inspector, workbench).

* Composition is cheap: new capability is often “attach a component,” not “rewrite architecture.”

* External systems (web, processes, OBS, LiveKit, etc.) become first-class nodes, so ops and product logic share one substrate.

* It shortens the idea loop: model -> run -> inspect -> tweak -> persist, all in one environment.

The deeper cool factor: it feels like building a tiny operating system for your domain, where structure, state, and tooling are unified. That’s rare.

### How would you explain it to a six year old?

Imagine you have a giant box of smart LEGO pieces.

* Some LEGO pieces are things (like houses, trees, cars).
* Some LEGO pieces are skills (like “can sing,” “can drive,” “can shine light”).
* You can snap skills onto things to make them do stuff.

This system is like that, but for computer worlds.

And here’s the magic part:

* There is a special control room where you can watch all your LEGO world pieces.
* While the world is running, you can change pieces, add new skills, and see what happens right away.
* If you like it, you can save your world and open it later exactly how it was.

So it’s cool because it feels less like writing boring instructions, and more like building and steering a living toy world.

### How would you explain it to a 12 year old?

Think of it like building a game engine where the game world is made of objects + powers:

* An object is a thing (room, bot, server, file, camera).
* A component is a power/ability you attach to that thing (serve web pages, connect to LiveKit, read files, react to events).

So instead of writing one giant program, you build a graph of pieces that connect.

What makes this system special is:

* You can inspect that graph live (like a debug map of the whole world).
* You can turn components on/off while it runs.
* You can edit values live (like changing settings in a running game).
* You can save the state and load it back later.

A good mental model is:

* Minecraft redstone + Unity components + live devtools in one system.

Why that’s powerful:

you’re not just coding static files - you’re shaping a running system you can observe, tweak, and evolve in real time.

### How would you explain it to a 24 year old computer science phd?

A concise framing:

Rig is a reflective, persistent, runtime-editable object-component graph with first-class lifecycle semantics and remote structural introspection.

If I were pitching it to a CS PhD:

* The core data structure is a typed-ish heterogeneous graph (obj/com nodes) with stable identities and store-backed resolution.
* Behavior is attached compositionally via component values (Go objects), not inheritance or static wiring.
* Execution is lifecycle-driven (Activate/Deactivate, enable/disable) with optional dependency-ordered assembly (depgraph over Assemble/Provides signatures).
* State mutation is evented via a generic signal system; propagation is graph-aware (local, store, component, ancestor).
* Introspection/manipulation is formalized as a path calculus (telepath): get/set/list/meta/delete/insert/call, with local and RPC-backed roots.
* Persistence snapshots graph state and reconstructs referential structure by pointer-path reconciliation, not merely naive value serialization.
* The workbench/runtime layer closes the loop: live topology inspection, field editing, component creation/toggling, and remote control over a running system.

Why it’s interesting (research-adjacent angle):

* It sits at an intersection of entity-component systems, dataflow/reactive runtime design, reflective metaprogramming, and live programming environments.

* It treats “program” as an operable stateful artifact rather than a compiled static object.

* It effectively unifies model, control plane, and observability in one substrate.

* The architecture invites experimentation in gradual typing over reflective graphs, consistency models for live mutation, and richer dependency/lifecycle semantics (e.g., partial-order scheduling, transactional graph updates).

In one line: it’s a practical live systems substrate that makes the runtime graph—not source files—the primary locus of composition and control.