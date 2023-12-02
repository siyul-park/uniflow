package controllx

import (
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
)

// SwitchNodeConfig holds the configuration for creating a SwitchNode.
type SwitchNodeConfig struct {
	ID ulid.ULID
}

// SwitchNode represents a node that switches packets based on conditions.
type SwitchNode struct {
	*node.OneToManyNode
	conditions []condition
	mu         sync.RWMutex
}

// SwitchSpec represents the specification for the SwitchNode.
type SwitchSpec struct {
	scheme.SpecMeta `map:",inline"`
	Match           []Condition `map:"match"`
}

// Condition represents a condition for the SwitchNode.
type Condition struct {
	When string `map:"when"`
	Port string `map:"port"`
}

type condition struct {
	when *jsonata.Expr
	port string
}

// KindSwitch is the kind identifier for SwitchNode.
const KindSwitch = "switch"

var _ node.Node = (*SwitchNode)(nil)
var _ scheme.Spec = (*SwitchSpec)(nil)

// NewSwitchNode creates a new SwitchNode with the given configuration.
func NewSwitchNode(config SwitchNodeConfig) *SwitchNode {
	id := config.ID

	n := &SwitchNode{}
	n.OneToManyNode = node.NewOneToManyNode(node.OneToManyNodeConfig{
		ID:     id,
		Action: n.action,
	})

	return n
}

// Add adds a new condition to the SwitchNode.
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

// Close closes the SwitchNode and clears its conditions.
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
		if output, _ := cond.when.Eval(input); output != nil && !reflect.ValueOf(output).IsZero() {
			if i, ok := port.GetIndex(node.PortOut, cond.port); ok {
				outPcks := make([]*packet.Packet, i+1)
				outPcks[i] = inPck

				return outPcks, nil
			}
		}
	}

	return nil, packet.WithError(node.ErrDiscardPacket, inPck)
}
