package main

import (
	"os"

	"github.com/dorfire/heavenly/pkg/earthfile"
	"github.com/dorfire/heavenly/pkg/earthfilefmt"
	"github.com/google/go-cmp/cmp"
	"github.com/urfave/cli/v2"
)

// TODO: support '#' comments
func formatEarthfile(ctx *cli.Context) error {
	ef, err := earthfile.Parse(ctx.Args().First())
	if err != nil {
		return err
	}
	logger.DebugPrintf("Parsed %s", ef.Path)

	orig, err := os.ReadFile(ef.Path)
	if err != nil {
		return err
	}

	formatted := earthfilefmt.Format(ef.Spec)

	logger.Printf(cmp.Diff(formatted, string(orig)))

	return nil
}
