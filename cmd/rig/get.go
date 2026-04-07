package main

import (
	"fmt"
	"log"

	"github.com/kr/pretty"
	"github.com/progrium/rig/pkg/node"
	"tractor.dev/toolkit-go/engine/cli"
)

func getCmd() *cli.Command {
	return &cli.Command{
		Usage: "get <node>",
		Short: "Inspect an object or component",
		Long:  ``,
		Args:  cli.ExactArgs(1),
		Run:   get,
	}
}

func get(ctx *cli.Context, args []string) {
	_, realm := dialManifold()
	raw := realm.Resolve(args[0])
	if raw == nil {
		log.Fatal("node not found")
	}
	snapshot := node.Snapshot(raw)
	fmt.Printf("%# v", pretty.Formatter(node.Raw{
		ID:         snapshot.ID,
		Kind:       snapshot.Kind,
		Name:       snapshot.Name,
		Value:      snapshot.Value,
		Component:  snapshot.Component,
		Parent:     snapshot.Parent,
		Attrs:      snapshot.Attrs,
		Children:   snapshot.Children,
		Components: snapshot.Components,
		N:          snapshot.N,
	}))
}
