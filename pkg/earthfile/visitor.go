package earthfile

import (
	"github.com/earthly/earthly/ast/spec"
)

type StmtVisitor interface {
	VisitCommand(spec.Command)
	VisitWith(spec.WithStatement)
	VisitIf(spec.IfStatement)
	VisitFor(spec.ForStatement)
	VisitWait(spec.WaitStatement)
}

type UnimplementedStmtVisitor struct {
}

func WalkRecipe(recipe spec.Block, v StmtVisitor) {
	for _, stmt := range recipe {
		if stmt.Command != nil {
			v.VisitCommand(*stmt.Command)
		} else if stmt.With != nil {
			v.VisitWith(*stmt.With)
			v.VisitCommand(stmt.With.Command)
			WalkRecipe(stmt.With.Body, v)
		} else if stmt.If != nil {
			v.VisitIf(*stmt.If)
			WalkRecipe(stmt.If.IfBody, v)
			for _, b := range stmt.If.ElseIf {
				WalkRecipe(b.Body, v)
			}
			if stmt.If.ElseBody != nil {
				WalkRecipe(*stmt.If.ElseBody, v)
			}
		} else if stmt.For != nil {
			v.VisitFor(*stmt.For)
			WalkRecipe(stmt.For.Body, v)
		} else if stmt.Wait != nil {
			v.VisitWait(*stmt.Wait)
			WalkRecipe(stmt.Wait.Body, v)
		}
	}
}

func (v *UnimplementedStmtVisitor) VisitCommand(_ spec.Command) {
}

func (v *UnimplementedStmtVisitor) VisitWith(_ spec.WithStatement) {
}

func (v *UnimplementedStmtVisitor) VisitIf(_ spec.IfStatement) {
}

func (v *UnimplementedStmtVisitor) VisitFor(_ spec.ForStatement) {
}

func (v *UnimplementedStmtVisitor) VisitWait(_ spec.WaitStatement) {
}
