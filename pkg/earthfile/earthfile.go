package earthfile

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/earthly/earthly/ast"
	"github.com/earthly/earthly/ast/spec"
	"github.com/samber/lo"
)

const (
	earthfileName = "Earthfile"
)

type Earthfile struct {
	Dir, Path string
	Spec      spec.Earthfile
	Globals   map[string]string
}

func Parse(path string) (*Earthfile, error) {
	// ast.Parse seems to not do anything of importance with ctx, so passing context.Background()
	a, err := ast.Parse(context.Background(), path, false)
	if err != nil {
		return nil, err
	}

	// Look for ARG commands in the base recipe
	baseArgs, err := parseArgs(a.BaseRecipe)
	if err != nil {
		return nil, err
	}

	return &Earthfile{
		Dir:     filepath.Dir(path),
		Path:    path,
		Spec:    a,
		Globals: baseArgs,
	}, nil
}

func parseArgs(recipe spec.Block) (map[string]string, error) {
	res := map[string]string{}
	for _, s := range recipe {
		if s.Command != nil && s.Command.Name == "ARG" {
			if len(s.Command.Args) != 3 || s.Command.Args[1] != "=" {
				return nil, fmt.Errorf("earthfile: unexpected ARG syntax: %v", s.Command.Args)
			}
			res[s.Command.Args[0]] = s.Command.Args[2]
		}
	}
	return res, nil
}

func ParseTarget(path string) (*Earthfile, *spec.Target, error) {
	pathParts := strings.SplitN(path, "+", 2)
	//earthDir, tName, hasPlus := strings.Cut(path, "+"); hasPlus {
	if len(pathParts) != 2 {
		return nil, nil, fmt.Errorf("earthfile: invalid Earthly target '%s'", path)
	}
	return relTarget(pathParts[0], pathParts[1])
}

func (f *Earthfile) TargetNames() []string {
	return lo.Map(f.Spec.Targets, func(t spec.Target, _ int) string { return t.Name })
}

func (f *Earthfile) ExpandArgs(s string) string {
	if strings.ContainsRune(s, '$') {
		// TODO: support target-specific args?
		for globalName, globalVal := range f.Globals {
			s = strings.ReplaceAll(s, "$"+globalName, globalVal)
		}
	}
	return s
}

// Target looks up an Earthly target.
// path may be a simple name, like "src", or a qualified path relative to the current Earthfile, like "../+src".
func (f *Earthfile) Target(path string) (*Earthfile, *spec.Target, error) {
	if earthDir, tName, hasPlus := strings.Cut(path, "+"); hasPlus && earthDir != "" {
		return relTarget(filepath.Join(f.Dir, earthDir), tName)
	}
	t, err := f.localTarget(path)
	return f, t, err
}

func relTarget(dir, target string) (*Earthfile, *spec.Target, error) {
	// Look up Earthfile in target dir
	targetFile := path.Join(dir, earthfileName)
	if _, err := os.Lstat(targetFile); err != nil {
		return nil, nil, fmt.Errorf("earthfile: could not stat '%s': %w", targetFile, err.(*os.PathError).Err)
	}

	ef, err := Parse(targetFile)
	if err != nil {
		return nil, nil, fmt.Errorf("earthfile: could not parse '%s': %w", targetFile, err)
	}

	t, err := ef.localTarget(target)
	return ef, t, err
}

func (f *Earthfile) localTarget(name string) (*spec.Target, error) {
	name = strings.TrimPrefix(name, "+")

	for _, t := range f.Spec.Targets {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("earthfile: local target '%s' not found. available targets: %v", name, strings.Join(f.TargetNames(), ", "))
}

func cmdRepr(c spec.Command) string {
	return fmt.Sprintf("%s %s", c.Name, strings.Join(c.Args, " "))
}
