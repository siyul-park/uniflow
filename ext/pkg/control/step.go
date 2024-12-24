package control

import (
	"fmt"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// StepNodeSpec defines the specification for creating a StepNode.
type StepNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []*spec.Unstructured `map:"specs"`
}

const KindStep = "step"

// NewStepNodeCodec creates a new codec for StepNodeSpec.
func NewStepNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(root *StepNodeSpec) (node.Node, error) {
		symbols := make([]*symbol.Symbol, 0, len(root.Specs))
		for _, sp := range root.Specs {
			if sp.GetNamespace() == "" {
				sp.SetNamespace(fmt.Sprintf("%s/%s", root.GetNamespace(), root.GetID()))
			}

			sp, err := s.Decode(sp)
			if err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			}

			n, err := s.Compile(sp)
			if err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			}

			sp.SetPorts(nil)

			symbols = append(symbols, &symbol.Symbol{
				Spec: sp,
				Node: n,
			})
		}

		for i := 0; i < len(symbols)-1; i++ {
			curr := symbols[i]
			next := symbols[i+1]

			curr.SetPorts(map[string][]spec.Port{
				node.PortOut: {
					{
						ID:   next.ID(),
						Port: node.PortIn,
					},
				},
			})
		}

		cluster := symbol.NewCluster(symbols)

		if len(symbols) > 1 {
			first := symbols[0]
			last := symbols[len(symbols)-1]

			cluster.Inbound(node.PortIn, spec.Port{
				ID:   first.ID(),
				Port: node.PortIn,
			})
			cluster.Outbound(node.PortOut, spec.Port{
				ID:   last.ID(),
				Port: node.PortOut,
			})
		}

		for _, sb := range symbols {
			cluster.Outbound(node.PortError, spec.Port{
				ID:   sb.ID(),
				Port: node.PortError,
			})
		}

		return cluster, nil
	})
}
