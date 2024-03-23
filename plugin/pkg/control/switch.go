package control

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// SwitchNode represents a switch node that directs incoming packets based on specified conditions.
type SwitchNode struct {
	*node.OneToManyNode
	lang  string
	whens []func(primitive.Value) (bool, error)
	ports []int
	mu    sync.RWMutex
}

// SwitchNodeSpec holds the specifications for creating a SwitchNode.
type SwitchNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string      `map:"lang,omitempty"`
	Match           []Condition `map:"match"`
}

// Condition represents a condition for directing packets to specific ports.
type Condition struct {
	When string `map:"when"`
	Port string `map:"port"`
}

const KindSwitch = "swtich"

// NewSwitchNode creates a new SwitchNode with the specified language.
func NewSwitchNode() *SwitchNode {
	n := &SwitchNode{}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

// SetLanguage sets the language for the SwitchNode.
func (n *SwitchNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

// AddMatch adds a condition to the SwitchNode, associating it with a specific output port.
func (n *SwitchNode) AddMatch(when, port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if when == "" {
		when = "true"
	}

	lang := n.lang
	transform, err := language.CompileTransform(when, &lang)
	if err != nil {
		return err
	}

	index, ok := node.IndexOfPort(node.PortOut, port)
	if !ok {
		return errors.WithStack(node.ErrUnsupportedPort)
	}

	n.whens = append(n.whens, func(value primitive.Value) (bool, error) {
		var input any
		switch lang {
		case language.Typescript, language.Javascript, language.JSONata:
			input = primitive.Interface(value)
		}

		out, err := transform(input)
		if err != nil {
			return false, err
		}
		return !reflect.ValueOf(out).IsZero(), nil
	})
	n.ports = append(n.ports, index)
	return nil
}

func (n *SwitchNode) action(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()

	outPcks := make([]*packet.Packet, len(n.whens))
	for i, when := range n.whens {
		port := n.ports[i]
		if ok, err := when(inPayload); err != nil {
			return nil, packet.WithError(err, inPck)
		} else if ok {
			outPcks[port] = inPck
			break
		}
	}
	return outPcks, nil
}

// NewSwitchNodeCodec creates a new codec for SwitchNodeSpec.
func NewSwitchNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*SwitchNodeSpec](func(spec *SwitchNodeSpec) (node.Node, error) {
		n := NewSwitchNode()
		n.SetLanguage(spec.Lang)
		for _, condition := range spec.Match {
			if err := n.AddMatch(condition.When, condition.Port); err != nil {
				_ = n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
