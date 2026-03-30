package main

import "tractor.dev/toolkit-go/engine/cli"

func inspectCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "inspect",
		Long:  ``,
		Run:   inspect,
	}
	return cmd
}

func inspect(ctx *cli.Context, args []string) {
}
