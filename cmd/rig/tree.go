package main

import "tractor.dev/toolkit-go/engine/cli"

func treeCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "tree",
		Long:  ``,
		Run:   tree,
	}
	return cmd
}

func tree(ctx *cli.Context, args []string) {
}
