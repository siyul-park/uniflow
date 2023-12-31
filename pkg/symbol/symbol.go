package symbol

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// Symbol represents an object that binds a Node and a Spec.
type Symbol struct {
	spec scheme.Spec
	node node.Node

	links   map[string][]scheme.PortLocation
	unlinks map[string][]scheme.PortLocation
	linked  map[string][]scheme.PortLocation
}

var _ node.Node = (*Symbol)(nil)

// New returns a new Symbol.
func New(spec scheme.Spec, node node.Node) *Symbol {
	return &Symbol{
		spec:    spec,
		node:    node,
		links:   spec.GetLinks(),
		unlinks: make(map[string][]scheme.PortLocation),
		linked:  make(map[string][]scheme.PortLocation),
	}
}

// ID returns the unique identifier.
func (s *Symbol) ID() ulid.ULID {
	return s.spec.GetID()
}

// Kind returns the kind.
func (s *Symbol) Kind() string {
	return s.spec.GetKind()
}

// Namespace returns the namespace.
func (s *Symbol) Namespace() string {
	return s.spec.GetNamespace()
}

// Name returns the name.
func (s *Symbol) Name() string {
	return s.spec.GetName()
}

// Port returns the specified port.
func (s *Symbol) Port(name string) (*port.Port, bool) {
	return s.node.Port(name)
}

// Close closes the Symbol, invoking the Close method of its Node.
func (s *Symbol) Close() error {
	return s.node.Close()
}
