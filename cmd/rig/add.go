package main

import (
	"context"
	"fmt"
	"log"

	"tractor.dev/toolkit-go/engine/cli"
)

func addCmd() *cli.Command {
	return &cli.Command{
		Usage: "add <node> <name>",
		Short: "Add an object",
		Long:  `Add an object with NAME to object NODE`,
		Args:  cli.ExactArgs(2),
		Run:   add,
	}
}

func add(ctx *cli.Context, args []string) {
	peer, realm := dialManifold()
	raw := realm.Resolve(args[0])
	if raw == nil {
		log.Fatal("node not found")
	}

	var id string
	_, err := peer.Call(context.Background(), "addObject", action{
		Selector: args[0],
		Value:    args[1],
	}, &id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}

func addComponentCmd() *cli.Command {
	return &cli.Command{
		Usage: "add-component <node> <type>",
		Short: "Add a component",
		Long:  `Add a component of TYPE to object NODE`,
		Args:  cli.ExactArgs(2),
		Run:   addComponent,
	}
}

func addComponent(ctx *cli.Context, args []string) {
	peer, realm := dialManifold()
	raw := realm.Resolve(args[0])
	if raw == nil {
		log.Fatal("node not found")
	}

	var id string
	_, err := peer.Call(context.Background(), "addComponent", action{
		Selector: args[0],
		Value:    args[1],
	}, &id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
