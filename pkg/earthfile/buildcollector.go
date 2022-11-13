package earthfile

import (
	"github.com/earthly/earthly/ast/spec"
)

type BuildCmd struct {
	Line   string // Earthfile syntax of this command
	Base   string // Path to the directory in which this command resides
	Target string // Path to the Earthly target this command builds
}

type buildCmdCollector struct {
	UnimplementedStmtVisitor
	ef     *Earthfile
	builds []BuildCmd
}

func CollectBuildCommands(f *Earthfile, t *spec.Target) []BuildCmd {
	visitor := &buildCmdCollector{ef: f}
	WalkRecipe(t.Recipe, visitor)
	return visitor.builds
}

func (v *buildCmdCollector) VisitCommand(c spec.Command) {
	switch c.Name {
	case "BUILD":
		v.builds = append(v.builds, BuildCmd{
			Line:   cmdRepr(c),
			Base:   v.ef.Dir,
			Target: c.Args[0],
		})
	}
}
