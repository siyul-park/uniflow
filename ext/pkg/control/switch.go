package control

import (
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SwitchNodeSpec holds specifications for creating a SwitchNode.
type SwitchNodeSpec struct {
	spec.Meta `map:",inline"`
	Matches   []Condition `map:"matches"`
}

// Condition represents a condition for directing packets to specific ports.
type Condition struct {
	When string `map:"when"`
	Port string `map:"port"`
}

// SwitchNode directs packets to different ports based on specified conditions.
type SwitchNode struct {
	*node.OneToManyNode
	conditions []func(any) (bool, error)
	ports      []int
	mu         sync.RWMutex
}

const KindSwitch = "switch"

// NewSwitchNodeCodec creates a new codec for SwitchNodeSpec.
func NewSwitchNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *SwitchNodeSpec) (node.Node, error) {
		conditions := make([]func(any) (bool, error), len(spec.Matches))
		for i, condition := range spec.Matches {
			program, err := compiler.Compile(condition.When)
			if err != nil {
				return nil, err
			}

			conditions[i] = func(env any) (bool, error) {
				res, err := program.Run(env)
				if err != nil {
					return false, err
				}
				return !reflect.ValueOf(res).IsZero(), nil
			}
		}

		n := NewSwitchNode()
		for i, condition := range spec.Matches {
			n.Match(conditions[i], condition.Port)
		}
		return n, nil
	})
}

// NewSwitchNode creates a new SwitchNode instance.
func NewSwitchNode() *SwitchNode {
	n := &SwitchNode{}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

// Match associates a condition with a specific output port in the SwitchNode.
func (n *SwitchNode) Match(condition func(any) (bool, error), port string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	index, ok := node.IndexOfPort(port)
	if !ok || node.NameOfPort(port) != node.PortOut {
		return
	}

	n.conditions = append(n.conditions, condition)
	n.ports = append(n.ports, index)
}

func (n *SwitchNode) action(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

	outPcks := make([]*packet.Packet, len(n.conditions))
	for i, condition := range n.conditions {
		port := n.ports[i]
		if ok, err := condition(input); err != nil {
			return nil, packet.New(types.NewError(err))
		} else if ok {
			outPcks[port] = inPck
			break
		}
	}

	return outPcks, nil
}
