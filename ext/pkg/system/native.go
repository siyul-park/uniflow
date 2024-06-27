package system

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// NativeNode represents a node for executing internal calls.
type NativeNode struct {
	*node.OneToOneNode
	operator reflect.Value
	mu       sync.RWMutex
}

// NativeNodeSpec holds the specifications for creating a NativeNode.
type NativeNodeSpec struct {
	spec.Meta `map:",inline"`
	Opcode    string `map:"opcode"`
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

func (n *NativeNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inPayload := inPck.Payload()

	ins := make([]reflect.Value, n.operator.Type().NumIn())

	offset := 0
	if n.operator.Type().NumIn() > 0 {
		if n.operator.Type().In(0).Implements(contextInterface) {
			ins[0] = reflect.ValueOf(ctx)
			offset++
		}
	}

	if remains := len(ins) - offset; remains == 1 {
		in := reflect.New(n.operator.Type().In(offset))
		if err := object.Unmarshal(inPayload, in.Interface()); err != nil {
			return nil, packet.New(object.NewError(err))
		}
		ins[offset] = in.Elem()
	} else if remains > 1 {
		var arguments []object.Object
		if v, ok := inPayload.(object.Slice); ok {
			arguments = v.Values()
		} else {
			arguments = append(arguments, v)
		}

		for i := offset; i < len(ins); i++ {
			in := reflect.New(n.operator.Type().In(i))
			if err := object.Unmarshal(arguments[i-offset], in.Interface()); err != nil {
				return nil, packet.New(object.NewError(err))
			}
			ins[i] = in.Elem()
		}
	}

	outs := n.operator.Call(ins)

	if n.operator.Type().NumOut() > 0 && n.operator.Type().Out(n.operator.Type().NumOut()-1).Implements(errorInterface) {
		last := outs[len(outs)-1].Interface()
		outs = outs[:len(outs)-1]

		if err, ok := last.(error); ok {
			if err != nil {
				return nil, packet.New(object.NewError(err))
			}
		}
	}

	outPayloads := make([]object.Object, len(outs))
	for i, out := range outs {
		if outPayload, err := object.MarshalText(out.Interface()); err != nil {
			return nil, packet.New(object.NewError(err))
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
	return packet.New(object.NewSlice(outPayloads...)), nil
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
func (t *NativeModule) Load(opcode string) (any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if fn, ok := t.data[opcode]; !ok {
		return nil, errors.WithStack(ErrInvalidOperation)
	} else {
		return fn, nil
	}
}

// NewNativeNodeCodec creates a new codec for NativeNodeSpec.
func NewNativeNodeCodec(module *NativeModule) scheme.Codec {
	return scheme.CodecWithType(func(spec *NativeNodeSpec) (node.Node, error) {
		fn, err := module.Load(spec.Opcode)
		if err != nil {
			return nil, err
		}
		return NewNativeNode(fn)
	})
}
