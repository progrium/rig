package telepath

import (
	"context"
	"path"
	"strings"

	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/rpc"
)

type Remote struct {
	rpc.Caller
	prefix string
}

func NewRemote(caller rpc.Caller, prefix string) Root {
	return Remote{
		Caller: caller,
		prefix: prefix,
	}
}

func (r Remote) Select(path ...string) Cursor {
	return C{Root: r, Telepath: Telepath{strings.TrimLeft(strings.Join(path, "/"), "/")}}
}

func (r Remote) call(selector string, args, ret any) error {
	// TODO: timeout field on client for context
	_, err := r.Caller.Call(context.Background(), path.Join(r.prefix, selector), args, ret)
	if err != nil {
		return err
	}
	return nil
}

func (r Remote) Get(path string) (ret any, err error) {
	err = r.call("Get", fn.Args{path}, &ret)
	return
}
func (r Remote) Set(path string, v any) (err error) {
	err = r.call("Set", fn.Args{path, v}, nil)
	return
}
func (r Remote) List(path string) (ret []string, err error) {
	err = r.call("List", fn.Args{path}, &ret)
	return
}
func (r Remote) Meta(path string) (ret Metadata, err error) {
	err = r.call("Meta", fn.Args{path}, &ret)
	return
}
func (r Remote) Delete(path string) (err error) {
	err = r.call("Delete", fn.Args{path}, nil)
	return
}
func (r Remote) Insert(path string, idx int, v any) (err error) {
	err = r.call("Insert", fn.Args{path, idx, v}, nil)
	return
}
func (r Remote) Call(path string, args []any) (ret []any, err error) {
	err = r.call("Call", fn.Args{path, args}, &ret)
	return
}
