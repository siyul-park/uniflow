package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// SequentialNodeSpec defines the specification for creating a SequentialNode.
type SequentialNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []spec.Spec `map:"specs"`
}

const KindSequential = "sequential"

// NewSequentialNodeCodec creates a new codec for SequentialNodeSpec.
func NewSequentialNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(sp *SequentialNodeSpec) (node.Node, error) {
		symbols := make([]*symbol.Symbol, 0, len(sp.Specs))
		for _, sp := range sp.Specs {
			sp, err := s.Decode(sp)
			if err != nil {
				for _, sb := range symbols {
					sb.Close()
				}
				return nil, err
			}

			n, err := s.Compile(sp)
			if err != nil {
				for _, sb := range symbols {
					sb.Close()
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
