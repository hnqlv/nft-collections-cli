package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	app := &cli.App{
		Name:    "opensea",
		Usage:   "Opensea Inspector Testing",
		Version: "0.0.1",
	}

	app.Commands = []*cli.Command{
		collectionsCommand,
	}

	err := app.Run(os.Args)

	if err != nil {
		sugar.Fatal(err)
	}
}
