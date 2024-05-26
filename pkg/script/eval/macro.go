package eval

import (
	"github.com/siyul-park/uniflow/pkg/script/ast"
	"github.com/siyul-park/uniflow/pkg/script/object"
)

// DefineMacros finds macro definitions in the program, saves them to a given environment and
// removes them from the AST.
func DefineMacros(program *ast.Program, env object.Environment) {
	defs := make([]int, 0)
	stmts := program.Statements

	for pos, stmt := range stmts {
		if isMacroDefinition(stmt) {
			addMacro(stmt, env)
			defs = append(defs, pos)
		}
	}

	for i := len(defs) - 1; i >= 0; i-- {
		pos := defs[i]
		program.Statements = append(stmts[:pos], stmts[pos+1:]...)
	}
}

func isMacroDefinition(node ast.Statement) bool {
	letStmt, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = letStmt.Value.(*ast.MacroLiteral)
	return ok
}

func addMacro(stmt ast.Statement, env object.Environment) {
	letStmt := stmt.(*ast.LetStatement)
	macroLit := letStmt.Value.(*ast.MacroLiteral)
	macro := &object.Macro{
		Parameters: macroLit.Parameters,
		Env:        env,
		Body:       macroLit.Body,
	}
	env.Set(letStmt.Name.Value, macro)
}

// ExpandMacros expands defined macros and replaces AST nodes with the result of macro expansion.
func ExpandMacros(program ast.Node, env object.Environment) ast.Node {
	modifier := func(node ast.Node) ast.Node {
		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(call, env)
		if !ok {
			return node
		}

		args := quoteArgs(call)
		evalEnv := extendMacroEnv(macro, args)

		quote, ok := Eval(macro.Body, evalEnv).(*object.Quote)
		if !ok {
			panic("we only support returning AST-nodes from macros")
		}

		return quote.Node
	}

	return ast.Modify(program, modifier)
}

func isMacroCall(call *ast.CallExpression, env object.Environment) (macro *object.Macro, ok bool) {
	ident, ok := call.Function.(*ast.Ident)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(ident.Value)
	if !ok {
		return nil, false
	}

	macro, ok = obj.(*object.Macro)
	return macro, ok
}

func quoteArgs(call *ast.CallExpression) []*object.Quote {
	args := make([]*object.Quote, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		args = append(args, &object.Quote{Node: arg})
	}
	return args
}

func extendMacroEnv(macro *object.Macro, args []*object.Quote) object.Environment {
	extended := object.NewEnclosedEnvironment(macro.Env)
	for i, param := range macro.Parameters {
		extended.Set(param.Value, args[i])
	}
	return extended
}
