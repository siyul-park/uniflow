package control

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/xiatechs/jsonata-go"
)

// SwitchNode represents a switch node that directs incoming packets based on specified conditions.
type SwitchNode struct {
	*node.OneToManyNode
	lang  string
	whens []func(any) (bool, error)
	ports []int
	mu    sync.RWMutex
}

// SwitchNodeSpec holds the specifications for creating a SwitchNode.
type SwitchNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string      `map:"lang"`
	Match           []Condition `map:"match"`
}

// Condition represents a condition for directing packets to specific ports.
type Condition struct {
	When string `map:"when"`
	Port string `map:"port"`
}

const KindSwitch = "swtich"

// NewSwitchNode creates a new SwitchNode with the specified language.
func NewSwitchNode(lang string) *SwitchNode {
	n := &SwitchNode{lang: lang}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

// Add adds a condition to the SwitchNode, associating it with a specific output port.
func (n *SwitchNode) Add(when, port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if when == "" {
		when = "true"
	}

	index, ok := node.IndexOfMultiPort(node.PortOut, port)
	if !ok {
		return errors.WithStack(node.ErrUnsupportedPort)
	}

	switch n.lang {
	case language.Javascript, language.Typescript:
		var err error
		if n.lang == language.Typescript {
			if when, err = js.Transform(when, api.TransformOptions{Loader: api.LoaderTS}); err != nil {
				return err
			}
		}

		code := fmt.Sprintf("module.exports = ($) => { return %s }", when)
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

		n.whens = append(n.whens, func(input any) (bool, error) {
			vm := vms.Get().(*goja.Runtime)
			defer vms.Put(vm)

			defaults := js.Export(vm, "default")
			when, _ := goja.AssertFunction(defaults)

			if output, err := when(goja.Undefined(), vm.ToValue(input)); err != nil {
				return false, err
			} else {
				output := output.ToBoolean()
				return output, nil
			}
		})
	case language.JSONata:
		exp, err := jsonata.Compile(when)
		if err != nil {
			return err
		}
		n.whens = append(n.whens, func(input any) (bool, error) {
			if output, err := exp.Eval(input); err != nil {
				return false, err
			} else {
				output, _ := output.(bool)
				return output, nil
			}
		})
	default:
		return errors.WithStack(language.ErrUnsupportedLanguage)
	}

	n.ports = append(n.ports, index)
	return nil
}

func (n *SwitchNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := inPayload.Interface()

	outPcks := make([]*packet.Packet, len(n.whens))
	for i, when := range n.whens {
		port := n.ports[i]
		if ok, err := when(input); err != nil {
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
		n := NewSwitchNode(spec.Lang)
		for _, condition := range spec.Match {
			if err := n.Add(condition.When, condition.Port); err != nil {
				_ = n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
