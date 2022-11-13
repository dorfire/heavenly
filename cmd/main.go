package main

import (
	"fmt"
	"log"
	"os"

	"github.com/earthly/earthly/conslogging"
	"github.com/earthly/earthly/util/fileutil"
	cli "github.com/urfave/cli/v2"
)

var (
	logger      = conslogging.Current(conslogging.AutoColor, conslogging.DefaultPadding, conslogging.Info)
	gitDiffArgs = []cli.Flag{
		&cli.StringFlag{Name: "from-ref"},
		&cli.StringFlag{Name: "to-commit"},
	}
)

func main() {
	app := &cli.App{
		Name:        "heavenly",
		Usage:       "manages Earthly from above",
		Description: "heavenly is a CLI tool that formats, lints and analyzes Earthly repos and the Earthfiles in them.",
		Commands:    appCommands(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "chdir",
				Action: chdir,
			},
			&cli.BoolFlag{
				Name: "debug",
				Action: func(ctx *cli.Context, v bool) error {
					if v {
						logger = logger.WithLogLevel(conslogging.Debug)
					}
					return nil
				},
			},
		},
	}

	log.SetFlags(0)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func appCommands() []*cli.Command {
	return []*cli.Command{
		{
			// Draws inspiration from bazel-buildifier:
			// https://pkg.go.dev/github.com/bazelbuild/buildtools/buildifier
			Name:    "format",
			Aliases: []string{"fmt"},
			Usage:   "format Earthfiles in the current repo according to a set of rules",
			UsageText: "rules to be written:\n" +
				"- indentation in Earthfile blocks\n" +
				"- duplicate COPY commands",
			Action: formatEarthfile,
		},
		{
			// Draws inspiration from bazel-gazelle:
			// https://github.com/bazelbuild/bazel-gazelle
			Name:  "lint",
			Usage: "lint the current repo according to a set of rules",
			UsageText: "rules to be written:\n" +
				"- Earthfile target COPY command for a nonexistent path\n" +
				"- Go code that imports a Go `main` package\n" +
				"- Go package import without a corresponding COPY command\n" +
				"- Dart package import without a corresponding COPY command\n" +
				"- Go package directory without a corresponding +src target\n",
			Action: func(cCtx *cli.Context) error {
				fmt.Println("UNIMPLEMENTED")
				return nil
			},
		},
		{
			Name:   "changed",
			Usage:  "analyze a given Earthly target and exit with 0 if it has any changed input files. exit with 1 otherwise.",
			Action: failIfTargetUnchanged,
			Flags:  gitDiffArgs,
		},
		{
			// Draws inspiration from bazel-diff
			// https://github.com/Tinder/bazel-diff
			Name:   "matrix",
			Usage:  "analyze a given Earthly target and output the BUILD commands within it that need rebuilding",
			Action: outputChangedChildBuilds,
			Flags:  append([]cli.Flag{&cli.BoolFlag{Name: "json"}}, gitDiffArgs...),
		},
		{
			Name:      "inspect",
			Aliases:   []string{"inputs"},
			Usage:     "analyze a given Earthly target and show which source files it depends on",
			ArgsUsage: "target path",
			Action:    inspectTargetInputs,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "pretty"},
			},
		},
		{
			Name:  "dlearthly",
			Usage: "download an Earthly binary suitable for the current OS/arch and verify it against a given hash",
			Action: func(cCtx *cli.Context) error {
				fmt.Println("UNIMPLEMENTED")
				return nil
			},
		},
	}
}

func chdir(ctx *cli.Context, p string) error {
	ok, err := fileutil.DirExists(p)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("requested working dir '%s' does not exist", p)
	}
	return os.Chdir(p)
}
