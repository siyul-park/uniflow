package system

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v2"
)

type SyscallNode struct {
	*node.OneToOneNode
	fn        reflect.Value
	lang      string
	arguments []func(any) (any, error)
	mu        sync.RWMutex
}

var ErrInvalidOperation = errors.New("operation is invalid")

func NewSyscallNode(fn any) (*SyscallNode, error) {
	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	n := &SyscallNode{fn: rfn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

func (n *SyscallNode) SetArguments(lang string, arguments ...string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
	n.arguments = nil

	for _, argument := range arguments {
		argument := argument

		switch n.lang {
		case language.Text:
			n.arguments = append(n.arguments, func(_ any) (any, error) {
				return argument, nil
			})
		case language.JSON, language.YAML:
			var data any
			var err error
			if n.lang == language.JSON {
				err = json.Unmarshal([]byte(argument), &data)
			} else if n.lang == language.YAML {
				err = yaml.Unmarshal([]byte(argument), &data)
			}
			if err != nil {
				return err
			}

			n.arguments = append(n.arguments, func(_ any) (any, error) {
				return data, nil
			})
		case language.Javascript, language.Typescript:
			var err error
			if n.lang == language.Typescript {
				if argument, err = js.Transform(argument, api.TransformOptions{Loader: api.LoaderTS}); err != nil {
					return err
				}
			}

			code := fmt.Sprintf("module.exports = ($) => { return %s }", argument)
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

			n.arguments = append(n.arguments, func(input any) (any, error) {
				vm := vms.Get().(*goja.Runtime)
				defer vms.Put(vm)

				defaults := js.Export(vm, "default")
				argument, _ := goja.AssertFunction(defaults)

				if output, err := argument(goja.Undefined(), vm.ToValue(input)); err != nil {
					return false, err
				} else {
					return output.Export(), nil
				}
			})
		case language.JSONata:
			exp, err := jsonata.Compile(argument)
			if err != nil {
				return err
			}
			n.arguments = append(n.arguments, func(input any) (any, error) {
				if output, err := exp.Eval(input); err != nil {
					return false, err
				} else {
					return output, nil
				}
			})
		default:
			return errors.WithStack(language.ErrUnsupportedLanguage)
		}
	}

	return nil
}

func (n *SyscallNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := inPayload.Interface()

	ins := make([]reflect.Value, n.fn.Type().NumIn())

	for i := range ins {
		in := reflect.New(n.fn.Type().In(i))
		if len(n.arguments) > i {
			if argument, err := n.arguments[i](input); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if argument, err := primitive.MarshalBinary(argument); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if err := primitive.Unmarshal(argument, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}
		ins[i] = in.Elem()
	}

	outs := n.fn.Call(ins)

	if len(outs) > 0 {
		if err, ok := outs[len(outs)-1].Interface().(error); ok {
			outs = outs[:len(outs)-1]
			if err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}
	}

	outPayloads := make([]primitive.Value, len(outs))
	for i, out := range outs {
		if outPayload, err := primitive.MarshalBinary(out.Interface()); err != nil {
			return nil, packet.WithError(err, inPck)
		} else {
			outPayloads[i] = outPayload
		}
	}

	if len(outPayloads) == 0 {
		return packet.New(nil), nil
	}
	if len(outPayloads) == 1 {
		return packet.New(outPayloads[0]), nil
	}
	return packet.New(primitive.NewSlice(outPayloads...)), nil
}
