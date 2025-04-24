package symbol

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/spec"
)

// Symbol represents a Node that is identifiable within a Spec.
type Symbol struct {
	Spec spec.Spec                // Spec holds the specification of the Symbol.
	Node node.Node                // Node is the underlying node of the Symbol.
	ins  map[string]*port.InPort  // ins is a map of input ports.
	outs map[string]*port.OutPort // outs is a map of output ports.
	mu   sync.RWMutex
}

var _ node.Node = (*Symbol)(nil)
var _ node.Proxy = (*Symbol)(nil)

// ID returns the unique identifier of the Symbol.
func (s *Symbol) ID() uuid.UUID {
	return s.Spec.GetID()
}

// SetID sets the unique identifier of the Symbol.
func (s *Symbol) SetID(id uuid.UUID) {
	s.Spec.SetID(id)
}

// Kind returns the type of the Symbol.
func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

// SetKind sets the type of the Symbol.
func (s *Symbol) SetKind(kind string) {
	s.Spec.SetKind(kind)
}

// Namespace returns the namespace to which the Symbol belongs.
func (s *Symbol) Namespace() string {
	return s.Spec.GetNamespace()
}

// SetNamespace sets the namespace of the Symbol.
func (s *Symbol) SetNamespace(namespace string) {
	s.Spec.SetNamespace(namespace)
}

// Name returns the human-readable name of the Symbol.
func (s *Symbol) Name() string {
	return s.Spec.GetName()
}

// SetName sets the human-readable name of the Symbol.
func (s *Symbol) SetName(name string) {
	s.Spec.SetName(name)
}

// NamespacedName returns the namespaced identifier.
func (s *Symbol) NamespacedName() string {
	return s.Spec.GetNamespacedName()
}

// Annotations returns the annotations associated with the Symbol.
func (s *Symbol) Annotations() map[string]string {
	return s.Spec.GetAnnotations()
}

// SetAnnotations sets the annotations of the Symbol.
func (s *Symbol) SetAnnotations(annotations map[string]string) {
	s.Spec.SetAnnotations(annotations)
}

// Env returns the environment variables associated with the Symbol.
func (s *Symbol) Env() map[string]spec.Value {
	return s.Spec.GetEnv()
}

// SetEnv sets the environment variables of the Symbol.
func (s *Symbol) SetEnv(env map[string]spec.Value) {
	s.Spec.SetEnv(env)
}

// Ports returns the ports associated with the Symbol.
func (s *Symbol) Ports() map[string][]spec.Port {
	return s.Spec.GetPorts()
}

// SetPorts sets the ports of the Symbol.
func (s *Symbol) SetPorts(ports map[string][]spec.Port) {
	s.Spec.SetPorts(ports)
}

// Ins returns the input ports associated with the Symbol.
func (s *Symbol) Ins() map[string]*port.InPort {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ins := make(map[string]*port.InPort, len(s.ins))
	for name, in := range s.ins {
		ins[name] = in
	}
	return ins
}

// In returns the input port with the specified name, caching the result.
func (s *Symbol) In(name string) *port.InPort {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ins == nil {
		s.ins = make(map[string]*port.InPort)
	}

	p, ok := s.ins[name]
	if !ok && s.Node != nil {
		if p = s.Node.In(name); p != nil {
			s.ins[name] = p
		}
	}
	return p
}

// Outs returns the output ports associated with the Symbol.
func (s *Symbol) Outs() map[string]*port.OutPort {
	s.mu.RLock()
	defer s.mu.RUnlock()

	outs := make(map[string]*port.OutPort, len(s.outs))
	for name, out := range s.outs {
		outs[name] = out
	}
	return outs
}

// Out returns the output port with the specified name, caching the result.
func (s *Symbol) Out(name string) *port.OutPort {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.outs == nil {
		s.outs = make(map[string]*port.OutPort)
	}

	p, ok := s.outs[name]
	if !ok && s.Node != nil {
		if p = s.Node.Out(name); p != nil {
			s.outs[name] = p
		}
	}
	return p
}

// Unwrap returns the underlying Node from the Symbol.
func (s *Symbol) Unwrap() node.Node {
	return s.Node
}

// Close frees all resources held by the Symbol.
func (s *Symbol) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ins = nil
	s.outs = nil

	if s.Node == nil {
		return nil
	}
	return s.Node.Close()
}
