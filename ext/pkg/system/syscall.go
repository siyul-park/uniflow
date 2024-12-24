package system

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SyscallNodeSpec specifies the creation parameters for a SyscallNode.
type SyscallNodeSpec struct {
	spec.Meta `map:",inline"`
	OPCode    string `map:"opcode" validate:"required"`
}

// SyscallNode executes synchronized function.
type SyscallNode struct {
	*node.OneToOneNode
	fn func(context.Context, []any) ([]any, error)
}

const KindSyscall = "syscall"

// NewSyscallNodeCodec returns a codec for SyscallNodeSpec.
func NewSyscallNodeCodec(functions map[string]func(ctx context.Context, arguments []any) ([]any, error)) scheme.Codec {
	if functions == nil {
		functions = make(map[string]func(ctx context.Context, arguments []any) ([]any, error))
	}

	return scheme.CodecWithType[*SyscallNodeSpec](func(spec *SyscallNodeSpec) (node.Node, error) {
		fn, ok := functions[spec.OPCode]
		if !ok {
			return nil, errors.WithStack(ErrInvalidOperation)
		}

		return NewSyscallNode(fn)
	})
}

// NewSyscallNode creates a new SyscallNode from a function.
func NewSyscallNode(fn func(context.Context, []any) ([]any, error)) (*SyscallNode, error) {
	n := &SyscallNode{fn: fn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n, nil
}

func (n *SyscallNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
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
