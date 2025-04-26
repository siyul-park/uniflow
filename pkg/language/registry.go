package language

import (
	"sync"

	"github.com/pkg/errors"
)

// Registry manages a collection of compilers for different languages.
type Registry struct {
	compilers map[string]Compiler
	mu        sync.RWMutex
}

var (
	ErrConflict = errors.New("language conflict occurred")
	ErrNotFound = errors.New("language not found")
)

// NewRegistry creates and returns a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{compilers: make(map[string]Compiler)}
}

// Register adds a new language and its compiler to the registry.
func (r *Registry) Register(language string, compiler Compiler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.compilers[language]; ok {
		return errors.WithStack(ErrConflict)
	}
	r.compilers[language] = compiler
	return nil
}

// Lookup retrieves the compiler for a given language.
func (r *Registry) Lookup(language string) (Compiler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	compiler, ok := r.compilers[language]
	if !ok {
		return nil, errors.WithStack(ErrNotFound)
	}
	return compiler, nil
}
