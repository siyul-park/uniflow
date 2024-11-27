package system

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"reflect"
)

// NativeNodeCodec maps opcode strings to their corresponding function implementations for encoding and decoding native operations.
type NativeNodeCodec struct {
	operators map[string]any
}

// NativeNodeSpec specifies the creation parameters for a NativeNode.
type NativeNodeSpec struct {
	spec.Meta `map:",inline"`
	OPCode    string `map:"opcode"`
}

// NativeNode executes synchronized function.
type NativeNode struct {
	*node.OneToOneNode
	fn func(context.Context, []any) ([]any, error)
}

const KindNative = "native"

var ErrInvalidOperation = errors.New("operation is invalid")

var _ scheme.Codec = (*NativeNodeCodec)(nil)

// NewNativeNodeCodec returns a codec for NativeNodeSpec.
func NewNativeNodeCodec(operators map[string]any) *NativeNodeCodec {
	if operators == nil {
		operators = make(map[string]any)
	}
	return &NativeNodeCodec{operators: operators}
}

// Compile compiles a NativeNodeSpec into a Node instance.
func (n *NativeNodeCodec) Compile(spc spec.Spec) (node.Node, error) {
	if spc, ok := spc.(*NativeNodeSpec); ok {
		fn, err := n.Load(spc.OPCode)
		if err != nil {
			return nil, err
		}
		return NewNativeNode(fn)
	}
	return nil, errors.WithStack(encoding.ErrUnsupportedType)
}

// Load returns a function that can be used for executing the operation associated with the given opcode.
func (n *NativeNodeCodec) Load(opcode string) (func(context.Context, []any) ([]any, error), error) {
	raw := n.operators[opcode]
	if raw == nil {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	fn := reflect.ValueOf(raw)
	if fn.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	typeContext := reflect.TypeOf((*context.Context)(nil)).Elem()
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	opType := fn.Type()
	numIn := opType.NumIn()
	numOut := opType.NumOut()

	return func(ctx context.Context, arguments []any) ([]any, error) {
		ins := make([]reflect.Value, numIn)
		offset := 0

		if numIn > 0 && opType.In(0).Implements(typeContext) {
			ins[0] = reflect.ValueOf(ctx)
			offset++
		}

		for i := offset; i < numIn; i++ {
			if i-offset < len(arguments) {
				arg, err := types.Marshal(arguments[i-offset])
				if err != nil {
					return nil, err
				}
				in := reflect.New(opType.In(i)).Interface()
				if err := types.Unmarshal(arg, in); err != nil {
					return nil, err
				}
				ins[i] = reflect.ValueOf(in).Elem()
			} else {
				ins[i] = reflect.Zero(opType.In(i))
			}
		}

		outs := fn.Call(ins)

		if numOut > 0 && opType.Out(numOut-1).Implements(typeError) {
			if err, ok := outs[numOut-1].Interface().(error); ok && err != nil {
				return nil, err
			}
			outs = outs[:numOut-1]
		}

		returns := make([]any, len(outs))
		for i, out := range outs {
			returns[i] = out.Interface()
		}
		return returns, nil
	}, nil
}

// NewNativeNode creates a new NativeNode from a function.
func NewNativeNode(fn func(context.Context, []any) ([]any, error)) (*NativeNode, error) {
	n := &NativeNode{fn: fn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n, nil
}

func (n *NativeNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	ctx := proc.Context()
	inPayload := inPck.Payload()

	var arguments []any
	if v, ok := inPayload.(types.Slice); ok {
		arguments = v.Slice()
	} else {
		arguments = append(arguments, types.InterfaceOf(inPayload))
	}

	returns, err := n.fn(ctx, arguments)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	outPayloads := make([]types.Value, len(returns))
	for i, out := range returns {
		outPayload, err := types.Marshal(out)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
		outPayloads[i] = outPayload
	}

	if len(outPayloads) == 0 {
		return packet.New(nil), nil
	}
	if len(outPayloads) == 1 {
		return packet.New(outPayloads[0]), nil
	}
	return packet.New(types.NewSlice(outPayloads...)), nil
}
