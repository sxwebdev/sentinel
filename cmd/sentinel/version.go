package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func versionCMD() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "print current version",
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Printf("%s version %s\n", appName, getBuildVersion())
			return nil
		},
	}
}
