package main

import (
	"context"
	"fmt"
	"log"

	"tractor.dev/toolkit-go/engine/cli"
)

func renameCmd() *cli.Command {
	return &cli.Command{
		Usage: "rename <node> <name>",
		Short: "Rename an object",
		Long:  ``,
		Args:  cli.ExactArgs(2),
		Run:   rename,
	}
}

func rename(ctx *cli.Context, args []string) {
	peer, realm := dialManifold()
	raw := realm.Resolve(args[0])
	if raw == nil {
		log.Fatal("node not found")
	}

	if _, err := peer.Call(context.Background(), "callMethod", action{
		Selector: fmt.Sprintf("%s/SetName", args[0]),
		Args:     []any{args[1]},
	}); err != nil {
		log.Fatal(err)
	}
}

type action struct {
	Selector string
	Type     string
	Value    any
	Args     []any
}
