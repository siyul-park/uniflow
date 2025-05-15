package plugin

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// Registry manages a list of plugins and controls their lifecycle.
type Registry struct {
	proxies []*Proxy
	mu      sync.RWMutex
}

var (
	ErrConflict = errors.New("plugin conflict occurred")
	ErrNotFound = errors.New("plugin not found")
)

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.proxies {
		if p.Unwrap() == plugin {
			return errors.WithStack(ErrConflict)
		}
	}
	r.proxies = append(r.proxies, NewProxy(plugin))
	return nil
}

// Unregister removes a plugin from the registry.
func (r *Registry) Unregister(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.proxies {
		if p.Unwrap() == plugin {
			r.proxies = append(r.proxies[:i], r.proxies[i+1:]...)
			return nil
		}
	}
	return errors.WithStack(ErrNotFound)
}

// Plugins returns a slice of all registered plugins.
func (r *Registry) Plugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, len(r.proxies))
	for i, p := range r.proxies {
		plugins[i] = p.Unwrap()
	}
	return plugins
}

// Inject attempts to inject the given dependency into all registered plugins.
func (r *Registry) Inject(dependency any) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, p := range r.proxies {
		if ok, err := p.Inject(dependency); err != nil {
			return 0, err
		} else if ok {
			count++
		}
	}
	return count, nil
}

// Load calls Load on all registered plugins.
func (r *Registry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < len(r.proxies); i++ {
		if err := r.proxies[i].Load(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Unload calls Unload on all registered plugins in reverse order.
func (r *Registry) Unload(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.proxies) - 1; i >= 0; i-- {
		if err := r.proxies[i].Unload(ctx); err != nil {
			return err
		}
	}
	return nil
}
