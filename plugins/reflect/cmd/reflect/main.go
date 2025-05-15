package main

import (
	"context"
	"database/sql"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/sqlbridge/driver"
	"github.com/siyul-park/sqlbridge/schema"
	driver2 "github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/runtime"

	driver3 "github.com/siyul-park/uniflow/plugins/reflect/pkg/driver"
	runtime2 "github.com/siyul-park/uniflow/plugins/reflect/pkg/runtime"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	agent       *runtime.Agent
	conn        driver2.Conn
	hookBuilder *hook.Builder
	mu          sync.Mutex
}

var (
	name    string
	version string
)

var drv = driver.New()

var (
	_ plugin.Plugin = (*Plugin)(nil)
	_ hook.Register = (*Plugin)(nil)
)

func init() {
	sql.Register("runtime", drv)
}

// New creates a new Plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// SetAgent sets the agent for the plugin.
func (p *Plugin) SetAgent(agent *runtime.Agent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.agent = agent
}

// SetConn sets the connection for the plugin.
func (p *Plugin) SetConn(conn driver2.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conn = conn
}

// SetHookBuilder sets the hook builder for the plugin.
func (p *Plugin) SetHookBuilder(builder *hook.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hookBuilder = builder
}

// Name returns the plugin's package path as its name.
func (p *Plugin) Name() string {
	return name
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// Load registers testing nodes and hooks to the schema and hook builder.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.agent == nil || p.conn == nil || p.hookBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	reg := schema.NewInMemoryRegistry(map[string]schema.Catalog{
		"": schema.NewCompositeCatalog(
			runtime2.NewCatalog(p.agent),
			driver3.NewCatalog(p.conn),
		),
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

// AddToHook adds plugin-related hooks to the provided hook.
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
