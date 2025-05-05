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
	hookBuilder *hook.Builder
	mu          sync.Mutex
}

var _ plugin.Plugin = (*Plugin)(nil)

// New returns a new Plugin instance.
func New() *Plugin {
	return &Plugin{}
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

	if p.hookBuilder == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "missing hook builder")
	}

	p.hookBuilder.Register(hook.RegisterFunc(func(h *hook.Hook) error {
		agent := runtime.NewAgent()

		drv := driver.New(driver.WithRegistry(schema.NewInMemoryRegistry(map[string]schema.Catalog{
			"system": schema.NewInMemoryCatalog(map[string]schema.Table{
				"frames":    table.NewFrameTable(agent),
				"processes": table.NewProcessTable(agent),
				"symbols":   table.NewSymbolTable(agent),
			}),
		})))

		h.AddLoadHook(agent)
		h.AddUnloadHook(agent)

		sql.Register("runtime", drv)

		return nil
	}))
	return nil
}

// Unload releases plugin resources (no-op).
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
