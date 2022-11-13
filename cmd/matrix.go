package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dorfire/heavenly/pkg/earthfile"
	"github.com/samber/lo"
	cli "github.com/urfave/cli/v2"
)

func outputChangedChildBuilds(ctx *cli.Context) error {
	tPath := ctx.Args().First()
	if tPath == "" {
		return errors.New("missing Earthly target argument")
	}

	ef, target, err := earthfile.ParseTarget(tPath)
	if err != nil {
		return err
	}

	repoChanges, err := gitDiff(ctx, ef.Dir)
	if err != nil {
		return err
	}

	buildsInTarget := earthfile.CollectBuildCommands(ef, target)
	targetsWithChanges := []string{}
	for _, t := range buildsInTarget {
		// TODO: parallelize?
		// TODO: cache resolved deps across targets?
		if changed := lo.Must(targetInputsChanged(ctx, repoChanges, t.Target)); changed {
			targetsWithChanges = append(targetsWithChanges, t.Target)
		}
	}

	logger.DebugPrintf("Targets with changed inputs:")
	logger.DebugPrintf(strings.Join(targetsWithChanges, "\n"))

	if !ctx.Bool("json") {
		logger.Printf(strings.Join(targetsWithChanges, "\n"))
		return nil
	}

	jsonBytes, err := json.Marshal(targetsWithChanges)
	if err != nil {
		return err
	}
	logger.PrintBytes(jsonBytes)

	if ghOutputPath := os.Getenv("GITHUB_OUTPUT"); ghOutputPath != "" {
		logger.DebugPrintf("GitHub env detected; outputting to %s", ghOutputPath)
		return appendGitHubOutput(ghOutputPath, "targets", string(jsonBytes))
	}

	return nil
}

func appendGitHubOutput(path, name, val string) error {
	ghOutput, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = ghOutput.WriteString(fmt.Sprintf("%s=%s\n", name, val))
	return err
}
