package control

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SwitchNode represents a switch node that directs incoming packets based on specified conditions.
type SwitchNode struct {
	*node.OneToManyNode
	whens []func(any) (bool, error)
	ports []int
	mu    sync.RWMutex
}

// SwitchNodeSpec holds the specifications for creating a SwitchNode.
type SwitchNodeSpec struct {
	spec.Meta `map:",inline"`
	Match     []Condition `map:"match"`
}

// Condition represents a condition for directing packets to specific ports.
type Condition struct {
	When string `map:"when"`
	Port string `map:"port"`
}

const KindSwitch = "switch"

// NewSwitchNode creates a new SwitchNode with the specified language.
func NewSwitchNode() *SwitchNode {
	n := &SwitchNode{}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

// AddMatch adds a condition to the SwitchNode, associating it with a specific output port.
func (n *SwitchNode) AddMatch(when func(any) (bool, error), port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	index, ok := node.IndexOfPort(node.PortOut, port)
	if !ok {
		return errors.WithStack(node.ErrUnsupportedPort)
	}

	n.whens = append(n.whens, when)
	n.ports = append(n.ports, index)
	return nil
}

func (n *SwitchNode) action(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

	outPcks := make([]*packet.Packet, len(n.whens))
	for i, when := range n.whens {
		port := n.ports[i]
		if ok, err := when(input); err != nil {
			return nil, packet.New(types.NewError(err))
		} else if ok {
			outPcks[port] = inPck
			break
		}
	}

	return outPcks, nil
}

// NewSwitchNodeCodec creates a new codec for SwitchNodeSpec.
func NewSwitchNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *SwitchNodeSpec) (node.Node, error) {
		whens := make([]func(any) (bool, error), len(spec.Match))
		for i, condition := range spec.Match {
			program, err := compiler.Compile(condition.When)
			if err != nil {
				return nil, err
			}

			whens[i] = func(env any) (bool, error) {
				res, err := program.Run(env)
				if err != nil {
					return false, err
				}
				return !reflect.ValueOf(res).IsZero(), nil
			}
		}

		n := NewSwitchNode()
		for i, condition := range spec.Match {
			if err := n.AddMatch(whens[i], condition.Port); err != nil {
				_ = n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
