package main

import (
	"fmt"

	"tractor.dev/toolkit-go/engine/cli"
)

func signalCmd() *cli.Command {
	var listen bool
	cmd := &cli.Command{
		Usage: "signal [--listen] <node> [<signal> <args>...]",
		Long:  ``,
		Run: func(ctx *cli.Context, args []string) {
			_ = listen
			fmt.Println("TODO")
		},
	}
	cmd.Flags().BoolVar(&listen, "listen", false, "listen for signals on the node")
	return cmd
}
