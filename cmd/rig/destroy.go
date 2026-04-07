package main

import (
	"context"
	"fmt"
	"log"

	"tractor.dev/toolkit-go/engine/cli"
)

func destroyCmd() *cli.Command {
	return &cli.Command{
		Usage: "destroy <node>",
		Short: "Remove an object or component",
		Args:  cli.ExactArgs(1),
		Run:   destroy,
	}
}

func destroy(ctx *cli.Context, args []string) {
	peer, realm := dialManifold()
	raw := realm.Resolve(args[0])
	if raw == nil {
		log.Fatal("node not found")
	}

	if _, err := peer.Call(context.Background(), "callMethod", action{
		Selector: fmt.Sprintf("%s/Destroy", args[0]),
	}); err != nil {
		log.Fatal(err)
	}
}
