package main

import (
	"context"
	"dim/logger"
	"flag"

	"dim/examples/mock"

	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "fim",
		Version: version,
		Short:   "server",
	}

	ctx := context.Background()

	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("could not run command")
	}
}
