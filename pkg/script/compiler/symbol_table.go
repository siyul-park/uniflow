package compiler

// SymbolScope represents a scope of symbols.
type SymbolScope string

const (
	// GlobalScope represents a global scope, i.e. top level context of a program.
	GlobalScope SymbolScope = "GLOBAL"
	// LocalScope represents a local scope, i.e. a function level context.
	LocalScope SymbolScope = "LOCAL"
	// BuiltinScope represents a scope for built-in functions.
	BuiltinScope SymbolScope = "BUILTIN"
	// FreeScope represents a scope for closures referencing free variables.
	FreeScope SymbolScope = "FREE"
	// FunctionScope represents a scope for self-referencing functions.
	FunctionScope SymbolScope = "FUNCTION"
)

// Symbol is a symbol defined in a scope with an identifier (name).
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable is a mapping table of identifiers (names) and defined symbols.
type SymbolTable struct {
	freeSymbols []Symbol

	outer *SymbolTable

	store   map[string]Symbol
	numDefs int
}

// NewSymbolTable creates a new symbol table.
func NewSymbolTable() *SymbolTable {
	return NewEnclosedSymbolTable(nil)
}

// NewEnclosedSymbolTable creates a new symbol table with an outer one.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		freeSymbols: make([]Symbol, 0),
		outer:       outer,
		store:       make(map[string]Symbol),
	}
}

// Define defines an identifier as a symbol in a scope.
func (s *SymbolTable) Define(name string) Symbol {
	scope := GlobalScope
	if s.hasOuter() {
		scope = LocalScope
	}

	sym := s.define(name, scope, s.numDefs)
	s.numDefs++
	return sym
}

// DefineBuiltin defines a built-in function with `name` at the `index`.
func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	return s.define(name, BuiltinScope, index)
}

// DefineFunctionName defines a built-in function with `name` at the `index`.
func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	return s.define(name, FunctionScope, 0)
}

// defineFree defines a free symbol based on the `original` one.
func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.freeSymbols = append(s.freeSymbols, original)

	return s.define(original.Name, FreeScope, len(s.freeSymbols)-1)
}

func (s *SymbolTable) define(name string, scope SymbolScope, index int) Symbol {
	sym := Symbol{Name: name, Scope: scope, Index: index}
	s.store[name] = sym
	return sym
}

// Resolve resolves an identifier and returns a defined symbol and `true` if any.
// If the identifier is not found anywhere within a chain of symbol tables, it returns an empty
// symbol and `false`.
func (s *SymbolTable) Resolve(name string) (sym Symbol, exists bool) {
	if sym, exists = s.store[name]; exists || !s.hasOuter() {
		return sym, exists
	}

	sym, exists = s.outer.Resolve(name)
	if exists && (sym.Scope == LocalScope || sym.Scope == FreeScope) {
		// Define an outer local or free variable as a free variable in the current scope
		sym = s.defineFree(sym)
	}
	return sym, exists
}

// ResolveCurrentScope resolves an identifier within the current scope and returns a defined
// symbol and `true` if it is defined, otherwise returns an empty symbol and `false`.
func (s *SymbolTable) ResolveCurrentScope(name string) (sym Symbol, exists bool) {
	sym, exists = s.store[name]
	return sym, exists
}

// hasOuter returns true if `s` has an outer symbol table, otherwise false.
func (s *SymbolTable) hasOuter() bool {
	return s.outer != nil
}
