package symbol

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Symbol represents a Node that is identifiable within a Spec.
type Symbol struct {
	spec    spec.Spec
	node    node.Node
	links   map[string][]spec.PortLocation
	unlinks map[string][]spec.PortLocation
	linked  map[string][]spec.PortLocation
}

var _ node.Node = (*Symbol)(nil)

// New creates and returns a new Symbol instance.
func New(s spec.Spec, n node.Node) *Symbol {
	return &Symbol{
		spec:    s,
		node:    n,
		links:   s.GetLinks(),
		unlinks: make(map[string][]spec.PortLocation),
		linked:  make(map[string][]spec.PortLocation),
	}
}

// ID returns the unique identifier of the Symbol.
func (s *Symbol) ID() uuid.UUID {
	return s.spec.GetID()
}

// Kind returns the kind or type of the Symbol.
func (s *Symbol) Kind() string {
	return s.spec.GetKind()
}

// Namespace returns the namespace of the Symbol.
func (s *Symbol) Namespace() string {
	return s.spec.GetNamespace()
}

// Name returns the human-readable name of the Symbol.
func (s *Symbol) Name() string {
	return s.spec.GetName()
}

// Annotations returns the annotations associated with the Symbol.
func (s *Symbol) Annotations() map[string]string {
	return s.spec.GetAnnotations()
}

// Spec returns the Spec associated with the Symbol.
func (s *Symbol) Spec() spec.Spec {
	return s.spec
}

// Unwrap returns the underlying Node wrapped by the Symbol.
func (s *Symbol) Unwrap() node.Node {
	return s.node
}

// In returns the input port with the specified name.
func (s *Symbol) In(name string) *port.InPort {
	return s.node.In(name)
}

// Out returns the output port with the specified name.
func (s *Symbol) Out(name string) *port.OutPort {
	return s.node.Out(name)
}

// Close frees all resources held by the Symbol.
func (s *Symbol) Close() error {
	return s.node.Close()
}
