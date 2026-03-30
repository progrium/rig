package depgraph

import (
	"fmt"
	"reflect"
)

type GraphNode struct {
	Type       reflect.Type
	Provides   []reflect.Type
	Resolved   bool
	DependsOn  map[*GraphNode]struct{}
	Dependants map[*GraphNode]struct{}
}

func (n *GraphNode) String() string {
	var depson []string
	for nn, _ := range n.DependsOn {
		depson = append(depson, nn.Type.String())
	}
	var depant []string
	for nn, _ := range n.Dependants {
		depant = append(depant, nn.Type.String())
	}
	return fmt.Sprintf("%s %v DependsOn: %v Dependants: %v", n.Type, n.Provides, depson, depant)
}

type DependencyGraph struct {
	Nodes    map[reflect.Type]*GraphNode
	Provided map[reflect.Type]*GraphNode
	Missing  map[reflect.Type]struct{}
}

func NewDependencyGraph(types []reflect.Type, depsMethod, providesMethod string) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes:    make(map[reflect.Type]*GraphNode),
		Provided: make(map[reflect.Type]*GraphNode),
		Missing:  make(map[reflect.Type]struct{}),
	}
	interfaceImpls := map[reflect.Type][]*GraphNode{}
	ifaces := map[reflect.Type]struct{}{}

	// Initialize graph nodes
	for _, t := range types {
		if _, exists := graph.Nodes[t]; exists {
			continue
		}

		node := &GraphNode{
			Type:       t,
			DependsOn:  map[*GraphNode]struct{}{},
			Dependants: map[*GraphNode]struct{}{},
		}
		graph.Nodes[t] = node

		method, ok := t.MethodByName(providesMethod)
		if ok {
			mt := method.Func.Type()
			for i := 0; i < mt.NumOut(); i++ {
				provideType := mt.Out(i)
				node.Provides = append(node.Provides, provideType)
				graph.Provided[provideType] = node
			}
		}

		method, ok = t.MethodByName(depsMethod)
		if !ok {
			continue // Skip if no Deps method
		}
		mt := method.Func.Type()
		for i := 1; i < mt.NumIn(); i++ {
			depType := mt.In(i)
			// normalize slice type to the elem type
			if depType.Kind() == reflect.Slice {
				depType = depType.Elem()
			}
			if depType.Kind() == reflect.Interface {
				ifaces[depType] = struct{}{}
			}
		}
	}

	for _, node := range graph.Nodes {
		// Check and map interfaces this type implements
		for iface := range ifaces {
			if node.Type.Implements(iface) {
				interfaceImpls[iface] = append(interfaceImpls[iface], node)
			}
		}
	}

	// Link dependencies based on Deps methods and interfaces
	for _, node := range graph.Nodes {
		method, ok := node.Type.MethodByName(depsMethod)
		if !ok {
			continue // Skip if no Deps method
		}

		mt := method.Func.Type()
		for i := 1; i < mt.NumIn(); i++ {
			depType := mt.In(i)
			// normalize slice type to the elem type
			if depType.Kind() == reflect.Slice {
				depType = depType.Elem()
			}
			if impls, found := interfaceImpls[depType]; found {
				// Handle interface dependencies
				for _, implNode := range impls {

					node.DependsOn[implNode] = struct{}{}
					implNode.Dependants[node] = struct{}{}
				}
			} else {
				// Handle direct dependencies
				if depNode, exists := graph.Nodes[depType]; exists {
					node.DependsOn[depNode] = struct{}{}
					depNode.Dependants[node] = struct{}{}
				} else if depNode, exists := graph.Provided[depType]; exists {
					node.DependsOn[depNode] = struct{}{}
					depNode.Dependants[node] = struct{}{}
				} else {
					graph.Missing[depType] = struct{}{}
					// fmt.Printf("Warning: Dependency type %v not found in graph nodes\n", depType)
				}
			}
		}
	}

	return graph
}

func (g *DependencyGraph) Resolve() ([]reflect.Type, error) {
	resolved := []*GraphNode{}
	toResolve := []*GraphNode{}

	for _, node := range g.Nodes {
		if len(node.DependsOn) == 0 {
			toResolve = append(toResolve, node)
		}
	}

	for len(toResolve) > 0 {
		current := toResolve[0]
		toResolve = toResolve[1:]
		resolved = append(resolved, current)
		for dependant := range current.Dependants {
			allDepsResolved := true
			for dep := range dependant.DependsOn {
				if !contains(resolved, dep) {
					allDepsResolved = false
					break
				}
			}
			if allDepsResolved && !contains(resolved, dependant) {
				toResolve = append(toResolve, dependant)
			}
		}
	}

	if len(resolved) != len(g.Nodes) {
		return nil, fmt.Errorf("circular dependency detected or unresolved dependencies exist")
	}

	resolvedTypes := make([]reflect.Type, len(resolved))
	for i, node := range resolved {
		resolvedTypes[i] = node.Type
	}
	return resolvedTypes, nil
}

func contains(slice []*GraphNode, node *GraphNode) bool {
	for _, n := range slice {
		if n == node {
			return true
		}
	}
	return false
}

func ResolveValues(values ...any) (resolved []reflect.Type, missing []reflect.Type, err error) {
	var rv []reflect.Type
	for _, v := range values {
		rv = append(rv, reflect.TypeOf(v))
	}
	graph := NewDependencyGraph(rv, "Deps", "Provides")
	resolved, err = graph.Resolve()
	for t := range graph.Missing {
		missing = append(missing, t)
	}
	return
}
