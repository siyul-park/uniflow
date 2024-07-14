package system

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SyscallNodeSpec holds the specifications for creating a SyscallNode.
type SyscallNodeSpec struct {
	spec.Meta `map:",inline"`
	OPCode    string `map:"opcode"`
}

// SyscallNode represents a node for executing internal calls.
type SyscallNode struct {
	*node.OneToOneNode
	operator reflect.Value
}

const KindSyscall = "syscall"

var typeContext = reflect.TypeOf((*context.Context)(nil)).Elem()
var typeError = reflect.TypeOf((*error)(nil)).Elem()

// NewSyscallNodeCodec creates a new codec for SyscallNodeSpec.
func NewSyscallNodeCodec(table *Table) scheme.Codec {
	return scheme.CodecWithType(func(spec *SyscallNodeSpec) (node.Node, error) {
		fn, err := table.Load(spec.OPCode)
		if err != nil {
			return nil, err
		}
		return NewSyscallNode(fn)
	})
}

// NewSyscallNode creates a new SyscallNode with the provided function.
// It returns an error if the provided function is not valid.
func NewSyscallNode(operator any) (*SyscallNode, error) {
	op := reflect.ValueOf(operator)
	if op.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	n := &SyscallNode{operator: op}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

func (n *SyscallNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	ctx := proc.Context()

	inPayload := inPck.Payload()

	ins := make([]reflect.Value, n.operator.Type().NumIn())

	offset := 0
	if n.operator.Type().NumIn() > 0 {
		if n.operator.Type().In(0).Implements(typeContext) {
			ins[0] = reflect.ValueOf(ctx)
			offset++
		}
	}

	if remains := len(ins) - offset; remains == 1 {
		in := reflect.New(n.operator.Type().In(offset))
		if err := types.Decoder.Decode(inPayload, in.Interface()); err != nil {
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
			if err := types.Decoder.Decode(arguments[i-offset], in.Interface()); err != nil {
				return nil, packet.New(types.NewError(err))
			}
			ins[i] = in.Elem()
		}
	}

	outs := n.operator.Call(ins)

	if n.operator.Type().NumOut() > 0 && n.operator.Type().Out(n.operator.Type().NumOut()-1).Implements(typeError) {
		last := outs[len(outs)-1].Interface()
		outs = outs[:len(outs)-1]

		if err, ok := last.(error); ok {
			if err != nil {
				return nil, packet.New(types.NewError(err))
			}
		}
	}

	outPayloads := make([]types.Value, len(outs))
	for i, out := range outs {
		if outPayload, err := types.TextEncoder.Encode(out.Interface()); err != nil {
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
