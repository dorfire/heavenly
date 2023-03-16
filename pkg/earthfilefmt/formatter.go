package earthfilefmt

import (
	"fmt"
	"strings"

	"github.com/earthly/earthly/ast/spec"
)

const (
	indentationPrefix = "    " // four spaces
)

func Format(ef spec.Earthfile) string {
	w := new(strBuilder)

	w.Write(FormatCmd("VERSION", ef.Version.Args))
	w.WriteNl()
	w.WriteNl()

	formatRecipe(w, 0, ef.BaseRecipe)
	w.WriteNl()

	for _, r := range ef.Targets {
		w.Write(r.Name)
		w.WriteRune(':')
		w.WriteNl()
		formatRecipe(w, 1, r.Recipe)
		w.WriteNl()
	}

	// TODO: also format ef.UserCommands

	return w.String()
}

func FormatCmd(cmd string, args []string) string {
	return fmt.Sprintf("%s %s", cmd, formatArgs(args))
}

// TODO: recurse
// TODO: support '#' comments
func formatRecipe(w *strBuilder, indent int, r spec.Block) {
	for _, c := range r {
		if c.Command == nil {
			panic("unimplemented: non-command tokens in recipe")
		}

		if indent != 0 {
			w.Write(strings.Repeat(indentationPrefix, indent))
		}

		w.Write(FormatCmd(c.Command.Name, c.Command.Args))
		w.WriteNl()
	}
}

func formatArgs(args []string) string {
	// Some args deserve special treatment, such as '=' which is not preceded/followed by a space; e.g. "ENV X=Y".
	// Thus, a plain strings.Join isn't a good fit here
	w := new(strBuilder)
	for i, a := range args {
		w.Write(a)

		nextArgIsntEq := i < len(args)-1 && args[i+1] != "="
		if a != "=" && nextArgIsntEq {
			w.WriteRune(' ')
		}
	}
	return w.String()
}
