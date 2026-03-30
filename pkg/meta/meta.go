package meta

import (
	"fmt"
	"log"
	"path"
	"reflect"
)

var Components map[string]reflect.Type
var Factories map[string]reflect.Value

func NodeFactories() (f map[string]reflect.Value) {
	f = make(map[string]reflect.Value)
	for path, val := range Factories {
		if val.Type().NumOut() != 1 {
			continue
		}
		if val.Type().Out(0).String() == "*node.Raw" {
			f[path] = val
		}
	}
	return
}

func ComponentFactory(name string) Factory {
	return Factory{name: name}
}

type Factory struct {
	name string
}

func (cf Factory) New() (any, string) {
	f, ok := Factories[cf.name]
	if !ok {
		log.Panicf("factory not found: %s", cf.name)
	}
	ret := f.Call(nil)
	return ret[0].Interface(), cf.name
}

type hasRelated interface {
	RelatedComponents() []string
}

func Related(typeFQN string) (related []string) {
	t, ok := Components[typeFQN]
	if !ok {
		return
	}
	v := reflect.New(t).Interface()
	matchers := make(map[string]bool)
	if rc, ok := v.(hasRelated); ok {
		for _, r := range rc.RelatedComponents() {
			fqn := fmt.Sprintf("%s.*", path.Clean(path.Join(t.PkgPath(), r)))
			matchers[fqn] = true
		}
	}
	for matcher := range matchers {
		for p := range Components {
			ok, err := path.Match(matcher, p)
			if err != nil {
				panic(err)
			}
			if ok {
				related = append(related, p)
			}
		}
	}
	return
}
