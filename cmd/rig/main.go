package main

import (
	"context"
	"log"
	"os"

	"tractor.dev/toolkit-go/engine/cli"
)

var Version = "dev"

func main() {
	log.SetFlags(log.Lshortfile)

	cmd := &cli.Command{
		Version: Version,
		Usage:   "rig",
		Short:   "Hi",
		Long:    `Hello world\nAgain\n\n`,
	}
	cmd.AddCommand(serveCmd())
	cmd.AddCommand(inspectCmd())
	cmd.AddCommand(treeCmd())

	if err := cli.Execute(context.Background(), cmd, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
