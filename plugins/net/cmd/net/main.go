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
	hookBuilder    *hook.Builder
	schemeBuilder  *scheme.Builder
	schemeRegister scheme.Register
	hookRegister   hook.Register
	mu             sync.Mutex
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

// SetSchemeBuilder sets the scheme builder for the plugin.
func (p *Plugin) SetSchemeBuilder(builder *scheme.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.schemeBuilder = builder
}

// Load registers network nodes to the scheme.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil || p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	if p.hookRegister == nil {
		p.hookRegister = hook.RegisterFunc(func(h *hook.Hook) error {
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
		})
	}
	if p.schemeRegister == nil {
		p.schemeRegister = scheme.RegisterFunc(func(s *scheme.Scheme) error {
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
		})
	}

	p.hookBuilder.Register(p.hookRegister)
	p.schemeBuilder.Register(p.schemeRegister)

	return nil
}

// Unload releases plugin resources (no-op).
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil || p.schemeBuilder == nil {
		return nil
	}

	if p.hookRegister != nil {
		p.hookBuilder.Unregister(p.hookRegister)
	}
	if p.schemeRegister != nil {
		p.schemeBuilder.Unregister(p.schemeRegister)
	}
	return nil
}
