package main

import (
	"fmt"

	"tractor.dev/toolkit-go/engine/cli"
)

func moveCmd() *cli.Command {
	return &cli.Command{
		Usage: "move <node> <to>",
		Short: "Move an object or component",
		Long:  ``,
		Args:  cli.ExactArgs(2),
		Run: func(ctx *cli.Context, args []string) {
			fmt.Println("TODO")
		},
	}
}
