package language

import (
	"sync"

	"github.com/pkg/errors"
)

// Module represents a collection of compilers identified by language.
type Module struct {
	compilers map[string]Compiler
	mu        sync.RWMutex
}

var ErrInvalidLanguage = errors.New("language is invalid")

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
func (m *Module) Load(lang string) (Compiler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if compiler, ok := m.compilers[lang]; !ok {
		return nil, errors.WithStack(ErrInvalidLanguage)
	} else {
		return compiler, nil
	}
}
