package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/dorfire/heavenly/pkg/godepresolver"
)

func printCopyCommandsForGoDeps(cCtx *cli.Context) error {
	pkgPath := cCtx.Args().First()
	if pkgPath == "" {
		return errors.New("missing Go package argument")
	}

	r, err := godepresolver.New(cCtx.String("go-mod-dir"), logger)
	if err != nil {
		return err
	}

	includeTransitive := cCtx.Bool("include-transitive")
	copies, testCopies, err := r.ResolveImportsToCopyCommands(pkgPath, includeTransitive)
	if err != nil {
		return err
	}

	docArg := " "
	if includeTransitive {
		docArg = " --include-transitive "
	}
	docCmd := fmt.Sprintf("heavenly gocopies%s%s", docArg, pkgPath)

	fmt.Printf("\n# Go imports (generated with `%s`)\n", docCmd)
	fmt.Println(godepresolver.FormatCopyCommands(copies))

	fmt.Printf("\n# Go test imports (generated with `%s`)\n", docCmd)
	fmt.Println(godepresolver.FormatCopyCommands(testCopies))
	fmt.Println()

	return nil
}
