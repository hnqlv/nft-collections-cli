package main

import (
	"github.com/urfave/cli/v2"
)

var (
	collectionsCommand = &cli.Command{
		Name:   "collections",
		Usage:  "save active collections in the last 30d",
		Action: retrieveCollections,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "total",
				Aliases:     []string{"t"},
				Value:       5000,
				Usage:       "total collections to be retrieved",
				DefaultText: "5000",
			},
		},
	}
)
