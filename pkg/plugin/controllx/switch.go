package controllx

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
)

type (
	SwitchNodeConfig struct {
		ID ulid.ULID
	}

	SwitchNode struct {
		*node.OneToManyNode
		conditions []condition
		mu         sync.RWMutex
	}

	SwitchSpec struct {
		scheme.SpecMeta `map:",inline"`
		Match           []Condition `map:"match"`
	}

	Condition struct {
		When string `map:"when"`
		Port string `map:"port"`
	}

	condition struct {
		when *jsonata.Expr
		port string
	}
)

const (
	KindSwitch = "switch"
)

var _ node.Node = &SwitchNode{}

func NewSwitchNode(config SwitchNodeConfig) *SwitchNode {
	id := config.ID

	n := &SwitchNode{}
	n.OneToManyNode = node.NewOneToManyNode(node.OneToManyNodeConfig{
		ID:     id,
		Action: n.action,
	})

	return n
}

func (n *SwitchNode) Add(when string, port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	exp, err := jsonata.Compile(when)
	if err != nil {
		return err
	}

	n.conditions = append(n.conditions, condition{when: exp, port: port})
	return nil
}

func (n *SwitchNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.conditions = nil
	return n.OneToManyNode.Close()
}

func (n *SwitchNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()

	var input any
	if inPayload != nil {
		input = inPayload.Interface()
	}

	for _, cond := range n.conditions {
		if output, _ := cond.when.Eval(input); !util.IsZero(output) {
			if i, ok := port.GetIndex(node.PortOut, cond.port); ok {
				outPcks := make([]*packet.Packet, i+1)
				outPcks[i] = inPck

				return outPcks, nil
			}
		}
	}

	return nil, packet.NewError(node.ErrDiscardPacket, inPck)
}
