package language

import (
	"sync"

	"github.com/pkg/errors"
)

// Registry manages a collection of compilers for different languages.
type Registry struct {
	compilers map[string]Compiler
	language  string
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
		return errors.Wrapf(ErrConflict, "failed to register language '%s'", language)
	}
	r.compilers[language] = compiler
	return nil
}

// Unregister removes a language and its compiler from the registry.
func (r *Registry) Unregister(language string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.compilers[language]; !ok {
		return errors.Wrapf(ErrNotFound, "failed to unregister language '%s'", language)
	}
	delete(r.compilers, language)
	return nil
}

// Lookup retrieves the compiler for a given language.
func (r *Registry) Lookup(language string) (Compiler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.compilers[language]
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "compiler not found for language '%s'", language)
	}
	return c, nil
}

// SetDefault adds a new language and its compiler to the registry.
func (r *Registry) SetDefault(language string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.language = language
}

// Default retrieves the default compiler.
func (r *Registry) Default() (Compiler, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.compilers[r.language]
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "default compiler not found for language '%s'", r.language)
	}
	return c, nil
}

// Close shuts down all registered drivers and clears the registry.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.compilers = make(map[string]Compiler)
	r.language = ""
	return nil
}
