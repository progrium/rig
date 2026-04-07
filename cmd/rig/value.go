package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kr/pretty"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/telepath"
	"tractor.dev/toolkit-go/engine/cli"
)

func valueCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "value",
		Short: "Node value commands",
	}
	cmd.AddCommand(getValueCmd())
	cmd.AddCommand(setValueCmd())
	cmd.AddCommand(delValueCmd())
	cmd.AddCommand(metaValueCmd())
	cmd.AddCommand(callValueCmd())
	return cmd
}

func getValueCmd() *cli.Command {
	return &cli.Command{
		Usage: "get <field>",
		Short: "Get a field value",
		Long:  ``,
		Args:  cli.MinArgs(1),
		Run: func(ctx *cli.Context, args []string) {
			_, realm := dialManifold()
			parts := strings.SplitN(args[0], "/", 2)
			raw := realm.Resolve(parts[0])
			if raw == nil {
				log.Fatal("node not found")
			}
			value, err := telepath.Select(node.Value(raw), parts[1]).Value()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%# v\n", pretty.Formatter(value))
		},
	}
}

func setValueCmd() *cli.Command {
	var appendVal bool
	var insert int
	cmd := &cli.Command{
		Usage: "set [--append] [--insert=<idx>] <field> <value>",
		Short: "Set a field value",
		Long:  ``,
		Args:  cli.ExactArgs(2),
		Run: func(ctx *cli.Context, args []string) {
			_ = appendVal // todo
			_ = insert    // todo
			peer, realm := dialManifold()
			parts := strings.SplitN(args[0], "/", 2)
			raw := realm.Resolve(parts[0])
			if raw == nil {
				log.Fatal("node not found")
			}
			_, err := peer.Call(context.Background(), "setValue", action{
				Selector: args[0],
				Value:    args[1],
			})
			if err != nil {
				log.Fatal(err)
			}

		},
	}
	cmd.Flags().BoolVar(&appendVal, "append", false, "append to list value")
	cmd.Flags().IntVar(&insert, "insert", -1, "insert at index (list values); -1 means replace/set normally")
	return cmd
}

func delValueCmd() *cli.Command {
	return &cli.Command{
		Usage: "del <field>",
		Short: "Unset/delete a field value",
		Long:  ``,
		Args:  cli.RangeArgs(1, 2),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}

func metaValueCmd() *cli.Command {
	return &cli.Command{
		Usage: "meta <field>",
		Short: "Metadata/schema for a field value",
		Long:  ``,
		Args:  cli.ExactArgs(1),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}

func callValueCmd() *cli.Command {
	return &cli.Command{
		Usage: "call <field> [<args>...]",
		Short: "Call a method on a field value",
		Long:  ``,
		Args:  cli.MinArgs(1),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}
