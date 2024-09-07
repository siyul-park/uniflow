package system

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// NativeNodeSpec specifies the creation parameters for a NativeNode.
type NativeNodeSpec struct {
	spec.Meta `map:",inline"`
	OPCode    string `map:"opcode"`
}

// NativeTable stores system call operations.
type NativeTable struct {
	data map[string]any
	mu   sync.RWMutex
}

// NativeNode executes internal system call operations.
type NativeNode struct {
	*node.OneToOneNode
	operator reflect.Value
}

const KindNative = "native"

var ErrInvalidOperation = errors.New("operation is invalid")

var (
	typeContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeError   = reflect.TypeOf((*error)(nil)).Elem()
)

// NewNativeNodeCodec returns a codec for NativeNodeSpec.
func NewNativeNodeCodec(table *NativeTable) scheme.Codec {
	return scheme.CodecWithType(func(spec *NativeNodeSpec) (node.Node, error) {
		fn, err := table.Load(spec.OPCode)
		if err != nil {
			return nil, err
		}
		return NewNativeNode(fn)
	})
}

// NewNativeTable creates a new NativeTable.
func NewNativeTable() *NativeTable {
	return &NativeTable{
		data: make(map[string]any),
	}
}

// Store adds or updates an operation in the table.
func (t *NativeTable) Store(opcode string, fn any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[opcode] = fn
}

// Load retrieves an operation from the table.
func (t *NativeTable) Load(opcode string) (any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fn, ok := t.data[opcode]
	if !ok {
		return nil, ErrInvalidOperation
	}
	return fn, nil
}

// NewNativeNode creates a new NativeNode from a function.
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
	ctx := proc.Context()
	inPayload := inPck.Payload()

	ins := make([]reflect.Value, n.operator.Type().NumIn())
	offset := 0

	if n.operator.Type().NumIn() > 0 && n.operator.Type().In(0).Implements(typeContext) {
		ins[0] = reflect.ValueOf(ctx)
		offset++
	}

	if remains := len(ins) - offset; remains == 1 {
		in := reflect.New(n.operator.Type().In(offset))
		if err := types.Unmarshal(inPayload, in.Interface()); err != nil {
			return nil, packet.New(types.NewError(err))
		}
		ins[offset] = in.Elem()
	} else if remains > 1 {
		var arguments []types.Value
		if v, ok := inPayload.(types.Slice); ok {
			arguments = v.Values()
		} else {
			arguments = append(arguments, v)
		}

		for i := offset; i < len(ins); i++ {
			in := reflect.New(n.operator.Type().In(i))
			if err := types.Unmarshal(arguments[i-offset], in.Interface()); err != nil {
				return nil, packet.New(types.NewError(err))
			}
			ins[i] = in.Elem()
		}
	}

	outs := n.operator.Call(ins)

	if n.operator.Type().NumOut() > 0 && n.operator.Type().Out(n.operator.Type().NumOut()-1).Implements(typeError) {
		last := outs[len(outs)-1].Interface()
		outs = outs[:len(outs)-1]

		if err, ok := last.(error); ok && err != nil {
			return nil, packet.New(types.NewError(err))
		}
	}

	outPayloads := make([]types.Value, len(outs))
	for i, out := range outs {
		if outPayload, err := types.Marshal(out.Interface()); err != nil {
			return nil, packet.New(types.NewError(err))
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
	return packet.New(types.NewSlice(outPayloads...)), nil
}
