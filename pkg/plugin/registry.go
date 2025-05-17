package plugin

import (
	"context"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Registry manages a list of plugins and controls their lifecycle.
type Registry struct {
	plugins []Plugin
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

	for _, p := range r.plugins {
		if p == plugin {
			return errors.WithStack(ErrConflict)
		}
	}
	r.plugins = append(r.plugins, plugin)
	return nil
}

// Unregister removes a plugin from the registry.
func (r *Registry) Unregister(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.plugins {
		if p == plugin {
			r.plugins = append(r.plugins[:i], r.plugins[i+1:]...)
			return nil
		}
	}
	return errors.WithStack(ErrNotFound)
}

// Plugins returns a slice of all registered plugins.
func (r *Registry) Plugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]Plugin(nil), r.plugins...)
}

// Inject attempts to inject the given dependency into all registered plugins.
func (r *Registry) Inject(dependency any) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, p := range r.plugins {
		pv := reflect.ValueOf(p)
		pt := pv.Type()

		dv := reflect.ValueOf(dependency)
		dt := dv.Type()

		for i := 0; i < pt.NumMethod(); i++ {
			m := pt.Method(i)
			if !strings.HasPrefix(m.Name, "Set") {
				continue
			}

			mv := pv.Method(i)
			mt := mv.Type()

			if mt.NumIn() == 1 && dt.AssignableTo(mt.In(0)) {
				ret := mv.Call([]reflect.Value{dv})
				if len(ret) > 0 {
					if err, ok := ret[0].Interface().(error); ok && err != nil {
						return 0, err
					}
				}
				count++
			}
		}
	}
	return count, nil
}

// Load calls Load on all registered plugins.
func (r *Registry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < len(r.plugins); i++ {
		if err := r.plugins[i].Load(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Unload calls Unload on all registered plugins in reverse order.
func (r *Registry) Unload(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.plugins) - 1; i >= 0; i-- {
		if err := r.plugins[i].Unload(ctx); err != nil {
			return err
		}
	}
	return nil
}
