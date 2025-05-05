package main

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"

	node2 "github.com/siyul-park/uniflow/plugins/sql/pkg/node"
)

// Plugin implements the plugin that registers testing-related nodes.
type Plugin struct {
	schemeBuilder  *scheme.Builder
	schemeRegister scheme.Register
	mu             sync.Mutex
}

var _ plugin.Plugin = (*Plugin)(nil)

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

	if p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	if p.schemeRegister == nil {
		p.schemeRegister = scheme.RegisterFunc(func(s *scheme.Scheme) error {
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
		})
	}

	p.schemeBuilder.Register(p.schemeRegister)

	return nil
}

// Unload releases plugin resources.
func (p *Plugin) Unload(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schemeBuilder == nil {
		return errors.WithStack(plugin.ErrMissingDependency)
	}

	if p.schemeRegister != nil {
		p.schemeBuilder.Unregister(p.schemeRegister)
	}
	return nil
}
