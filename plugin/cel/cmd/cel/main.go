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

// Config defines the plugin configuration with a list of modules to load.
type Config struct {
	Modules []string `json:"modules"`
}

// Plugin implements the CEL plugin interface, managing modules and the language registry.
type Plugin struct {
	registry *language.Registry
	modules  []string
	mu       sync.Mutex
}

var ErrUnsupportedModule = errors.New("unsupported module requested")

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
	return &Plugin{modules: config.Modules}
}

// SetRegistry assigns a language registry to the plugin.
func (p *Plugin) SetRegistry(registry *language.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.registry = registry
}

// Load registers the specified modules with the registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.registry == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "failed to load plugin: missing language registry")
	}

	var opts []cel.EnvOption
	for _, module := range p.modules {
		opt, ok := options[module]
		if !ok {
			return errors.Wrapf(ErrUnsupportedModule, "failed to load plugin: unsupported module '%s'", module)
		}
		opts = append(opts, opt)
	}
	return p.registry.Register(cel2.Language, cel2.NewCompiler())
}

// Unload is a placeholder for cleanup when unloading the plugin.
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
