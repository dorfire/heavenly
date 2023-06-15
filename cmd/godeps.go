package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/earthly/earthly/ast/spec"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/dorfire/heavenly/pkg/earthdir"
	"github.com/dorfire/heavenly/pkg/earthfilefmt"
	"github.com/dorfire/heavenly/pkg/goparse"
)

// resolveGoImports resolves internal Go imports in a given package to the COPY commands they probably depend on.
// This is a poor man's Gazelle - it presumes every imported directory has an Earthfile in it, or above it, with a +src
// target, but doesn't actually ensure it includes the required Go source files or that it even exists.
func resolveGoImports(ctx *cli.Context, testOnly bool) error {
	goModDirArg := ctx.String("go-mod-dir")
	goModRoot, err := filepath.Abs(goModDirArg)
	if err != nil {
		return err
	}

	goModFile, err := goparse.ModFile(goModDirArg)
	if err != nil {
		return fmt.Errorf("could not parse go.mod in %s: %w", goModDirArg, err)
	}

	pkgPath := ctx.Args().First()
	if pkgPath == "" {
		return errors.New("missing Go package argument")
	}

	pkgPathAbs, err := filepath.Abs(pkgPath)
	if err != nil {
		return err
	}

	pkgImports, err := goparse.PackageImports(pkgPath, false)
	if err != nil {
		return err
	}
	if testOnly {
		testPkgImports, err := goparse.PackageImports(pkgPath, true)
		if err != nil {
			return err
		}
		pkgImports = testPkgImports.Difference(pkgImports)
	}

	logger.DebugPrintf("Detected Go imports:")
	logger.DebugPrintf(strings.Join(pkgImports.ToSlice(), "\n"))

	projRoot, err := earthdir.InOrAbove(goModRoot, "/", false)
	if err != nil {
		return fmt.Errorf("could not find top-most Earthfile from %s: %w", goModRoot, err)
	}
	logger.DebugPrintf("Detected project root: %s", projRoot)

	var resolvedCopyCmds []spec.Command
	modPath := goModFile.Module.Mod.Path
	for imp := range pkgImports.Iter() {
		impDir := strings.Replace(imp, modPath, goModRoot, 1)
		if !strings.HasPrefix(imp, modPath) || impDir == pkgPathAbs {
			continue // Ignore non-internal imports and same-dir imports
		}

		cmd, err := resolveCopyCommandForGoImport(impDir, goModRoot, projRoot)
		if err != nil {
			return err
		}

		resolvedCopyCmds = append(resolvedCopyCmds, cmd)
	}

	cmdLines := lo.Map(resolvedCopyCmds, func(c spec.Command, _ int) string {
		return earthfilefmt.FormatCmd(c.Name, c.Args)
	})

	sort.Strings(cmdLines)
	logger.Printf(strings.Join(cmdLines, "\n"))

	return nil
}

func resolveCopyCommandForGoImport(impDir string, goModRoot string, projRoot string) (spec.Command, error) {
	resolvedEarthdir, err := earthdir.InOrAbove(impDir, goModRoot, true)
	if err != nil {
		return spec.Command{}, err
	}

	src, err := formatCopyCmdSrc(projRoot, impDir, resolvedEarthdir)
	if err != nil {
		return spec.Command{}, err
	}

	return spec.Command{
		Name: "COPY",
		Args: []string{"--dir", src, replaceRootWithTopArg(impDir, goModRoot)},
	}, nil
}

func formatCopyCmdSrc(projRoot, dirInProject, closestEarthdir string) (string, error) {
	relToClosestEarthfile, err := filepath.Rel(closestEarthdir, dirInProject)
	if err != nil {
		return "", err
	}

	srcDirWithTopArg := replaceRootWithTopArg(closestEarthdir, projRoot)

	if relToClosestEarthfile != "." {
		return fmt.Sprintf("%s/+src/%s/*", srcDirWithTopArg, relToClosestEarthfile), nil
	}
	return fmt.Sprintf("%s/+src/*", srcDirWithTopArg), nil
}

func replaceRootWithTopArg(s, projRoot string) string {
	return strings.Replace(s, projRoot, "$TOP", 1)
}
