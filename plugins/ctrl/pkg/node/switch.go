package node

import (
	"context"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SwitchNodeSpec holds specifications for creating a SwitchNode.
type SwitchNodeSpec struct {
	spec.Meta `json:",inline"`
	Matches   []Condition   `json:"matches" validate:"required"`
	Timeout   time.Duration `json:"timeout,omitempty"`
}

// Condition represents a condition for directing packets to specific ports.
type Condition struct {
	When string `json:"when" validate:"required"`
	Port string `json:"port" validate:"required"`
}

// SwitchNode directs packets to different ports based on specified conditions.
type SwitchNode struct {
	*node.OneToManyNode
	conditions []func(context.Context, any) (bool, error)
	ports      []int
	mu         sync.RWMutex
}

const KindSwitch = "switch"

// NewSwitchNodeCodec creates a new codec for SwitchNodeSpec.
func NewSwitchNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *SwitchNodeSpec) (node.Node, error) {
		conditions := make([]func(context.Context, any) (bool, error), len(spec.Matches))
		for i, condition := range spec.Matches {
			program, err := compiler.Compile(condition.When)
			if err != nil {
				return nil, err
			}

			conditions[i] = language.Predicate[any](language.Timeout(program, spec.Timeout))
		}

		n := NewSwitchNode()
		for i, condition := range spec.Matches {
			n.Match(condition.Port, conditions[i])
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
func (n *SwitchNode) Match(port string, condition func(context.Context, any) (bool, error)) {
	n.mu.Lock()
	defer n.mu.Unlock()

	index, ok := node.IndexOfPort(port)
	if !ok || node.NameOfPort(port) != node.PortOut {
		return
	}

	n.conditions = append(n.conditions, condition)
	n.ports = append(n.ports, index)
}

func (n *SwitchNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

	outPcks := make([]*packet.Packet, len(n.conditions))
	for i, condition := range n.conditions {
		port := n.ports[i]
		if ok, err := condition(proc, input); err != nil {
			return nil, packet.New(types.NewError(err))
		} else if ok {
			outPcks[port] = inPck
			break
		}
	}

	return outPcks, nil
}
