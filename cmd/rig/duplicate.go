package main

import (
	"fmt"

	"tractor.dev/toolkit-go/engine/cli"
)

func duplicateCmd() *cli.Command {
	return &cli.Command{
		Usage: "duplicate <node>",
		Short: "Copy an object or component",
		Long:  ``,
		Args:  cli.ExactArgs(1),
		Run: func(ctx *cli.Context, args []string) {
			fmt.Println("TODO")
		},
	}
}
