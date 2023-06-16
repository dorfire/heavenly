package godepresolver

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dorfire/heavenly/pkg/earthfilefmt"
	"github.com/earthly/earthly/ast/spec"
	"github.com/earthly/earthly/conslogging"
	"github.com/samber/lo"

	"github.com/dorfire/heavenly/pkg/earthdir"
	"github.com/dorfire/heavenly/pkg/goparse"
)

const (
	selfCopyCmd = "COPY --dir +src/* ."
)

var (
	dirsToSkip = mapset.NewThreadUnsafeSet("testdata", "node_modules")
)

type GoDepResolver struct {
	goModRoot string
	projRoot  string
	goModFile *modfile.File
	log       conslogging.ConsoleLogger
}

func New(goModDir string, log conslogging.ConsoleLogger) (*GoDepResolver, error) {
	goModRoot, err := filepath.Abs(goModDir)
	if err != nil {
		return nil, err
	}

	projRoot, err := earthdir.InOrAbove(goModRoot, "/", false)
	if err != nil {
		return nil, fmt.Errorf("could not find top-most Earthfile from %s: %w", goModRoot, err)
	}
	log.DebugPrintf("Detected project root: %s", projRoot)

	goModFile, err := goparse.ModFile(goModRoot)
	if err != nil {
		return nil, fmt.Errorf("could not parse go.mod in %s: %w", goModRoot, err)
	}

	return &GoDepResolver{goModRoot, projRoot, goModFile, log}, nil
}
func (r *GoDepResolver) ResolveImportsToCopyCommands(
	pkgPath string, includeTransitive bool,
) (imports []spec.Command, testOnlyImports []spec.Command, err error) {
	err = filepath.Walk(pkgPath, func(p string, inf fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("could not access path %q: %v\n", p, err)
		}

		if !inf.IsDir() {
			return nil
		}

		if dirsToSkip.Contains(inf.Name()) {
			return filepath.SkipDir
		}

		r.log.DebugPrintf("Resolving %q", p)
		i, t, err := r.resolve(p, includeTransitive)
		if err != nil {
			// This could be ok; some dirs only contain other packages
			r.log.DebugPrintf("could not resolve imports for pkg %q: %v\n", p, err)
			return nil
		}

		imports = append(imports, i...)
		testOnlyImports = append(testOnlyImports, t...)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return sortAndDedupCopies(imports), sortAndDedupCopies(testOnlyImports), nil
}

// ResolveImportsToCopyCommands resolves internal Go imports in a given package to the COPY commands they probably depend on.
// This is a poor man's Gazelle - it presumes every imported directory has an Earthfile in it, or above it, with a +src
// target, but doesn't actually ensure it includes the required Go source files or that it even exists.
// If includeTransitive is true, this func will recursively collect all transitive imports from non-test deps.
func (r *GoDepResolver) resolve(
	pkgPath string, includeTransitive bool,
) (imports []spec.Command, testOnlyImports []spec.Command, err error) {
	pkgImports, err := goparse.PackageImports(pkgPath, false)
	if err != nil {
		return nil, nil, err
	}

	if includeTransitive {
		// for each module-internal package, recursively resolve its non-test imports
		for dep := range pkgImports.Iter() {
			if isInternal := strings.HasPrefix(dep, r.goModFile.Module.Mod.Path); !isInternal {
				continue
			}
			r.log.DebugPrintf("Resolving transitive import %q from %q", dep, pkgPath)
			depImports, err := goparse.PackageImports(r.pkgGoPathToFSDir(dep), false)
			if err != nil {
				return nil, nil, err
			}
			pkgImports = pkgImports.Union(depImports)
		}
	}

	testPkgImports, _ := goparse.PackageImports(pkgPath, true)
	if testPkgImports != nil {
		testPkgImports = testPkgImports.Difference(pkgImports)
	}

	pkgPathAbs, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, nil, err
	}

	r.log.DebugPrintf("Detected Go imports:")
	r.log.DebugPrintf(strings.Join(pkgImports.ToSlice(), "\n"))

	imports, err = r.resolveCopyCommandsForImports(pkgImports, pkgPathAbs)
	if err != nil {
		return nil, nil, err
	}

	if testPkgImports != nil {
		testOnlyImports, err = r.resolveCopyCommandsForImports(testPkgImports, pkgPathAbs)
		if err != nil {
			return nil, nil, err
		}
	}

	return
}

func FormatCopyCommands(cmds []spec.Command) string {
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

func (r *GoDepResolver) resolveCopyCommandsForImports(pkgImports mapset.Set[string], pkgPathAbs string) ([]spec.Command, error) {
	var res []spec.Command
	modPath := r.goModFile.Module.Mod.Path
	for imp := range pkgImports.Iter() {
		impDir := r.pkgGoPathToFSDir(imp)
		if !strings.HasPrefix(imp, modPath) || impDir == pkgPathAbs {
			continue // Ignore non-internal imports and same-dir imports
		}

		cmd, err := r.resolveCopyCommandForGoImport(pkgPathAbs, impDir)
		if err != nil {
			return nil, err
		}

		res = append(res, cmd)
	}
	return res, nil
}

func (r *GoDepResolver) resolveCopyCommandForGoImport(importerDir, importeeDir string) (spec.Command, error) {
	resolvedEarthdir, err := earthdir.InOrAbove(importeeDir, r.goModRoot, true) // TODO: cache here to save some I/O?
	if err != nil {
		return spec.Command{}, err
	}

	if importerDir == importeeDir || /* Hack: */ importerDir == importeeDir+"/cmd" {
		return spec.Command{Name: "COPY", Args: []string{"--dir", "+src/*", "."}}, nil
	}

	src, err := formatCopyCmdSrc(r.projRoot, importeeDir, resolvedEarthdir)
	if err != nil {
		return spec.Command{}, err
	}
	dst := replaceRootWithTopArg(importeeDir, r.goModRoot)
	return spec.Command{Name: "COPY", Args: []string{"--dir", src, dst}}, nil
}

func (r *GoDepResolver) pkgGoPathToFSDir(importPath string) string {
	return strings.Replace(importPath, r.goModFile.Module.Mod.Path, r.goModRoot, 1)
}

func sortAndDedupCopies(copies []spec.Command) []spec.Command {
	sort.SliceStable(copies, func(i, j int) bool {
		cmdI, cmdJ := earthfilefmt.FormatArgs(copies[i].Args), earthfilefmt.FormatArgs(copies[j].Args)
		return strings.Compare(cmdI, cmdJ) < 0
	})
	return lo.UniqBy[spec.Command, string](copies, func(c spec.Command) string {
		return earthfilefmt.FormatArgs(c.Args)
	})
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
