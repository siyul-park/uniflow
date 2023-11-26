package symbol

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	Symbol struct {
		Node node.Node
		Spec scheme.Spec
	}
)

var _ node.Node = (*Symbol)(nil)

func (s *Symbol) ID() ulid.ULID {
	return s.Spec.GetID()
}

func (s *Symbol) Kind() string {
	return s.Spec.GetKind()
}

func (s *Symbol) Namespace() string {
	return s.Spec.GetNamespace()
}

func (s *Symbol) Name() string {
	return s.Spec.GetName()
}

func (s *Symbol) Links() map[string][]scheme.PortLocation {
	return s.Spec.GetLinks()
}

func (s *Symbol) Port(name string) (*port.Port, bool) {
	return s.Node.Port(name)

}

func (s *Symbol) Close() error {
	return s.Node.Close()
}
