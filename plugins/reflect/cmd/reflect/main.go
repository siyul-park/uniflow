package main

import (
	"context"
	"database/sql"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/sqlbridge/driver"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/runtime"

	"github.com/siyul-park/uniflow/plugins/reflect/pkg/table"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	agent       *runtime.Agent
	hookBuilder *hook.Builder
	mu          sync.Mutex
}

var drv = driver.New()

func init() {
	sql.Register("runtime", drv)
}

var (
	_ plugin.Plugin = (*Plugin)(nil)
	_ hook.Register = (*Plugin)(nil)
)

// New returns a new Plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// SetAgent sets the agent for the plugin.
func (p *Plugin) SetAgent(agent *runtime.Agent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.agent = agent
}

// SetHookBuilder sets the hook builder for the plugin.
func (p *Plugin) SetHookBuilder(builder *hook.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hookBuilder = builder
}

// Load registers testing nodes and hooks to the scheme and hook builder.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.agent == nil || p.hookBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	reg := schema.NewInMemoryRegistry(map[string]schema.Catalog{
		"": schema.NewInMemoryCatalog(map[string]schema.Table{
			"frames":    table.NewFrameTable(p.agent),
			"processes": table.NewProcessTable(p.agent),
			"symbols":   table.NewSymbolTable(p.agent),
		}),
	})
	driver.WithRegistry(reg)(drv)

	p.hookBuilder.Register(p)
	return nil
}

// Unload releases plugin resources.
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	reg := schema.NewInMemoryRegistry(nil)
	driver.WithRegistry(reg)(drv)

	p.hookBuilder.Unregister(p)
	return nil
}

func (p *Plugin) AddToHook(h *hook.Hook) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.agent == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	h.AddLoadHook(p.agent)
	h.AddUnloadHook(p.agent)
	return nil
}
