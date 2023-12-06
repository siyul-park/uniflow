package symbol

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// Symbol represents an object that binds a Node and a Spec.
type Symbol struct {
	Node node.Node
	Spec scheme.Spec
}

var _ node.Node = (*Symbol)(nil)

// ID returns the unique identifier.
func (s *Symbol) ID() ulid.ULID {
	return s.Spec.GetID()
}

// Kind returns the kind.
func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

// Namespace returns the namespace.
func (s *Symbol) Namespace() string {
	return s.Spec.GetNamespace()
}

// Name returns the name.
func (s *Symbol) Name() string {
	return s.Spec.GetName()
}

// Links returns the links.
func (s *Symbol) Links() map[string][]scheme.PortLocation {
	return s.Spec.GetLinks()
}

// Port returns the specified port.
func (s *Symbol) Port(name string) (*port.Port, bool) {
	return s.Node.Port(name)
}

// Close closes the Symbol, invoking the Close method of its Node.
func (s *Symbol) Close() error {
	return s.Node.Close()
}
