package goparse

import (
	"fmt"
	"go/parser"
	"go/token"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	goModName = "go.mod"
)

func ModFile(dir string) (*modfile.File, error) {
	goModPath := filepath.Join(dir, goModName)

	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", goModPath, err)
	}

	goModFile, err := modfile.Parse(goModPath, goModContent, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", goModPath, err)
	}

	return goModFile, nil
}

func PackageImports(path string, test bool) (mapset.Set[string], error) {
	fset := token.NewFileSet()
	incl := func(i fs.FileInfo) bool {
		isTestFile := strings.HasSuffix(i.Name(), "_test.go")
		if test {
			return isTestFile
		}
		return strings.HasSuffix(i.Name(), ".go") && !isTestFile
	}
	pkgs, err := parser.ParseDir(fset, path, incl, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	// Make sure they are all in one package.
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no source-code package in directory %s", path)
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("multiple packages in directory %s: %v", path, maps.Keys(pkgs))
	}

	pkg := pkgs[maps.Keys(pkgs)[0]]
	res := mapset.NewThreadUnsafeSet[string]()
	for _, pkgFile := range pkg.Files {
		for _, pkgImport := range pkgFile.Imports {
			res.Add(strings.Trim(pkgImport.Path.Value, `"`))
		}
	}
	return res, nil
}
