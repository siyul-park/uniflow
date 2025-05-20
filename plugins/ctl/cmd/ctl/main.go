package main

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"

	"github.com/siyul-park/uniflow/plugins/ctl/pkg/node"
)

// Plugin registers control nodes to the scheme and language registry.
type Plugin struct {
	schemeBuilder    *scheme.Builder
	languageRegistry *language.Registry
	mu               sync.Mutex
}

var (
	name    string
	version string
)

var (
	_ plugin.Plugin   = (*Plugin)(nil)
	_ scheme.Register = (*Plugin)(nil)
)

// New creates a new Plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// SetSchemeBuilder sets the scheme builder.
func (p *Plugin) SetSchemeBuilder(builder *scheme.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.schemeBuilder = builder
}

// SetLanguageRegistry sets the language registry.
func (p *Plugin) SetLanguageRegistry(registry *language.Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.languageRegistry = registry
}

// Name returns the plugin's package path as its name.
func (p *Plugin) Name() string {
	return name
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// Load registers control nodes to the scheme using the provided builder and registry.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}
	p.schemeBuilder.Register(p)
	return nil
}

// Unload performs cleanup when the plugin is unloaded.
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}
	p.schemeBuilder.Unregister(p)
	return nil
}

// AddToScheme registers node types and codecs to the scheme.
func (p *Plugin) AddToScheme(s *scheme.Scheme) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.languageRegistry == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	compiler, err := p.languageRegistry.Default()
	if err != nil {
		return err
	}

	definitions := []struct {
		kind  string
		codec scheme.Codec
		spec  spec.Spec
	}{
		{node.KindBlock, node.NewBlockNodeCodec(s), &node.BlockNodeSpec{}},
		{node.KindFor, node.NewForNodeCodec(), &node.ForNodeSpec{}},
		{node.KindFork, node.NewForkNodeCodec(), &node.ForkNodeSpec{}},
		{node.KindIf, node.NewIfNodeCodec(compiler), &node.IfNodeSpec{}},
		{node.KindMerge, node.NewMergeNodeCodec(), &node.MergeNodeSpec{}},
		{node.KindNOP, node.NewNOPNodeCodec(), &node.NOPNodeSpec{}},
		{node.KindPipe, node.NewPipeNodeCodec(), &node.PipeNodeSpec{}},
		{node.KindRetry, node.NewRetryNodeCodec(), &node.RetryNodeSpec{}},
		{node.KindSleep, node.NewSleepNodeCodec(), &node.SleepNodeSpec{}},
		{node.KindSnippet, node.NewSnippetNodeCodec(p.languageRegistry), &node.SnippetNodeSpec{}},
		{node.KindSplit, node.NewSplitNodeCodec(), &node.SplitNodeSpec{}},
		{node.KindStep, node.NewStepNodeCodec(s), &node.StepNodeSpec{}},
		{node.KindSwitch, node.NewSwitchNodeCodec(compiler), &node.SwitchNodeSpec{}},
		{node.KindThrow, node.NewThrowNodeCodec(), &node.ThrowNodeSpec{}},
		{node.KindTry, node.NewTryNodeCodec(), &node.TryNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}
	return nil
}
