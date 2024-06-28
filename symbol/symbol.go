package symbol

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/spec"
)

// Symbol represents an object that binds a Node and a Spec.
type Symbol struct {
	spec    spec.Spec
	node    node.Node
	links   map[string][]spec.PortLocation
	unlinks map[string][]spec.PortLocation
	linked  map[string][]spec.PortLocation
}

var _ node.Node = (*Symbol)(nil)

// New returns a new Symbol.
func New(s spec.Spec, n node.Node) *Symbol {
	return &Symbol{
		spec:    s,
		node:    n,
		links:   s.GetLinks(),
		unlinks: make(map[string][]spec.PortLocation),
		linked:  make(map[string][]spec.PortLocation),
	}
}

// ID returns the unique identifier.
func (s *Symbol) ID() uuid.UUID {
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

// Annotations returns the annotations.
func (s *Symbol) Annotations() map[string]string {
	return s.spec.GetAnnotations()
}

// Spec returns the spec.
func (s *Symbol) Spec() spec.Spec {
	return s.spec
}

// Unwrap returns the node wrapped.
func (s *Symbol) Unwrap() node.Node {
	return s.node
}

// In returns the specified InPort.
func (s *Symbol) In(name string) *port.InPort {
	return s.node.In(name)
}

// Out returns the specified OutPort.
func (s *Symbol) Out(name string) *port.OutPort {
	return s.node.Out(name)
}

// Close closes the Symbol, invoking the Close method of its Node.
func (s *Symbol) Close() error {
	return s.node.Close()
}
