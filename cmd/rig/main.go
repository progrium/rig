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
		Short:   "Rig object model and runtime",
	}
	cmd.AddCommand(serveCmd())
	cmd.AddCommand(treeCmd())
	cmd.AddCommand(getCmd())
	cmd.AddCommand(signalCmd())
	cmd.AddCommand(destroyCmd())
	cmd.AddCommand(duplicateCmd())
	cmd.AddCommand(moveCmd())
	cmd.AddCommand(renameCmd())
	cmd.AddCommand(attrCmd())
	cmd.AddCommand(valueCmd())
	cmd.AddCommand(catalogCmd())
	cmd.AddCommand(addCmd())
	cmd.AddCommand(addComponentCmd())

	if err := cli.Execute(context.Background(), cmd, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
