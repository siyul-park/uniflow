package main

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"

	node2 "github.com/siyul-park/uniflow/plugins/sql/pkg/node"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	schemeBuilder *scheme.Builder
	mu            sync.Mutex
}

var (
	_ plugin.Plugin   = (*Plugin)(nil)
	_ scheme.Register = (*Plugin)(nil)
)

// New returns a new Plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// SetSchemeBuilder sets the scheme builder for the plugin.
func (p *Plugin) SetSchemeBuilder(builder *scheme.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.schemeBuilder = builder
}

// Load registers testing nodes and hooks to the scheme and hook builder.
func (p *Plugin) Load(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schemeBuilder != nil {
		p.schemeBuilder.Register(p)
	}
	return nil
}

// Unload releases plugin resources.
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schemeBuilder != nil {
		p.schemeBuilder.Unregister(p)
	}
	return nil
}

// AddToScheme registers node types and codecs to the scheme.
func (p *Plugin) AddToScheme(s *scheme.Scheme) error {
	definitions := []struct {
		kind  string
		codec scheme.Codec
		spec  spec.Spec
	}{
		{node2.KindSQL, node2.NewSQLNodeCodec(), &node2.SQLNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}
	return nil
}
