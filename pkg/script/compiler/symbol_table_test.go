package compiler

import "testing"

func TestDefine(t *testing.T) {
	want := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != want["a"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "a", want["a"], a)
	}

	b := global.Define("b")
	if b != want["b"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "b", want["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")
	if c != want["c"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "c", want["c"], c)
	}

	d := firstLocal.Define("d")
	if d != want["d"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "d", want["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != want["e"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "e", want["e"], e)
	}

	f := secondLocal.Define("f")
	if f != want["f"] {
		t.Errorf("symbol %q: want=%#v, got=%#v", "f", want["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	wantSymbols := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, want := range wantSymbols {
		got, ok := global.Resolve(want.Name)
		if !ok {
			t.Errorf("name %q not resolvable", got.Name)
			continue
		}

		if got != want {
			t.Errorf("expected %q to resolve to %#v, but got %#v", want.Name, want, got)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expectedSymbols := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, want := range expectedSymbols {
		got, ok := local.Resolve(want.Name)
		if !ok {
			t.Errorf("name %q not resolvable", want.Name)
			continue
		}

		if got != want {
			t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table       *SymbolTable
		wantSymbols []Symbol
	}{
		{
			table: firstLocal,
			wantSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			table: secondLocal,
			wantSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, want := range tt.wantSymbols {
			got, ok := tt.table.Resolve(want.Name)
			if !ok {
				t.Errorf("name %q not resolvable", want.Name)
				continue
			}

			if got != want {
				t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
			}
		}
	}
}

func TestResolveCurrentScopeGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	wantSymbols := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, want := range wantSymbols {
		got, ok := global.ResolveCurrentScope(want.Name)
		if !ok {
			t.Errorf("name %q not resolvable", got.Name)
			continue
		}

		if got != want {
			t.Errorf("expected %q to resolve to %#v, but got %#v", want.Name, want, got)
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(global)

	wantSymbols := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range wantSymbols {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, want := range wantSymbols {
			got, ok := table.Resolve(want.Name)
			if !ok {
				t.Errorf("name %q not resolvable", want.Name)
				continue
			}

			if got != want {
				t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		wantSymbols     []Symbol
		wantFreeSymbols []Symbol
	}{
		{
			table: firstLocal,
			wantSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			wantFreeSymbols: []Symbol{},
		},
		{
			table: secondLocal,
			wantSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			wantFreeSymbols: []Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, want := range tt.wantSymbols {
			got, ok := tt.table.Resolve(want.Name)
			if !ok {
				t.Errorf("name %q not resolvable", want.Name)
				continue
			}

			if got != want {
				t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
			}
		}

		if gotLen, wantLen := len(tt.table.freeSymbols), len(tt.wantFreeSymbols); gotLen != wantLen {
			t.Errorf("wrong number of free symbols. want=%d, got=%d", wantLen, gotLen)
			continue
		}

		for i, want := range tt.wantFreeSymbols {
			if got := tt.table.freeSymbols[i]; got != want {
				t.Errorf("wrong free symbol. want=%+v, got=%+v", want, got)
			}
		}
	}
}

func TestResolveUnresolvable(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	wantSymbols := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, want := range wantSymbols {
		got, ok := secondLocal.Resolve(want.Name)
		if !ok {
			t.Errorf("name %q not resolvable", want.Name)
			continue
		}

		if got != want {
			t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
		}
	}

	wantUnresolvable := []string{"b", "d"}

	for _, name := range wantUnresolvable {
		if _, ok := secondLocal.Resolve(name); ok {
			t.Errorf("name %q resolved, but was expected not to", name)
		}
	}
}

func TestResolveCurrentScopeUnresolvable(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table       *SymbolTable
		wantSymbols []Symbol
	}{
		{
			table: firstLocal,
			wantSymbols: []Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			table: secondLocal,
			wantSymbols: []Symbol{
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, want := range tt.wantSymbols {
			got, ok := tt.table.Resolve(want.Name)
			if !ok {
				t.Errorf("name %q not resolvable", want.Name)
				continue
			}

			if got != want {
				t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
			}
		}
	}

	wantUnresolvable := []struct {
		table *SymbolTable
		names []string
	}{
		{
			table: firstLocal,
			names: []string{"a", "b", "e", "f", "g", "h"},
		},
		{
			table: secondLocal,
			names: []string{"a", "b", "c", "d", "g", "h"},
		},
	}

	for _, tt := range wantUnresolvable {
		for _, name := range tt.names {
			if _, ok := tt.table.ResolveCurrentScope(name); ok {
				t.Errorf("name %q resolved, but was expected not to", name)
			}
		}
	}
}

func TestDefineAndResolveFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	want := Symbol{Name: "a", Scope: FunctionScope, Index: 0}

	got, ok := global.Resolve(want.Name)
	if !ok {
		t.Fatalf("function name %q not resolvable", want.Name)
	}

	if got != want {
		t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
	}
}

func TestDefineAndResolveCurrentScopeFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	want := Symbol{Name: "a", Scope: FunctionScope, Index: 0}

	got, ok := global.ResolveCurrentScope(want.Name)
	if !ok {
		t.Fatalf("function name %q not resolvable", want.Name)
	}

	if got != want {
		t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
	}
}

func TestShadowingFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	want := Symbol{Name: "a", Scope: GlobalScope, Index: 0}

	got, ok := global.Resolve(want.Name)
	if !ok {
		t.Fatalf("function name %q not resolvable", want.Name)
	}

	if got != want {
		t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
	}
}

func TestShadowingFunctionNameCurrentScope(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	want := Symbol{Name: "a", Scope: GlobalScope, Index: 0}

	got, ok := global.ResolveCurrentScope(want.Name)
	if !ok {
		t.Fatalf("function name %q not resolvable", want.Name)
	}

	if got != want {
		t.Errorf("expected %q to resolve to %+v, but got %+v", want.Name, want, got)
	}
}
