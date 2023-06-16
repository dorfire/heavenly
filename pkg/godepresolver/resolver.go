package godepresolver

import (
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"io/fs"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dorfire/heavenly/pkg/earthdir"
	"github.com/dorfire/heavenly/pkg/goparse"
	"github.com/earthly/earthly/ast/spec"
	"github.com/earthly/earthly/conslogging"
)

var (
	dirsToSkip = mapset.NewThreadUnsafeSet("testdata", "node_modules")
)

type pkgImports map[string]bool // path -> is transitive

type GoDepResolver struct {
	goModRoot      string
	projRoot       string
	goModFile      *modfile.File
	log            conslogging.ConsoleLogger
	pkgImportCache map[string]pkgImports // path -> imports
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

	return &GoDepResolver{
		goModRoot,
		projRoot,
		goModFile,
		log,
		map[string]pkgImports{},
	}, nil
}

// ResolveImportsToCopyCommands resolves internal Go imports in a given package to the COPY commands they probably depend on.
// This is a poor man's Gazelle - it presumes every imported directory has an Earthfile in it, or above it, with a +src
// target, but doesn't actually ensure it includes the required Go source files or that it even exists.
// If includeTransitive is true, this func will recursively collect all transitive imports from non-test deps.
func (r *GoDepResolver) ResolveImportsToCopyCommands(
	pkgPath string, includeTransitive bool,
) (importCopyCmds []spec.Command, testOnlyCopyCmds []spec.Command, err error) {
	pkgPathAbs, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, nil, err
	}

	goImports := pkgImports{}
	goTestImports := pkgImports{}

	err = filepath.WalkDir(pkgPath, func(p string, inf fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not access path %q: %v\n", p, err)
		}

		if !inf.IsDir() {
			return nil
		}

		if dirsToSkip.Contains(inf.Name()) {
			return filepath.SkipDir
		}

		r.log.DebugPrintf("Resolving Go pkg dir %q", p)
		i, t, err := r.resolve(p, includeTransitive)
		if err != nil {
			// This could be ok; some dirs only contain other packages
			r.log.DebugPrintf("Could not resolve imports for pkg %q: %v\n", p, err)
			return nil
		}

		extend(goImports, i)
		extend(goTestImports, t)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	//r.log.DebugPrintf("*** Detected Go imports for %q: ***", pkgPath)
	//r.log.DebugPrintf("- %s", strings.Join(goImports.ToSlice(), "\n- "))
	importCopyCmds, err = r.resolveCopyCommandsForImports(goImports, pkgPathAbs)
	if err != nil {
		return nil, nil, err
	}

	if goTestImports != nil {
		maps.DeleteFunc(goTestImports, func(k string, _ bool) bool {
			_, ok := goImports[k]
			return ok
		})
		//r.log.DebugPrintf("*** Detected Go test imports for %q: ***", pkgPath)
		//r.log.DebugPrintf("- %s", strings.Join(goTestImports.ToSlice(), "\n- "))
		testOnlyCopyCmds, err = r.resolveCopyCommandsForImports(goTestImports, pkgPathAbs)
		if err != nil {
			return nil, nil, err
		}
	}

	return
}

func (r *GoDepResolver) resolve(pkgPath string, includeTransitive bool) (imports, testOnlyImports pkgImports, err error) {
	imports, err = r.collectImports(pkgPath, false, includeTransitive)
	if err != nil {
		return
	}

	// Ignoring err because some dirs don't have test files
	testOnlyImports, _ = r.collectImports(pkgPath, true, includeTransitive)
	return
}

func (r *GoDepResolver) resolveCopyCommandsForImports(imps pkgImports, pkgPathAbs string) ([]spec.Command, error) {
	var res []spec.Command
	for imp, isTransitive := range imps {
		if !r.isInternalImport(imp) {
			continue // Ignore non-internal imports
		}

		cmd, err := r.resolveCopyCommandForGoImport(pkgPathAbs, r.pkgGoPathToFSDir(imp), isTransitive)
		if err != nil {
			return nil, err
		}

		res = append(res, cmd)
	}
	return res, nil
}

func (r *GoDepResolver) resolveCopyCommandForGoImport(importerDir, importeeDir string, transitive bool) (spec.Command, error) {
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

	args := []string{"--dir", src, replaceRootWithTopArg(importeeDir, r.goModRoot) + "/"}
	if transitive {
		args = append(args, "# indirect")
	}
	return spec.Command{Name: "COPY", Args: args}, nil
}

func (r *GoDepResolver) isInternalImport(importPath string) bool {
	return strings.HasPrefix(importPath, r.goModFile.Module.Mod.Path)
}

func (r *GoDepResolver) pkgGoPathToFSDir(importPath string) string {
	return strings.Replace(importPath, r.goModFile.Module.Mod.Path, r.goModRoot, 1)
}

func pkgImportCacheKey(pkgPath string, test bool) string {
	return fmt.Sprintf("%s-%t", pkgPath, test)
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

func extend[K comparable, V any](dst, src map[K]V) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}
