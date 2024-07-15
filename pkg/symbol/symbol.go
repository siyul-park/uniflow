package symbol

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Symbol represents a Node that is identifiable within a Spec.
type Symbol struct {
	Spec   spec.Spec
	Value  any
	Node   node.Node
	linked map[string][]spec.PortLocation
}

var _ node.Node = (*Symbol)(nil)

// ID returns the unique identifier of the Symbol.
func (s *Symbol) ID() uuid.UUID {
	return s.Spec.GetID()
}

// Kind returns the kind or type of the Symbol.
func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

// Namespace returns the namespace of the Symbol.
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

// Links returns the links associated with the Symbol.
func (s *Symbol) Links() map[string][]spec.PortLocation {
	return s.Spec.GetLinks()
}

// In returns the input port with the specified name.
func (s *Symbol) In(name string) *port.InPort {
	if s.Node == nil {
		return nil
	}
	return s.Node.In(name)
}

// Out returns the output port with the specified name.
func (s *Symbol) Out(name string) *port.OutPort {
	if s.Node == nil {
		return nil
	}
	return s.Node.Out(name)
}

// Close frees all resources held by the Symbol.
func (s *Symbol) Close() error {
	if s.Node == nil {
		return nil
	}
	return s.Node.Close()
}
