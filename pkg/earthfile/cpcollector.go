package earthfile

import (
	"log"
	"strings"

	"github.com/earthly/earthly/ast/spec"
)

const (
	SentinelCopyCmdLine = "<sentinel>"
)

type CopyCmd struct {
	Line     string     // Earthfile syntax of this command
	File     *Earthfile // Earthfile where this command resides
	DirOpt   bool       // Whether --dir was passed to the command
	From, To string
}

type copyCmdCollector struct {
	UnimplementedStmtVisitor
	ef   *Earthfile
	cmds []CopyCmd
}

// CollectCopyCommands returns all COPY commands detected in the given Target.
// For simplicity, it also returns dummy `CopyCmd`s for detected Earthfile dependencies.
// TODO: separate COPY command collection from target dependency resolution.
func CollectCopyCommands(f *Earthfile, t *spec.Target) []CopyCmd {
	visitor := &copyCmdCollector{ef: f}
	WalkRecipe(t.Recipe, visitor)
	return visitor.cmds
}

func (v *copyCmdCollector) VisitCommand(c spec.Command) {
	switch c.Name {
	case "FROM":
		v.visitFromCommand(c)
	case "COPY":
		v.visitCopyCommand(c)
	case "BUILD":
		log.Printf("[WARNING] %s: skipping BUILD command parsing in copyCmdCollector", v.ef.Dir)
	}
}

func (v *copyCmdCollector) visitCopyCommand(c spec.Command) {
	if c.Name != "COPY" {
		panic("expected COPY command")
	}

	res := CopyCmd{
		Line: cmdRepr(c),
		File: v.ef,
	}

	if c.Args[0] == "--dir" {
		res.DirOpt = true
		c.Args = c.Args[1:]
	}

	if len(c.Args) == 2 {
		res.From, res.To = c.Args[0], c.Args[1]
		v.cmds = append(v.cmds, res)
		return
	}

	// For simplicity, split COPY commands with multiple input paths to multiple commands
	res.To = c.Args[len(c.Args)-1]
	for _, from := range c.Args[:len(c.Args)-1] {
		clone := res
		clone.From = from
		v.cmds = append(v.cmds, clone)
	}
}

func (v *copyCmdCollector) visitFromCommand(c spec.Command) {
	if c.Name != "FROM" {
		panic("expected FROM command")
	}

	// Avoid visiting remote image targets
	if !strings.ContainsRune(c.Args[0], '+') {
		return
	}

	ef, t, err := v.ef.Target(v.ef.ExpandArgs(c.Args[0]))
	if err != nil {
		panic(err)
	}

	// Add a fake COPY command for the Earthfile, to trick the pipeline into recognizing it as a dep.
	v.cmds = append(v.cmds, CopyCmd{Line: SentinelCopyCmdLine, File: v.ef, From: ef.Path})

	v.cmds = append(v.cmds, CollectCopyCommands(ef, t)...)
}
