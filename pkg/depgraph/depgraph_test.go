package depgraph

import (
	"reflect"
	"testing"
)

// Test objects for the dependency resolver
type TestA struct{}
type TestB struct{}
type TestC struct{}

func (a *TestA) Deps(b *TestB) {}
func (b *TestB) Deps(c *TestC) {}
func (c *TestC) Deps(a *TestA) {} // Introduces a circular dependency for testing

// Additional simple types for successful dependency graph test
type TestX struct{}
type TestY struct{}

func (x *TestX) Deps(y *TestY) {}

func TestDependencyGraph(t *testing.T) {
	// Test successful resolution
	types := []reflect.Type{
		reflect.TypeOf(&TestX{}),
		reflect.TypeOf(&TestY{}),
	}
	graph := NewDependencyGraph(types, "Deps", "Provides")
	order, err := graph.Resolve()
	if err != nil {
		t.Errorf("Failed to resolve dependency graph: %v", err)
	}
	if len(order) != 2 {
		t.Errorf("Dependency resolution order should have 2 types, got %d", len(order))
	}

	// Test circular dependency
	circularTypes := []reflect.Type{
		reflect.TypeOf(&TestA{}),
		reflect.TypeOf(&TestB{}),
		reflect.TypeOf(&TestC{}),
	}
	circularGraph := NewDependencyGraph(circularTypes, "Deps", "Provides")
	_, err = circularGraph.Resolve()
	if err == nil {
		t.Errorf("Expected error for circular dependency, got none")
	}
}

type Fooer interface {
	Foo()
}

type Bazer interface {
	Baz()
}

type NotAvailable struct{}

type Provided struct{}

type A struct{}

func (a *A) Baz() {}

func (a *A) Deps(f *F, na *NotAvailable) {}

type B struct{}

func (b *B) Foo() {}

func (b *B) Deps(fs []*F, a *A) {}

type C struct{}

func (c *C) Deps(p *Provided) {}

func (c *C) Foo() {}

type D struct{}

func (d *D) Deps(a *A, foos []Fooer) {}

type E struct{}

func (e *E) Deps(baz Bazer) {}

type F struct{}

func (f *F) Provides() *Provided {
	return nil
}

func TestBiggerDependencyGraph(t *testing.T) {
	values := []any{
		&A{},
		&A{},
		&B{},
		&C{},
		&D{},
		&E{},
		&F{},
		&F{},
	}
	order, missing, err := ResolveValues(values...)
	if err != nil {
		t.Errorf("Failed to resolve dependency graph: %v", err)
	}
	if len(missing) != 1 {
		t.Errorf("should be missing 1 dep: %v", missing)
	}
	if len(order) != (len(values) - 2) { // 2 of existing types (A, F)
		t.Errorf("Dependency resolution order should have %d types, got %d", len(values)-2, len(order))
	}
}
