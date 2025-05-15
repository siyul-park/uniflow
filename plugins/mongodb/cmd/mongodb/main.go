package main

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/plugin"

	driver2 "github.com/siyul-park/uniflow/plugins/mongodb/pkg/driver"
)

// Plugin provides a CEL plugin that registers a CEL compiler with optional extensions.
type Plugin struct {
	driverRegistry *driver.Registry
	extensions     []string
	mu             sync.Mutex
}

var (
	name    string
	version string
)

var _ plugin.Plugin = (*Plugin)(nil)

// New creates and returns a new CEL plugin with the given configuration.
func New() *Plugin {
	return &Plugin{}
}

// SetLanguageRegistry sets the language registry that will be used by the plugin.
func (p *Plugin) SetLanguageRegistry(registry *driver.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.driverRegistry = registry
}

// Name returns the plugin's package path as its name.
func (p *Plugin) Name() string {
	return name
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// Load registers the CEL compiler with the configured extensions in the language registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.driverRegistry == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}
	return p.driverRegistry.Register("mongodb", driver2.New())
}

// Unload cleans up resources when the plugin is unloaded (currently a no-op).
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.driverRegistry == nil {
		return nil
	}
	return p.driverRegistry.Unregister("mongodb")
}
