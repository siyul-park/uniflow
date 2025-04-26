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
	"github.com/siyul-park/uniflow/pkg/testing"

	node2 "github.com/siyul-park/uniflow/plugin/testing/pkg/node"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	hookBuilder   *hook.Builder
	schemeBuilder *scheme.Builder
	testingRunner *testing.Runner
	mu            sync.Mutex
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

// SetTestingRunner sets the testing runner for the plugin.
func (p *Plugin) SetTestingRunner(runner *testing.Runner) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.testingRunner = runner
}

// Load registers testing nodes and hooks to the scheme and hook builder.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hookBuilder == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "missing hook builder")
	}
	if p.schemeBuilder == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "missing scheme builder")
	}
	if p.testingRunner == nil {
		return errors.Wrap(plugin.ErrMissingDependency, "missing testing runner")
	}

	p.hookBuilder.Register(hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			var n *node2.TestNode
			if node.As(sb, &n) {
				p.testingRunner.Register(sb.NamespacedName(), n)
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			var n *node2.TestNode
			if node.As(sb, &n) {
				p.testingRunner.Unregister(sb.NamespacedName())
			}
			return nil
		}))
		return nil
	}))

	p.schemeBuilder.Register(scheme.RegisterFunc(func(s *scheme.Scheme) error {
		definitions := []struct {
			kind  string
			codec scheme.Codec
			spec  spec.Spec
		}{
			{node2.KindTest, node2.NewTestNodeCodec(), &node2.TestNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	}))

	return nil
}

// Unload releases plugin resources (no-op).
func (p *Plugin) Unload(_ context.Context) error {
	return nil
}
