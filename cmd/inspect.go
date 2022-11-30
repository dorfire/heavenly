package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dorfire/heavenly/pkg/earthfile"
	"github.com/earthly/earthly/ast/spec"
	"github.com/earthly/earthly/util/fileutil"
	"github.com/samber/lo"
	"github.com/tufin/asciitree"
	cli "github.com/urfave/cli/v2"
)

func inspectTargetInputs(ctx *cli.Context) error {
	tPath := ctx.Args().First()
	if tPath == "" {
		return errors.New("missing Earthly target argument")
	}

	targetInputs, err := analyzeTargetDeps(tPath)
	if err != nil {
		return err
	}

	var displayTargetInputs string
	if ctx.Bool("pretty") {
		displayTargetInputs = prettyPathTree(targetInputs)
	} else {
		displayTargetInputs = strings.Join(targetInputs.ToSlice(), "\n")
	}

	logger.Printf(displayTargetInputs)
	return nil
}

// analyzeTargetDeps analyzes an Earthly target at the given path and returns the files it is assumed to depend on.
func analyzeTargetDeps(tPath string) (mapset.Set[string], error) {
	ef, target, err := earthfile.ParseTarget(tPath)
	if err != nil {
		return nil, err
	}

	copies := earthfile.CollectCopyCommands(ef, target)

	logger.DebugPrintf("Inspecting Earthfile @ %s", ef.Dir)
	debugPrintCopyCommands(target, copies)

	var targetInputs []string
	for _, cp := range copies {
		files, err := expandCopyCmd(ef, cp)
		if err != nil {
			return nil, err
		}
		targetInputs = append(targetInputs, files...)
	}
	return mapset.NewSet(targetInputs...), nil
}

func prettyPathTree(paths mapset.Set[string]) string {
	res := new(bytes.Buffer)
	t := asciitree.Tree{}
	for _, p := range paths.ToSlice() {
		t.Add(p)
	}
	t.Fprint(res, false, "")
	return res.String()
}

func debugPrintCopyCommands(target *spec.Target, copies []earthfile.CopyCmd) {
	logger.DebugPrintf("COPY commands in target '%s':\n", target.Name)
	for _, c := range copies {
		logger.DebugPrintf(" - %s (base: %s)", c.Line, c.Base)
	}
}

func expandCopyCmd(ef *earthfile.Earthfile, cp earthfile.CopyCmd) (res []string, err error) {
	// Each 'COPY' command either references an Earthly target, a simple path, or a glob pattern.

	// First, replace $ARG refs with their underlying value
	cp.From = ef.ExpandArgs(cp.From)
	cp.To = ef.ExpandArgs(cp.To)

	fsFrom := filepath.Join(ef.Dir, cp.From)

	if strings.Contains(cp.From, "+") { // If 'from' path looks like an Earthly target
		targetPath, _ := splitTargetFileSelector(cp.From)
		fromEarthfile, fromTarget, err := ef.Target(targetPath)
		if err != nil {
			return nil, fmt.Errorf("could not find target '%s': %w", cp.From, err)
		}

		targetCPs := earthfile.CollectCopyCommands(fromEarthfile, fromTarget)

		res = lo.FlatMap(targetCPs, func(c earthfile.CopyCmd, _ int) []string {
			return lo.Must(expandCopyCmd(fromEarthfile, c))
		})
	} else if strings.Contains(cp.From, "*") {
		res, err = filepath.Glob(fsFrom)
		if err != nil {
			return nil, fmt.Errorf("could not glob pattern '%s' in '%s': %w", cp.From, ef.Dir, err)
		}
		res, err = expandGlobMatches(res)
		if err != nil {
			return nil, fmt.Errorf("could not expand glob matches in '%s': %w", ef.Dir, err)
		}
	} else if fileutil.DirExistsBestEffort(fsFrom) {
		res, err = filesInDir(fsFrom)
		if err != nil {
			return nil, fmt.Errorf("could not list files '%s': %w", ef.Dir, err)
		}
	} else {
		res = []string{filepath.Join(cp.Base, cp.From)}
	}

	logger.DebugPrintf("[%s] Expanded `%s` to =>\n  %v", ef.Dir, cp.Line, res)

	return res, nil
}

func expandGlobMatches(matches []string) ([]string, error) {
	res := make([]string, 0, len(matches))
	for _, m := range matches {
		if !fileutil.DirExistsBestEffort(m) {
			res = append(res, m)
			continue
		}

		logger.DebugPrintf("Expanding files in dir %s", m)
		files, err := filesInDir(m)
		if err != nil {
			return nil, err
		}

		res = append(res, files...)
	}
	return res, nil
}

// '+src/bla' -> 'src', 'bla'
func splitTargetFileSelector(path string) (target, selector string) {
	plusPos := strings.IndexRune(path, '+')

	pathAfterPlus := path[plusPos:]
	slashPos := strings.IndexRune(pathAfterPlus, '/')

	return path[:plusPos+slashPos], pathAfterPlus[slashPos:]
}

func filesInDir(d string) (res []string, err error) {
	err = filepath.WalkDir(d, func(path string, ent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !ent.IsDir() {
			res = append(res, path)
		}
		return nil
	})
	return res, err
}
