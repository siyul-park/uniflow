package main

import (
	"context"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/plugin"

	cel2 "github.com/siyul-park/uniflow/plugin/cel/pkg/cel"
)

// Config defines the plugin configuration with a list of extensions to load.
type Config struct {
	Extensions []string `json:"extensions"`
}

// Plugin implements the CEL plugin interface, managing extensions and the language registry.
type Plugin struct {
	registry   *language.Registry
	extensions []string
	mu         sync.Mutex
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

// New creates a new Plugin with the specified configuration.
func New(config Config) plugin.Plugin {
	return &Plugin{extensions: config.Extensions}
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

	if p.registry == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "failed to load plugin: missing language registry")
	}

	var opts []cel.EnvOption
	for _, e := range p.extensions {
		opt, ok := options[e]
		if !ok {
			return errors.Wrapf(ErrUnsupportedExtension, "failed to load plugin: unsupported extension '%s'", e)
		}
		opts = append(opts, opt)
	}
	return p.registry.Register(cel2.Language, cel2.NewCompiler(opts...))
}

// Unload is a placeholder for cleanup when unloading the plugin.
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
