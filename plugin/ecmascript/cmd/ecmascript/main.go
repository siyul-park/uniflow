package main

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/plugin"

	"github.com/siyul-park/uniflow/plugin/ecmascript/pkg/javascript"
	"github.com/siyul-park/uniflow/plugin/ecmascript/pkg/typescript"
)

// Plugin implements the CEL plugin interface, managing extensions and the language registry.
type Plugin struct {
	registry *language.Registry
	mu       sync.Mutex
}

var _ plugin.Plugin = (*Plugin)(nil)

// New creates a new Plugin with the specified configuration.
func New() *Plugin {
	return &Plugin{}
}

// SetRegistry assigns a language registry to the plugin.
func (p *Plugin) SetRegistry(registry *language.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.registry = registry
}

// Load registers the specified extensions with the registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.registry.Register(javascript.Language, javascript.NewCompiler()); err != nil {
		return err
	}
	return p.registry.Register(typescript.Language, typescript.NewCompiler())
}

// Unload is a placeholder for cleanup when unloading the plugin.
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
