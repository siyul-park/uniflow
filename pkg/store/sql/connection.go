package sql

import (
	"fmt"
	"github.com/araddon/gou"
	"github.com/araddon/qlbridge/exec"
	"github.com/araddon/qlbridge/expr"
	"github.com/araddon/qlbridge/lex"
	"github.com/araddon/qlbridge/plan"
	"github.com/araddon/qlbridge/schema"
	"github.com/araddon/qlbridge/value"
	"github.com/araddon/qlbridge/vm"
	"github.com/siyul-park/uniflow/pkg/store"
)

type connection struct {
	*exec.TaskBase
	store   store.Store
	table   *schema.Table
	filter  any
	options store.FindOptions
}

var _ schema.Conn = (*connection)(nil)
var _ plan.SourcePlanner = (*connection)(nil)
var _ exec.ExecutorSource = (*connection)(nil)

func (c *connection) WalkSourceSelect(planner plan.Planner, source *plan.Source) (plan.Task, error) {
	if len(source.Custom) == 0 {
		source.Custom = make(gou.JsonHelper)
	}

	source.Conn = c
	source.SourceExec = true

	if source.Proj == nil {
		source.Proj = plan.NewProjectionInProcess(source.Stmt.Source).Proj
	}

	c.TaskBase = exec.NewTaskBase(source.Context())
	c.options.Limit = source.Stmt.Source.Limit
	c.options.Skip = source.Stmt.Source.Offset

	if source.Stmt.Source.Where != nil {
		if _, err := c.walkNode(source.Stmt.Source.Where.Expr, &c.filter); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (c *connection) WalkExecSource(source *plan.Source) (exec.Task, error) {
	csr, err := c.store.Find(source.Context(), c.filter, c.options)
	if err != nil {
		return nil, err
	}
	return &cursor{
		TaskBase: exec.NewTaskBase(source.Context()),
		cursor:   csr,
	}, nil
}

func (c *connection) Close() error {
	return nil
}

func (c *connection) walkNode(node expr.Node, q *any) (value.Value, error) {
	switch n := node.(type) {
	case *expr.NumberNode, *expr.StringNode:
		val, ok := vm.Eval(nil, n)
		if !ok {
			return nil, fmt.Errorf("could not evaluate: %v", n.String())
		}
		return val, nil
	case *expr.BinaryNode:
		switch n.Operator.T {
		case lex.TokenLogicAnd, lex.TokenLogicOr:
			var lhq any
			if _, err := c.walkNode(n.Args[0], &lhq); err != nil {
				return nil, err
			}
			var rhq any
			if _, err := c.walkNode(n.Args[1], &rhq); err != nil {
				return nil, err
			}
			if n.Operator.T == lex.TokenLogicAnd {
				*q = map[string]any{"$and": []any{lhq, rhq}}
			} else {
				*q = map[string]any{"$or": []any{lhq, rhq}}
			}
			return nil, nil
		case lex.TokenEqual, lex.TokenEqualEqual, lex.TokenNE, lex.TokenGE, lex.TokenGT, lex.TokenLE, lex.TokenLT:
			lhval, ok := n.Args[0].(*expr.IdentityNode)
			if !ok {
				return nil, fmt.Errorf("invalid left hand side: %v", n.Args[0].String())
			}
			rhval, ok := vm.Eval(nil, n.Args[1])
			if !ok {
				return nil, fmt.Errorf("invalid right hand side: %v", n.Args[1].String())
			}
			if n.Operator.T == lex.TokenEqual || n.Operator.T == lex.TokenEqualEqual {
				*q = map[string]any{lhval.String(): rhval.Value()}
			} else if n.Operator.T == lex.TokenNE {
				*q = map[string]any{lhval.String(): map[string]any{"$ne": rhval.Value()}}
			} else if n.Operator.T == lex.TokenGE {
				*q = map[string]any{lhval.String(): map[string]any{"$gte": rhval.Value()}}
			} else if n.Operator.T == lex.TokenGT {
				*q = map[string]any{lhval.String(): map[string]any{"$gt": rhval.Value()}}
			} else if n.Operator.T == lex.TokenLE {
				*q = map[string]any{lhval.String(): map[string]any{"$lte": rhval.Value()}}
			} else if n.Operator.T == lex.TokenLT {
				*q = map[string]any{lhval.String(): map[string]any{"$lt": rhval.Value()}}
			}
			return nil, nil
		}
		return nil, nil
	}
	return nil, nil
}
