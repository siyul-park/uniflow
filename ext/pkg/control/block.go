package control

import (
	"fmt"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// BlockNodeSpec defines the specification for creating a BlockNode.
type BlockNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []*spec.Unstructured   `map:"specs"`
	Inbounds  map[string][]spec.Port `map:"inbounds,omitempty"`
	Outbounds map[string][]spec.Port `map:"outbounds,omitempty"`
}

const KindBlock = "block"

// NewBlockNodeCodec creates a new codec for BlockNodeSpec.
func NewBlockNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(root *BlockNodeSpec) (node.Node, error) {
		symbols := make([]*symbol.Symbol, 0, len(root.Specs))
		for i, sp := range root.Specs {
			if sp.GetNamespace() == "" {
				sp.SetNamespace(root.GetNamespacedName())
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

			if sp.GetName() == "" {
				sp.SetName(fmt.Sprintf("$%d", i))
			}

			symbols = append(symbols, &symbol.Symbol{
				Spec: sp,
				Node: n,
			})
		}

		cluster := symbol.NewCluster(symbols)

		for name, ports := range root.Inbounds {
			for _, port := range ports {
				cluster.Inbound(name, port)
			}
		}
		for name, ports := range root.Outbounds {
			for _, port := range ports {
				cluster.Outbound(name, port)
			}
		}

		return cluster, nil
	})
}
