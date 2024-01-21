package control

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/xiatechs/jsonata-go"
)

type SwitchNode struct {
	*node.OneToManyNode
	lang    string
	matches []func(any) (bool, error)
	ports   []int
	mu      sync.RWMutex
}

var _ node.Node = (*SwitchNode)(nil)

func NewSwitchNode(lang string) *SwitchNode {
	n := &SwitchNode{lang: lang}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

func (n *SwitchNode) Add(match, port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch n.lang {
	case LangJSONata:
		exp, err := jsonata.Compile(match)
		if err != nil {
			return err
		}
		index, ok := node.IndexOfMultiPort(node.PortOut, port)
		if !ok {
			return errors.WithStack(node.ErrUnsupportedPort)
		}

		n.matches = append(n.matches, func(input any) (bool, error) {
			if output, err := exp.Eval(input); err != nil {
				return false, err
			} else {
				output, _ := output.(bool)
				return output, nil
			}
		})
		n.ports = append(n.ports, index)
	default:
		return errors.WithStack(ErrUnsupportedLanguage)
	}
	return nil
}

func (n *SwitchNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := inPayload.Interface()

	outPcks := make([]*packet.Packet, len(n.matches))
	for i, match := range n.matches {
		port := n.ports[i]
		if ok, err := match(input); err != nil {
			return nil, packet.WithError(err, inPck)
		} else if ok {
			outPcks[port] = inPck
			break
		}
	}
	return outPcks, nil
}
