package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"tractor.dev/toolkit-go/engine/cli"
)

func catalogCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "catalog",
		Short: "Catalog commands",
	}
	cmd.AddCommand(genCatalogCmd())
	cmd.AddCommand(listCatalogCmd())
	return cmd
}

func listCatalogCmd() *cli.Command {
	return &cli.Command{
		Usage: "list",
		Short: "List types available in catalog",
		Run: func(ctx *cli.Context, args []string) {
			peer, _ := dialManifold()

			var symbols []string
			_, err := peer.Call(context.Background(), "listCatalog", nil, &symbols)
			if err != nil {
				log.Fatal(err)
			}

			sort.Strings(symbols)
			for _, s := range symbols {
				fmt.Println(s)
			}
		},
	}
}
