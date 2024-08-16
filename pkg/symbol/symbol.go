package symbol

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Symbol represents a Node that is identifiable within a Spec.
type Symbol struct {
	Spec     spec.Spec
	Node     node.Node
	inbounds map[string][]spec.Port
	ins      map[string]*port.InPort
	outs     map[string]*port.OutPort
}

var _ node.Node = (*Symbol)(nil)

// ID returns the unique identifier of the Symbol.
func (s *Symbol) ID() uuid.UUID {
	return s.Spec.GetID()
}

// Kind returns the type of the Symbol.
func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

// Namespace returns the namespace to which the Symbol belongs.
func (s *Symbol) Namespace() string {
	return s.Spec.GetNamespace()
}

// Name returns the human-readable name of the Symbol.
func (s *Symbol) Name() string {
	return s.Spec.GetName()
}

// Annotations returns the annotations associated with the Symbol.
func (s *Symbol) Annotations() map[string]string {
	return s.Spec.GetAnnotations()
}

// Ports returns the ports associated with the Symbol.
func (s *Symbol) Ports() map[string][]spec.Port {
	return s.Spec.GetPorts()
}

// Links returns the references associated with the Symbol.
func (s *Symbol) Links() map[string][]spec.Port {
	links := make(map[string][]spec.Port)
	for name, ports := range s.Spec.GetPorts() {
		links[name] = append(links[name], ports...)
	}
	for name, ports := range s.inbounds {
		links[name] = append(links[name], ports...)
	}
	return links
}

// Env returns the environment variables associated with the Symbol.
func (s *Symbol) Env() map[string][]spec.Secret {
	return s.Spec.GetEnv()
}

// Ins returns the input ports associated with the Symbol.
func (s *Symbol) Ins() []string {
	ins := make([]string, 0, len(s.ins))
	for name := range s.ins {
		ins = append(ins, name)
	}
	return ins
}

// In returns the input port with the specified name, caching the result.
func (s *Symbol) In(name string) *port.InPort {
	if s.ins == nil {
		s.ins = make(map[string]*port.InPort)
	}
	p := s.Node.In(name)
	if p != nil {
		s.ins[name] = p
	}
	return p
}

// Outs returns the output ports associated with the Symbol.
func (s *Symbol) Outs() []string {
	outs := make([]string, 0, len(s.outs))
	for name := range s.outs {
		outs = append(outs, name)
	}
	return outs
}

// Out returns the output port with the specified name, caching the result.
func (s *Symbol) Out(name string) *port.OutPort {
	if s.outs == nil {
		s.outs = make(map[string]*port.OutPort)
	}
	p := s.Node.Out(name)
	if p != nil {
		s.outs[name] = p
	}
	return p
}

// Close frees all resources held by the Symbol.
func (s *Symbol) Close() error {
	return s.Node.Close()
}
