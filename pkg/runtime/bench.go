package runtime

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/progrium/rig/pkg/field"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/meta"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/telepath"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/rpc"
)

type NodeProvider interface {
	Nodes(parent manifold.Node) node.Nodes
}

type NodeToggler interface {
	OnEnable(obj manifold.Node)
	OnDisable(obj manifold.Node)
}

type NodeDeleter interface {
	Delete(n manifold.Node) bool
}

type NodeSelector interface {
	OnSelected()
}

type Workbench struct {
	root      *node.Raw
	activated map[string]bool
}

func New(root *node.Raw) *Workbench {
	return &Workbench{root: root, activated: make(map[string]bool)}
}

func (t *Workbench) Fields(r rpc.Responder, c *rpc.Call) {
	var id string
	c.Receive(&id)

	n := t.findNode(id)
	if n == nil {
		r.Return(fmt.Errorf("id not found"))
		return
	}

	ch, err := r.Continue()
	if err != nil {
		log.Println(err)
		return
	}
	defer ch.Close()

	if n.Kind() == node.Component {
		if n.Value() == nil {
			return
		}
		com := field.FromValue(n.Value(), field.WithFieldInfo(n.Name(), n.ID()))
		for _, f := range field.ToData(com).Fields {
			if err := r.Send(f); err != nil {
				log.Println(err)
				return
			}
		}
	} else {
		for _, n := range n.Components().Nodes() {
			if n.Value() == nil {
				continue
			}
			f := field.FromValue(n.Value(), field.WithFieldInfo(n.Name(), n.ID()))
			if err := r.Send(field.ToData(f)); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (t *Workbench) Select(id string) error {
	n := t.findNode(id)
	if n == nil {
		return fmt.Errorf("id not found")
	}
	sel := node.Get[NodeSelector](n)
	if sel != nil {
		sel.OnSelected()
	}
	return nil
}

func toComponentCase(s string) string {
	// Convert to title case and remove non-letter characters
	var result strings.Builder
	toUpper := true

	for _, r := range s {
		if unicode.IsLetter(r) {
			if toUpper {
				result.WriteRune(unicode.ToUpper(r))
				toUpper = false
			} else {
				// Preserve original capitalization
				result.WriteRune(r)
			}
		} else {
			// Set toUpper for the start of the next word
			toUpper = true
		}
	}
	return result.String()
}

type newComponent struct {
	PkgPath  string
	FilePath string
}

func (t *Workbench) ImplementFor(iface, name, fieldID string) (newComponent, error) {
	parts := strings.SplitN(fieldID, "/", 2)
	fieldPath := parts[1]
	n := t.findNode(parts[0])
	if n == nil {
		return newComponent{}, fmt.Errorf("id not found")
	}

	_, err := exec.LookPath("impl")
	if err != nil {
		return newComponent{}, err
	}

	dotparts := strings.SplitN(iface, ".", 2)
	ifacePkg := path.Base(dotparts[0])

	var typeName, typePath, fileName string
	sources := []string{"package main\n\n"}
	if strings.HasPrefix(name, "main.") {
		// existing
		dotparts := strings.SplitN(name, ".", 2)
		typeName = dotparts[1]
		typePath = name
		fileName = fmt.Sprintf("%s.go", strings.ToLower(ifacePkg))
	} else {
		// new
		typeName = toComponentCase(name)
		typePath = fmt.Sprintf("main.%s", typeName)
		fileName = fmt.Sprintf("%s.go", strings.ToLower(typeName))
		sources = append(sources, fmt.Sprintf(`
type %s struct {
	manifold.Component
}

`, typeName))
	}

	cmd := exec.Command("impl", fmt.Sprintf("c *%s", typeName), iface)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return newComponent{}, fmt.Errorf("impl: %w", err)
	}
	sources = append(sources, string(b))

	if fileExists(fileName) {
		b, err := os.ReadFile(fileName)
		if err != nil {
			return newComponent{}, err
		}
		// replace the top chunk which should
		// just be the package declaration
		sources[0] = string(b)
	}

	source := strings.Join(sources, "\n")
	fmtr := exec.Command("goimports")
	fmtr.Stdin = bytes.NewBufferString(source)
	out, err := fmtr.CombinedOutput()
	if err != nil {
		return newComponent{}, fmt.Errorf("goimports: %w", err)
	}

	com := node.NewRaw(typePath, nil, "")
	com.Kind = node.Component
	com.Component = typePath
	// todo: check if this component has already been added
	c, err := n.Parent().AddComponent(com)
	if err != nil {
		return newComponent{}, err
	}
	if err := c.SetAttr("enabled", "true"); err != nil {
		return newComponent{}, err
	}

	if err := n.SetAttr(fmt.Sprintf("ref:%s", fieldPath), c.ID()); err != nil {
		return newComponent{}, err
	}

	if err := os.WriteFile(fileName, []byte(out), 0644); err != nil {
		return newComponent{}, err
	}
	dir, _ := os.Getwd()
	return newComponent{
		PkgPath:  typePath,
		FilePath: filepath.Join(dir, fileName),
	}, nil
}

func (t *Workbench) NewComponent(id, name string) (newComponent, error) {
	n := t.findNode(id)
	if n == nil {
		return newComponent{}, fmt.Errorf("id not found")
	}

	typeName := toComponentCase(name)
	typePath := fmt.Sprintf("main.%s", typeName)
	fileName := fmt.Sprintf("%s.go", strings.ToLower(typeName))
	source := fmt.Sprintf(`package main

import (
	"context"

	"github.com/progrium/rig/pkg/manifold"
)

type %s struct {
	manifold.Component
}

func (c *%s) Activate(ctx context.Context) error {
	return nil
}

`, typeName, typeName)

	com := node.NewRaw(typePath, nil, "")
	com.Kind = node.Component
	com.Component = typePath
	c, err := n.AddComponent(com)
	if err != nil {
		return newComponent{}, err
	}
	if err := c.SetAttr("enabled", "true"); err != nil {
		return newComponent{}, err
	}

	if err := os.WriteFile(fileName, []byte(source), 0644); err != nil {
		return newComponent{}, err
	}
	dir, _ := os.Getwd()
	return newComponent{
		PkgPath:  typePath,
		FilePath: filepath.Join(dir, fileName),
	}, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func (t *Workbench) Signal(id, signal string) error {
	n := t.findNode(id)
	if n == nil {
		return fmt.Errorf("id not found")
	}
	node.Send(n, signal, nil)
	return nil
}

func (t *Workbench) Dump(id string) {
	n := t.findNode(id)
	if n == nil {
		log.Panicf("unable to find id: %s", id)
	}
	spewer := spew.ConfigState{MaxDepth: 2, Indent: "\t"}
	spewer.Fdump(os.Stderr, n.Entity())
}

func (t *Workbench) GetTreeItem(id string) treeItem {
	n := t.findNode(id)
	if n == nil {
		log.Panicf("unable to find id: %s", id)
	}
	return t.nodeItem(n)
}

type method struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func nodeMethods(n manifold.Node, prefix string) (methods []method) {
	rv := reflect.ValueOf(n.Value())
	if rv.IsValid() && rv.Kind() == reflect.Pointer {
		for i := 0; i < rv.Elem().Type().NumMethod(); i++ {
			m := rv.Elem().Type().Method(i)
			if !m.IsExported() {
				continue
			}
			if m.Type.NumIn() != 1 {
				continue
			}
			methods = append(methods, method{
				ID:    fmt.Sprintf("%s/%s", n.ID(), m.Name),
				Label: fmt.Sprintf("%s%s", prefix, m.Name),
			})
		}
		for i := 0; i < rv.Type().NumMethod(); i++ {
			m := rv.Type().Method(i)
			if !m.IsExported() {
				continue
			}
			if m.Type.NumIn() != 1 {
				continue
			}
			methods = append(methods, method{
				ID:    fmt.Sprintf("%s/%s", n.ID(), m.Name),
				Label: fmt.Sprintf("%s%s", prefix, m.Name),
			})
		}
	}
	return
}

func (t *Workbench) Methods(id string) (methods []method) {
	n := t.findNode(id)
	if n == nil {
		return
	}
	methods = append(methods, nodeMethods(n, "")...)
	for _, com := range n.Components().Nodes() {
		methods = append(methods, nodeMethods(com, fmt.Sprintf("%s:", com.Name()))...)
	}
	return
}

// very rudimentary
func (t *Workbench) Call(id string) {
	parts := strings.Split(id, "/")
	n := t.findNode(parts[0])
	if n == nil {
		return
	}
	method := reflect.ValueOf(n.Value()).MethodByName(parts[1])
	if method.IsValid() {
		method.Call([]reflect.Value{})
		return
	}
	field := reflect.Indirect(reflect.ValueOf(n.Value())).FieldByName(parts[1])
	if field.IsValid() && field.Type().Kind() == reflect.Func && !field.IsNil() {
		field.Call([]reflect.Value{})
	}
}

func (t *Workbench) Shutdown() {
	for id, e := range t.activated {
		if e {
			t.Toggle(id, false)
		}
	}
}

func (t *Workbench) findNode(id string) manifold.Node {
	if id == "" {
		return manifold.FromEntity(t.root)
	}
	return manifold.FromEntity(node.GetStore(t.root).Resolve(id))
}

func (t *Workbench) Toggle(id string, state bool) error {

	var ids []string
	if id != "" {
		ids = []string{id}
	} else {
		ids = t.root.Children
	}

	for _, id := range ids {
		n := t.findNode(id)
		if n == nil {
			return fmt.Errorf("not found")
		}

		if node.IsObject(n) {
			if state {
				if err := node.Activate(context.TODO(), n); err != nil {
					log.Println(err)
					return nil
				}
				t.activated[id] = state
			} else {
				if err := node.Deactivate(context.TODO(), n); err != nil {
					log.Println(err)
					return nil
				}
				delete(t.activated, id)
			}
		} else {
			if state {
				if err := node.EnableComponent(n); err != nil {
					log.Println(err)
				}
			} else {
				if err := node.DisableComponent(n); err != nil {
					log.Println(err)
				}
			}
		}
	}

	return nil
}

func (t *Workbench) Rename(id, name string) {
	n := t.findNode(id)
	if n == nil {
		return
	}
	n.SetName(name)
}

func (t *Workbench) SwitchView(id, view string) {
	n := t.findNode(id)
	if n == nil {
		return
	}
	n.SetAttr("view", view)
}

type editField struct {
	OK       bool
	Name     string
	Value    string
	ParentID string
	Options  []string `json:",omitempty"`
}

func (t *Workbench) PreEditField(id string) editField {
	parts := strings.Split(id, "/")
	n := t.findNode(parts[0])
	if n == nil {
		return editField{}
	}
	cur := telepath.Select(n.Value(), strings.Join(parts[1:], "/"))
	v, _ := cur.Value()
	return editField{
		OK:       true,
		Name:     path.Base(id),
		Value:    fmt.Sprintf("%v", v),
		ParentID: n.Parent().ID(),
	}
}

func (t *Workbench) EditField(id string, val any) error {
	parts := strings.Split(id, "/")
	n := t.findNode(parts[0])
	if n == nil {
		return nil
	}
	cur := telepath.Select(n.Value(), strings.Join(parts[1:], "/"))
	//v, _ := cur.Value()
	if err := cur.Set(val); err != nil {
		return err
	}
	node.Send(n, "")
	return nil
}

func (t *Workbench) Delete(id string) error {
	n := t.findNode(id)
	if n == nil {
		return nil // todo: not found?
	}
	for _, deleter := range node.GetAll[NodeDeleter](n) {
		deleter.Delete(n)
	}
	return node.Destroy(n)
}

func (t *Workbench) Views(id string) (views []string) {
	n := t.findNode(id)
	if n == nil {
		return
	}
	views = append(views, "objects")
	if n.Kind() == node.Object {
		views = append(views, "components")
	}
	if n.Value() != nil {
		views = append(views, "fields")
		_, ok := n.Value().(NodeProvider)
		if ok {
			views = append(views, "nodes")
		}
	}
	if n.Kind() == node.Object && n.Components().Count() > 0 {
		views = append(views, "fields")
		for _, c := range n.Components().Nodes() {
			_, ok := c.Value().(NodeProvider)
			if ok {
				views = append(views, c.Name())
			}
		}
	}
	return
}

func isValidStructPtr(rv reflect.Value) bool {
	return rv.Type().Kind() == reflect.Ptr && !rv.IsNil() && rv.Type().Elem().Kind() == reflect.Struct
}

func (t *Workbench) structFields(store node.Store, parentID string, rv reflect.Value) (items node.Children) {
	rv = reflect.Indirect(rv)
	for i := 0; i < rv.Type().NumField(); i++ {
		f := rv.Type().Field(i)
		if !f.IsExported() {
			continue
		}
		fv := rv.Field(i)
		id := fmt.Sprintf("%s/%s", parentID, f.Name)
		var children node.Children
		// only allows 3 levels deep
		if (fv.Type().Kind() == reflect.Struct ||
			isValidStructPtr(fv)) &&
			strings.Count(id, "/") < 3 {
			children = t.structFields(store, id, fv)
		}
		attrs := node.Attributes{
			"desc":  fmt.Sprintf("%#v", fv.Interface()),
			"field": "",
		}
		if fv.IsValid() && f.Type.Kind() == reflect.Func && f.Type.NumIn() == 0 {
			attrs = node.Attributes{
				"callable": "",
				"icon":     "zap",
			}
		}
		n := node.NewID(id, f.Name, attrs, children)
		if err := node.SetStore(n, store); err != nil {
			log.Println(err)
		}
		items = append(items, n)
	}
	return
}

func (t *Workbench) valueFields(n manifold.Node) (items []treeItem) {
	rv := reflect.ValueOf(n.Value())
	if rv.IsNil() {
		return
	}
	store := node.GetStore(n)
	for _, child := range t.structFields(store, n.ID(), rv) {
		node.SetParent(child, n.ID())
		items = append(items, t.nodeItem(manifold.FromEntity(child)))
	}
	return
}

type NodeAdder interface {
	CanAddNode() []string
	AddNode(typ string, parent manifold.Node, curView string) (bool, error)
}

func (t *Workbench) AddItem(id, typ, name string) error {
	n := t.findNode(id)
	for _, adder := range node.GetAll[NodeAdder](n) {
		ok, err := adder.AddNode(typ, n, n.Attr("view"))
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	newnode := node.New(name)
	if f, ok := meta.NodeFactories()[strings.TrimRight(typ, "()")]; ok {
		v, err := fn.Call(f, []any{})
		if err != nil {
			return err
		}
		newnode = v[0].(*node.Raw)
		if err := node.SetName(newnode, name); err != nil {
			return err
		}
	}
	if err := n.Store().Store(newnode); err != nil {
		return err
	}
	if typ != "EmptyObject" && !strings.HasSuffix(typ, "()") {
		if err := t.AddComponent(newnode.ID, typ); err != nil {
			return err
		}
	}
	if err := n.Objects().Append(manifold.FromEntity(newnode)); err != nil {
		return err
	}
	return nil
}

func (w *Workbench) AddComponent(id, typ string) error {
	n := w.findNode(id)
	var v any
	if strings.HasSuffix(typ, "()") {
		f, ok := meta.Factories[strings.TrimSuffix(typ, "()")]
		if !ok {
			return fmt.Errorf("%s factory not found", typ)
		}
		ret := f.Call(nil)
		v = ret[0].Interface()
	} else {
		t, ok := meta.Components[typ]
		if !ok {
			return fmt.Errorf("%s type not found", typ)
		}
		v = reflect.New(t).Interface()
		if i, ok := v.(node.Initializer); ok {
			i.Initialize()
		}
	}
	c, err := n.AddComponent(v)
	if err != nil {
		return err
	}
	if strings.HasSuffix(typ, "()") {
		c.SetAttr("_factory", strings.TrimSuffix(typ, "()"))
	}
	return c.SetAttr("enabled", "true")
}

func (t *Workbench) GetMainComponents() (items []string) {
	for pkgpath := range meta.Components {
		if strings.HasPrefix(pkgpath, "main.") {
			items = append(items, pkgpath)
		}
	}
	return
}

func (t *Workbench) GetAddComponents(id string) (items []string) {
	n := t.findNode(id)
	relatedCom := make(map[string]bool)
	for _, com := range node.GetAll[any](n, node.Include{Parents: true}) {
		t := reflect.Indirect(reflect.ValueOf(com)).Type()
		name := t.PkgPath() + "." + t.Name()
		for _, pkgpath := range meta.Related(name) {
			relatedCom[pkgpath] = true
		}
	}
	var others []string
	var main []string
	var related []string
	for pkgpath := range meta.Components {
		if _, isRelated := relatedCom[pkgpath]; isRelated {
			related = append(related, pkgpath)
			continue
		}
		if strings.HasPrefix(pkgpath, "main.") {
			main = append(main, pkgpath)
			continue
		}
		others = append(others, pkgpath)
	}
	for path := range meta.Factories {
		main = append(main, path+"()")
	}
	items = append(items, "--Main")
	items = append(items, main...)
	items = append(items, "--Related")
	items = append(items, related...)
	items = append(items, "--Library")
	items = append(items, others...)
	return
}

func (t *Workbench) GetAddItems(id string) (items []string) {
	n := t.findNode(id)
	items = append(items, "EmptyObject")
	for _, adder := range node.GetAll[NodeAdder](n) {
		items = append(items, adder.CanAddNode()...)

	}
	for factorypath := range meta.NodeFactories() {
		items = append(items, factorypath+"()")
	}
	items = append(items, "--with Component")
	for pkgpath := range meta.Components {
		items = append(items, pkgpath)
	}
	return
}

func (t *Workbench) ToggleComponents(id string) {
	n := t.findNode(id)
	if n.Attr("view:components") != "" {
		n.DelAttr("view:components")
	} else {
		n.SetAttr("view:components", "true")
	}
}

func (t *Workbench) GetChildren(id string) (items []treeItem) {
	n := t.findNode(id)
	v := n.Attr("view")
	if v == "" && n.Kind() == node.Object {
		v = "objects"
	}
	switch v {
	case "":
		return
	case "fields":
		if n.Kind() == node.Object {
			for _, com := range n.Components().Nodes() {
				items = append(items, t.valueFields(com)...)
			}
		} else {
			return t.valueFields(n)
		}
	case "nodes":
		for _, pn := range ProvidedNodes(n) {
			items = append(items, t.nodeItem(pn))
		}
	case "components", "objects":
		var sub []manifold.Node
		if v == "components" || n.Attr("view:components") != "" {
			sub = n.Components().Nodes()
		}
		if v == "objects" {
			sub = append(sub, n.Objects().Nodes()...)
		}
		for _, child := range sub {
			items = append(items, t.GetTreeItem(child.ID()))
		}
	default:
		for _, c := range n.Components().Nodes() {
			if c.Name() == v && c.Value() != nil {
				for _, pn := range ProvidedNodes(c) {
					items = append(items, t.nodeItem(pn))
				}
				return
			}
		}
	}
	return
}

func ProvidedNodes(n manifold.Node) (nodes []manifold.Node) {
	np, ok := n.Value().(NodeProvider)
	if !ok {
		return
	}
	store := node.GetStore(n)
	for _, pn := range np.Nodes(n) {
		nn := store.Resolve(pn.Entity().GetID())
		if nn == nil {
			if err := node.SetStore(pn, store); err != nil {
				log.Println(err)
			}
			if err := node.SetParent(pn, n.ID()); err != nil {
				log.Println(err)
			}
			nn = pn.Entity()
		}
		nodes = append(nodes, manifold.FromEntity(nn))
	}
	return
}

func (t *Workbench) nodeItem(n manifold.Node) treeItem {
	item := treeItem{
		ID:    n.ID(),
		Label: n.Name(),
	}
	var contextValues []string
	if n.Attr("view") == "" {
		switch n.Kind() {
		case node.Component:
			n.SetAttr("view", "fields")
		case node.Object:
			n.SetAttr("view", "objects")
		}
	}
	switch n.Attr("view") {
	case "objects":
		count := n.Objects().Count()
		if n.Attr("view:components") != "" {
			count += n.Components().Count()
		}
		if count > 0 {
			item.CollapsibleState = 1
		}
	case "components":
		if n.Components().Count() > 0 {
			item.CollapsibleState = 1
		}
	default:
		item.CollapsibleState = 1
	}
	if n.Attr("error") != "" {
		n.SetAttr("color", "charts.red")
		n.SetAttr("icon", "circle-filled")
	} else if n.Attr("activated") != "" {
		if n.Attr("activated") == "true" {
			contextValues = append(contextValues, "deactivator")
			n.SetAttr("color", "charts.green")
			n.SetAttr("icon", "circle-filled")
		} else {
			contextValues = append(contextValues, "activator")
			n.SetAttr("color", "activityBar.inactiveForeground")
			n.SetAttr("icon", "circle-filled")
		}
	}
	desc := n.Attr("desc")
	if desc != "" {
		item.Description = desc
	}
	if n.Kind() == node.Component {
		item.Description = n.ComponentType()
		checkState := 0
		if node.ComponentEnabled(n) {
			checkState = 1
		}
		item.CheckboxState = &checkState
	} else {
		contextValues = append(contextValues, "obj")
	}
	if slices.Contains(n.Attrs(), "callable") {
		contextValues = append(contextValues, "callable")
	}
	if slices.Contains(n.Attrs(), "deletable") {
		contextValues = append(contextValues, "deletable")
	}
	if slices.Contains(n.Attrs(), "field") {
		contextValues = append(contextValues, "field")
	}
	addItems := t.GetAddItems(n.ID())
	if len(addItems) > 0 {
		contextValues = append(contextValues, "addable")
	}
	item.ContextValue = strings.Join(contextValues, ",")
	item.Attrs = node.AttrMap(n)
	item.Tooltip = fmt.Sprintf("%s\n---\n* ID: %s\n* View: %s\n* Kind: %s", n.Name(), n.ID(), n.Attr("view"), n.Kind())
	if n.Error() != nil {
		item.Tooltip = fmt.Sprintf("%s\n---\nError: %s", n.Name(), n.Error())
	}
	return item
}

type treeItem struct {
	ID               string `json:"id,omitempty"`
	Label            string `cbor:"label"`
	Description      string `json:"description,omitempty"`
	Tooltip          string `json:"tooltip,omitempty"`
	ContextValue     string `json:"contextValue,omitempty"`
	CollapsibleState int    `json:"collapsibleState,omitempty"`
	CheckboxState    *int   `json:"checkboxState,omitempty"`
	Attrs            map[string]string
}
