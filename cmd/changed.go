package main

import (
	"fmt"
	"strings"

	"github.com/dorfire/heavenly/pkg/gitutil"
	cli "github.com/urfave/cli/v2"
)

func failIfTargetUnchanged(ctx *cli.Context) error {
	targetPath := ctx.Args().First()

	targetDir, _, _ := strings.Cut(targetPath, "+")
	diff, err := gitDiff(ctx, targetDir)
	if err != nil {
		return err
	}

	changed, err := targetInputsChanged(ctx, diff, targetPath)
	if err != nil {
		return err
	}

	if !changed {
		return cli.Exit(fmt.Errorf("Earthly target %s has no input changes", targetPath), 1)
	}

	logger.Printf("🌍 Earthly target '%s' has changed inputs", targetPath)
	return nil
}

func targetInputsChanged(_ *cli.Context, repoChanges gitutil.ChangeSet, targetPath string) (bool, error) {
	targetInputs, err := analyzeTargetDeps(targetPath)
	if err != nil {
		return false, err
	}

	logger.DebugPrintf("+ Added files: %v", repoChanges.Added)
	logger.DebugPrintf("/ Modified files: %v", repoChanges.Modified)
	logger.DebugPrintf("- Deleted files: %v", repoChanges.Deleted)

	intersection := targetInputs.Intersect(repoChanges.All())
	logger.DebugPrintf("🔥 Changed inputs: %v", intersection)

	return intersection.Cardinality() > 0, nil
}
