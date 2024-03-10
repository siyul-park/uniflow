package system

import (
	"context"
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
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
)

// BridgeNode represents a node for executing internal calls.
type BridgeNode struct {
	*node.OneToOneNode
	fn        reflect.Value
	lang      string
	arguments []func(any) (any, error)
	mu        sync.RWMutex
}

// BridgeNodeSpec holds the specifications for creating a BridgeNode.
type BridgeNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Opcode          string   `map:"opcode"`
	Lang            string   `map:"lang"`
	Arguments       []string `map:"arguments"`
}

// BridgeTable represents a table of system call operations.
type BridgeTable struct {
	data map[string]any
	mu   sync.RWMutex
}

const KindBridge = "bridge"

var ErrInvalidOperation = errors.New("operation is invalid")

// NewBridgeNode creates a new BridgeNode with the provided function.
// It returns an error if the provided function is not valid.
func NewBridgeNode(fn any) (*BridgeNode, error) {
	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	n := &BridgeNode{fn: rfn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

// SetArguments sets the language and arguments for the BridgeNode.
// It processes the arguments based on the specified language.
func (n *BridgeNode) SetArguments(lang string, arguments ...string) error {
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

func (n *BridgeNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	inPayload := inPck.Payload()
	input := inPayload.Interface()

	ins := make([]reflect.Value, n.fn.Type().NumIn())

	offset := 0
	if n.fn.Type().NumIn() > 0 {
		if n.fn.Type().In(0).Implements(contextInterface) {
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-proc.Done()
				cancel()
			}()

			ins[0] = reflect.ValueOf(ctx)
			offset++
		}
	}

	for i := offset; i < len(ins); i++ {
		in := reflect.New(n.fn.Type().In(i))

		if len(n.arguments) > i-offset {
			if argument, err := n.arguments[i-offset](input); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if argument, err := primitive.MarshalBinary(argument); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if err := primitive.Unmarshal(argument, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		} else if i == offset {
			if argument, err := primitive.MarshalBinary(input); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if err := primitive.Unmarshal(argument, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}

		ins[i] = in.Elem()
	}

	outs := n.fn.Call(ins)

	if n.fn.Type().NumOut() > 0 && n.fn.Type().Out(n.fn.Type().NumOut()-1).Implements(errorInterface) {
		last := outs[len(outs)-1].Interface()
		outs = outs[:len(outs)-1]

		if err, ok := last.(error); ok {
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

// NewBridgeTable creates a new BridgeTable instance.
func NewBridgeTable() *BridgeTable {
	return &BridgeTable{
		data: make(map[string]any),
	}
}

// Store adds or updates a system call opcode in the table.
func (t *BridgeTable) Store(opcode string, fn any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[opcode] = fn
}

// Load retrieves a system call opcode from the table.
// It returns the opcode function and a boolean indicating if the opcode exists.
func (t *BridgeTable) Load(opcode string) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fn, ok := t.data[opcode]
	return fn, ok
}

// NewBridgeNodeCodec creates a new codec for BridgeNodeSpec.
func NewBridgeNodeCodec(table *BridgeTable) scheme.Codec {
	return scheme.CodecWithType[*BridgeNodeSpec](func(spec *BridgeNodeSpec) (node.Node, error) {
		fn, ok := table.Load(spec.Opcode)
		if !ok {
			return nil, errors.WithStack(ErrInvalidOperation)
		}
		n, err := NewBridgeNode(fn)
		if err != nil {
			return nil, err
		}
		if len(spec.Arguments) > 0 {
			if err := n.SetArguments(spec.Lang, spec.Arguments...); err != nil {
				_ = n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
