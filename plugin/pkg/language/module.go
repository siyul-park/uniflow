package language

import (
	"sync"
)

// Module represents a collection of compilers identified by language.
type Module struct {
	compilers map[string]Compiler
	mu        sync.RWMutex
}

// NewModule creates and returns a new Module instance.
func NewModule() *Module {
	return &Module{
		compilers: make(map[string]Compiler),
	}
}

// Store adds a compiler to the module for the given language.
func (m *Module) Store(lang string, compiler Compiler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.compilers[lang] = compiler
}

// Load retrieves the compiler for the given language from the module.
// Returns the compiler and a boolean indicating whether the compiler was found.
func (m *Module) Load(lang string) (Compiler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	compiler, ok := m.compilers[lang]
	return compiler, ok
}
