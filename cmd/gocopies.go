package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/earthly/earthly/ast/spec"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/dorfire/heavenly/pkg/earthfilefmt"
	"github.com/dorfire/heavenly/pkg/godepresolver"
)

const (
	selfCopyCmd = "COPY --dir +src/* ."
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
	fmt.Println(formatCopyCommands(copies))

	fmt.Printf("\n# Go test imports (generated with `%s`)\n", docCmd)
	fmt.Println(formatCopyCommands(testCopies))
	fmt.Println()

	return nil
}

func formatCopyCommands(cmds []spec.Command) string {
	cmdLines := lo.Map(cmds, func(c spec.Command, _ int) string { return earthfilefmt.FormatCmd(c.Name, c.Args) })
	sort.Strings(cmdLines)

	// "COPY --dir +src/* ." should be first
	selfCopyIdx := sort.SearchStrings(cmdLines, selfCopyCmd)
	if selfCopyIdx != 0 && selfCopyIdx < len(cmdLines) && cmdLines[selfCopyIdx] == selfCopyCmd {
		var tail []string
		if selfCopyIdx+1 < len(cmdLines) {
			tail = cmdLines[selfCopyIdx+1:]
		}
		cmdLines = append(cmdLines[:selfCopyIdx], tail...)
		cmdLines = append([]string{selfCopyCmd}, cmdLines...)
	}

	return strings.Join(cmdLines, "\n")
}
