package main

import (
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/urfave/cli/v2"
	"go/parser"
	"go/token"
	"golang.org/x/exp/maps"
	"io/fs"
	"strings"
)

const (
	wellKnownSrcTarget = "src" // Earthfiles are expected to contain this target
)

// resolveGoImports resolves Go imports in a given directory to the COPY commands they probably depend on.
// Poor man's Gazelle.
func resolveGoImports(ctx *cli.Context) error {
	pkgPath := ctx.Args().First()
	if pkgPath == "" {
		return errors.New("missing Go package argument")
	}

	pkgImports, err := parsePackageImports(pkgPath, false)
	if err != nil {
		return err
	}
	fmt.Println(pkgImports)

	logger.Printf(prettyPathTree(pkgImports))
	return nil
}

//
//func analyzeImports(path string) (mapset.Set[string], error) {
//	parser.ParseDir()
//}

func parsePackageImports(path string, test bool) (mapset.Set[string], error) {
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
			res.Add(pkgImport.Path.Value)
		}
	}
	return res, nil
}
