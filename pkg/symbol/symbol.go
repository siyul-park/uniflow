package symbol

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	// Symbol represents an object that binds a Node and a Spec.
	Symbol struct {
		Node node.Node
		Spec scheme.Spec
	}
)

// ID returns the unique identifier of the Symbol, based on its Node.
func (s *Symbol) ID() ulid.ULID {
	return s.Node.ID()
}

// Kind returns the kind of the Symbol, based on its Spec.
func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

// Namespace returns the namespace of the Symbol, based on its Spec.
func (s *Symbol) Namespace() string {
	return s.Spec.GetNamespace()
}

// Name returns the name of the Symbol, based on its Spec.
func (s *Symbol) Name() string {
	return s.Spec.GetName()
}

// Links returns the links of the Symbol, based on its Spec.
func (s *Symbol) Links() map[string][]scheme.PortLocation {
	return s.Spec.GetLinks()
}

// Port returns the specified port of the Symbol, based on its Node.
func (s *Symbol) Port(name string) (*port.Port, bool) {
	return s.Node.Port(name)
}

// Close closes the Symbol, invoking the Close method of its Node.
func (s *Symbol) Close() error {
	return s.Node.Close()
}
