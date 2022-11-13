package earthfilefmt

import (
	"strings"

	"github.com/samber/lo"
)

// strBuilder is a convenience wrapper around strings.Builder that doesn't return an error.
// See https://stackoverflow.com/a/70388629
type strBuilder struct {
	strings.Builder
}

func (b *strBuilder) Write(s string) {
	_ = lo.Must(b.Builder.WriteString(s))
}

func (b *strBuilder) WriteRune(r rune) {
	_ = lo.Must(b.Builder.WriteRune(r))
}

func (b *strBuilder) WriteNl() {
	b.WriteRune('\n')
}
