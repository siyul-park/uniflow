package main

import (
	"context"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/plugin"

	cel2 "github.com/siyul-park/uniflow/plugins/cel/pkg/cel"
)

// Config defines the plugin configuration, specifying CEL extensions to load.
type Config struct {
	Extensions []string `json:"extensions"`
}

// Plugin provides a CEL plugin that registers a CEL compiler with optional extensions.
type Plugin struct {
	languageRegistry *language.Registry
	extensions       []string
	mu               sync.Mutex
}

var ErrUnsupportedExtension = errors.New("unsupported extension requested")

var options = map[string]cel.EnvOption{
	"encoders":               ext.Encoders(),
	"math":                   ext.Math(),
	"lists":                  ext.Lists(),
	"sets":                   ext.Sets(),
	"strings":                ext.Strings(),
	"protos":                 ext.Protos(),
	"two_var_comprehensions": ext.TwoVarComprehensions(),
}

var _ plugin.Plugin = (*Plugin)(nil)

// New creates and returns a new CEL plugin with the given configuration.
func New(config Config) *Plugin {
	return &Plugin{extensions: config.Extensions}
}

// SetLanguageRegistry sets the language registry that will be used by the plugin.
func (p *Plugin) SetLanguageRegistry(registry *language.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.languageRegistry = registry
}

// Load registers the CEL compiler with the configured extensions in the language registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.languageRegistry == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	var opts []cel.EnvOption
	for _, e := range p.extensions {
		opt, ok := options[e]
		if !ok {
			return errors.Wrapf(ErrUnsupportedExtension, "unsupported extension '%s'", e)
		}
		opts = append(opts, opt)
	}
	return p.languageRegistry.Register(cel2.Language, cel2.NewCompiler(opts...))
}

// Unload cleans up resources when the plugin is unloaded.
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.languageRegistry == nil {
		return nil
	}
	return p.languageRegistry.Unregister(cel2.Language)
}
