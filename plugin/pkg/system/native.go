package system

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// NativeNode represents a node for executing internal calls.
type NativeNode struct {
	*node.OneToOneNode
	lang     string
	operator reflect.Value
	operands []func(any) (any, error)
	mu       sync.RWMutex
}

// NativeNodeSpec holds the specifications for creating a NativeNode.
type NativeNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string   `map:"lang,omitempty"`
	Opcode          string   `map:"opcode"`
	Operands        []string `map:"operands,omitempty"`
}

// NativeModule represents a table of system call operations.
type NativeModule struct {
	data map[string]any
	mu   sync.RWMutex
}

const KindNative = "native"

var ErrInvalidOperation = errors.New("operation is invalid")

// NewNativeNode creates a new NativeNode with the provided function.
// It returns an error if the provided function is not valid.
func NewNativeNode(operator any) (*NativeNode, error) {
	op := reflect.ValueOf(operator)
	if op.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	n := &NativeNode{operator: op}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

// SetLanguage sets the language for the NativeNode.
func (n *NativeNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

// SetOperands sets operands, it processes the operands based on the specified language.
func (n *NativeNode) SetOperands(operands ...string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.operands = nil

	for _, operand := range operands {
		lang := n.lang
		transform, err := language.CompileTransform(operand, &lang)
		if err != nil {
			return err
		}

		n.operands = append(n.operands, transform)
	}

	return nil
}

func (n *NativeNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	inPayload := inPck.Payload()
	input := primitive.Interface(inPayload)

	ins := make([]reflect.Value, n.operator.Type().NumIn())

	offset := 0
	if n.operator.Type().NumIn() > 0 {
		if n.operator.Type().In(0).Implements(contextInterface) {
			ins[0] = reflect.ValueOf(proc.Context())
			offset++
		}
	}

	for i := offset; i < len(ins); i++ {
		in := reflect.New(n.operator.Type().In(i))

		if len(n.operands) > i-offset {
			if operand, err := n.operands[i-offset](input); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if argument, err := primitive.MarshalText(operand); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if err := primitive.Unmarshal(argument, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		} else if i == offset {
			if err := primitive.Unmarshal(inPayload, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}

		ins[i] = in.Elem()
	}

	outs := n.operator.Call(ins)

	if n.operator.Type().NumOut() > 0 && n.operator.Type().Out(n.operator.Type().NumOut()-1).Implements(errorInterface) {
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
		if outPayload, err := primitive.MarshalText(out.Interface()); err != nil {
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

// NewNativeModule creates a new NativeModule instance.
func NewNativeModule() *NativeModule {
	return &NativeModule{
		data: make(map[string]any),
	}
}

// Store adds or updates a system call opcode in the table.
func (t *NativeModule) Store(opcode string, fn any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[opcode] = fn
}

// Load retrieves a system call opcode from the table.
// It returns the opcode function and a boolean indicating if the opcode exists.
func (t *NativeModule) Load(opcode string) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fn, ok := t.data[opcode]
	return fn, ok
}

// NewNativeNodeCodec creates a new codec for NativeNodeSpec.
func NewNativeNodeCodec(module *NativeModule) scheme.Codec {
	return scheme.CodecWithType(func(spec *NativeNodeSpec) (node.Node, error) {
		fn, ok := module.Load(spec.Opcode)
		if !ok {
			return nil, errors.WithStack(ErrInvalidOperation)
		}
		n, err := NewNativeNode(fn)
		if err != nil {
			return nil, err
		}
		n.SetLanguage(spec.Lang)
		if err := n.SetOperands(spec.Operands...); err != nil {
			_ = n.Close()
			return nil, err
		}
		return n, nil
	})
}
