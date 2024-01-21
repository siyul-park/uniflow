package control

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/plugin/internal/js"
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

	index, ok := node.IndexOfMultiPort(node.PortOut, port)
	if !ok {
		return errors.WithStack(node.ErrUnsupportedPort)
	}

	switch n.lang {
	case LangJavascript:
		code := fmt.Sprintf("module.exports = ($) => { return %s }", match)
		program, err := goja.Compile("", code, true)
		if err != nil {
			return err
		}
		vms := &sync.Pool{
			New: func() any {
				vm := js.New()
				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		n.matches = append(n.matches, func(input any) (bool, error) {
			vm := vms.Get().(*goja.Runtime)
			defer vms.Put(vm)

			defaults := js.Export(vm, "default")
			match, _ := goja.AssertFunction(defaults)

			if output, err := match(goja.Undefined(), vm.ToValue(input)); err != nil {
				return false, err
			} else {
				output := output.ToBoolean()
				return output, nil
			}
		})
	case LangJSONata:
		exp, err := jsonata.Compile(match)
		if err != nil {
			return err
		}
		n.matches = append(n.matches, func(input any) (bool, error) {
			if output, err := exp.Eval(input); err != nil {
				return false, err
			} else {
				output, _ := output.(bool)
				return output, nil
			}
		})
	default:
		return errors.WithStack(ErrUnsupportedLanguage)
	}

	n.ports = append(n.ports, index)
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
