package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/schollz/progressbar/v3"
	cli "github.com/urfave/cli/v2"

	"github.com/dorfire/heavenly/pkg/earthfile"
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
	progBar := newAnalysisProgressBar(len(buildsInTarget))

	var targetsWithChanges []string
	lop.ForEach(buildsInTarget, func(t earthfile.BuildCmd, _ int) {
		// TODO: cache resolved deps across targets?
		if changed := lo.Must(targetInputsChanged(ctx, repoChanges, t.Target)); changed {
			targetsWithChanges = append(targetsWithChanges, t.Target)
		}
		_ = progBar.Add(1)
	})

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

func listDependentBuildsForInputs(ctx *cli.Context) error {
	tPath := ctx.Args().First()
	if tPath == "" {
		return errors.New("missing Earthly target argument")
	}

	inputPaths := ctx.Args().Tail()
	if len(inputPaths) == 0 {
		return errors.New("missing input file paths")
	}

	ef, target, err := earthfile.ParseTarget(tPath)
	if err != nil {
		return err
	}

	buildsInTarget := earthfile.CollectBuildCommands(ef, target)
	progBar := newAnalysisProgressBar(len(buildsInTarget))

	stopTimer := timer(fmt.Sprintf("Analyzing %d targets", len(buildsInTarget)))
	var dependents []string
	lop.ForEach(buildsInTarget, func(t earthfile.BuildCmd, _ int) {
		// TODO: cache resolved deps across targets?
		buildInputs := lo.Must(analyzeTargetDeps(t.Target))
		if buildInputs.Contains(inputPaths...) {
			dependents = append(dependents, t.Target)
		}
		_ = progBar.Add(1)
	})
	stopTimer()

	logger.PrintPhaseHeader(
		fmt.Sprintf("\n%d targets depend on inputs %v:", len(dependents), inputPaths), false, "")
	logger.Printf(strings.Join(dependents, "\n"))
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

func newAnalysisProgressBar(targets int) *progressbar.ProgressBar {
	ctor := progressbar.Default
	if os.Getenv("CI") == "true" {
		ctor = progressbar.DefaultSilent
	}
	return ctor(int64(targets), "Analyzing BUILD commands")
}
