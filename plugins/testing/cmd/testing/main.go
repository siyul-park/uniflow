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

	node2 "github.com/siyul-park/uniflow/plugins/testing/pkg/node"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	runner        *testing.Runner
	agent            *runtime.Agent
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

// SetRunner sets the testing runner for the plugin.
func (p *Plugin) SetRunner(runner *testing.Runner) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.runner = runner
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

// Name returns the plugin's package path as its name.
func (p *Plugin) Name() string {
	return name
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// Load registers testing nodes and hooks to the scheme and hook builder.
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

// Unload releases plugin resources.
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

// AddToHook adds lifecycle hooks for test nodes.
func (p *Plugin) AddToHook(h *hook.Hook) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	testingRunner := p.testingRunner
	if testingRunner == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
		var n *node2.TestNode
		if node.As(sb, &n) {
			testingRunner.Register(sb.NamespacedName(), n)
		}
		return nil
	}))
	h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
		var n *node2.TestNode
		if node.As(sb, &n) {
			testingRunner.Unregister(sb.NamespacedName())
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
		{node2.KindTest, node2.NewTestNodeCodec(), &node2.TestNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}
	return nil
}
