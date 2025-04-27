package main

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/plugin"

	"github.com/siyul-park/uniflow/plugins/ecmascript/pkg/javascript"
	"github.com/siyul-park/uniflow/plugins/ecmascript/pkg/typescript"
)

// Plugin implements a plugin that registers ECMAScript language compilers (JavaScript and TypeScript).
type Plugin struct {
	languageRegistry *language.Registry
	mu               sync.Mutex
}

var _ plugin.Plugin = (*Plugin)(nil)

// New returns a new instance of the ECMAScript plugin.
func New() *Plugin {
	return &Plugin{}
}

// SetLanguageRegistry sets the language registry to be used by the plugin.
func (p *Plugin) SetLanguageRegistry(registry *language.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.languageRegistry = registry
}

// Load registers the JavaScript and TypeScript compilers with the language registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.languageRegistry == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "missing language registry")
	}

	if err := p.languageRegistry.Register(javascript.Language, javascript.NewCompiler()); err != nil {
		return err
	}
	return p.languageRegistry.Register(typescript.Language, typescript.NewCompiler())
}

// Unload performs cleanup when the plugin is unloaded (currently no-op).
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
