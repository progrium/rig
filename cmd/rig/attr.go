package main

import "tractor.dev/toolkit-go/engine/cli"

func attrCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "attr",
		Short: "Node attribute commands",
	}
	cmd.AddCommand(lsAttrCmd())
	cmd.AddCommand(getAttrCmd())
	cmd.AddCommand(setAttrCmd())
	cmd.AddCommand(delAttrCmd())
	return cmd
}

func lsAttrCmd() *cli.Command {
	return &cli.Command{
		Usage: "ls <node>",
		Short: "List attributes",
		Long:  `List attributes on NODE`,
		Args:  cli.ExactArgs(1),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}

func getAttrCmd() *cli.Command {
	return &cli.Command{
		Usage: "get <node> <attr>",
		Short: "Get attribute",
		Long:  `Get attribute ATTR on NODE`,
		Args:  cli.ExactArgs(2),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}

func setAttrCmd() *cli.Command {
	return &cli.Command{
		Usage: "set <node> <attr> <value>",
		Short: "Set attribute",
		Long:  `Set attribute ATTR to VALUE on NODE`,
		Args:  cli.ExactArgs(3),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}

func delAttrCmd() *cli.Command {
	return &cli.Command{
		Usage: "del <node> <attr>",
		Short: `Unset/delete attribute`,
		Long:  "Delete attribute ATTR on NODE",
		Args:  cli.ExactArgs(2),
		Run: func(ctx *cli.Context, args []string) {
		},
	}
}
