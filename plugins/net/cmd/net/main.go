package main

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"

	node2 "github.com/siyul-park/uniflow/plugins/net/pkg/node"
)

// Plugin implements the plugin that registers network-related node2s.
type Plugin struct {
	hookBuilder   *hook.Builder
	schemeBuilder *scheme.Builder
	mu            sync.Mutex
}

var (
	name    string
	version string
)

var (
	_ plugin.Plugin   = (*Plugin)(nil)
	_ hook.Register   = (*Plugin)(nil)
	_ scheme.Register = (*Plugin)(nil)
)

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

// SetSchemeBuilder sets the scheme builder for the plugin.
func (p *Plugin) SetSchemeBuilder(builder *scheme.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.schemeBuilder = builder
}

// Name returns the plugin's package path as its name.
func (p *Plugin) Name() string {
	return name
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// Load registers network nodes to the scheme.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil || p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}
	p.hookBuilder.Register(p)
	p.schemeBuilder.Register(p)
	return nil
}

// Unload releases plugin resources (no-op).
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil || p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}
	p.hookBuilder.Unregister(p)
	p.schemeBuilder.Unregister(p)
	return nil
}

// AddToHook registers lifecycle hooks for HTTPListenNode.
func (p *Plugin) AddToHook(h *hook.Hook) error {
	h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
		var n *node2.HTTPListenNode
		if node.As(sb, &n) {
			return n.Listen()
		}
		return nil
	}))
	h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
		var n *node2.HTTPListenNode
		if node.As(sb, &n) {
			return n.Shutdown()
		}
		return nil
	}))
	return nil
}

// AddToScheme registers node types and codecs to the scheme.
func (p *Plugin) AddToScheme(s *scheme.Scheme) error {
	definitions := []struct {
		kind  string
		codec scheme.Codec
		spec  spec.Spec
	}{
		{node2.KindHTTP, node2.NewHTTPNodeCodec(), &node2.HTTPNodeSpec{}},
		{node2.KindListener, node2.NewListenNodeCodec(), &node2.ListenNodeSpec{}},
		{node2.KindRouter, node2.NewRouteNodeCodec(), &node2.RouteNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}
	return nil
}
