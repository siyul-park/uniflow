package node

import (
	"context"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// AssertNodeSpec defines the specification for Assert node
type AssertNodeSpec struct {
	spec.Meta `json:",inline"`
	Expect    string        `json:"expect"`
	Target    *spec.Port    `json:"target,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"`
}

// AssertNode implements the Assert node functionality
type AssertNode struct {
	*node.OneToOneNode
	expect func(context.Context, any) (bool, error)
	target func(*process.Process, any, int) (any, int, error)
	mu     sync.RWMutex
}

const KindAssert = "assert"

var ErrAssertFail = errors.New("assert failed")

var _ node.Node = (*AssertNode)(nil)

// NewAssertNodeCodec creates a codec for AssertNode
func NewAssertNodeCodec(compiler language.Compiler, agent *runtime.Agent) scheme.Codec {
	return scheme.CodecWithType(func(spec *AssertNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.Expect)
		if err != nil {
			return nil, err
		}

		n := NewAssertNode(language.Predicate[any](language.Timeout(program, spec.Timeout)))

		if spec.Target != nil {
			n.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
				if index < 0 {
					index = 0
				}

				frames := agent.Frames(proc.ID())
				for i := index; i < len(frames); i++ {
					frame := frames[i]
					if frame.Symbol == nil {
						continue
					}

					ok := false
					if spec.Target.ID != uuid.Nil {
						ok = frame.Symbol.ID() == spec.Target.ID
					} else {
						ok = frame.Symbol.Namespace() == spec.GetNamespace() &&
							frame.Symbol.Name() == spec.Target.Name
					}
					if !ok {
						continue
					}

					if frame.InPort != nil && frame.InPort == frame.Symbol.In(spec.Target.Port) {
						if frame.InPck == nil {
							return nil, -1, errors.WithStack(ErrAssertFail)
						}
						return frame.InPck.Payload(), i, nil
					}
					if frame.OutPort != nil && frame.OutPort == frame.Symbol.Out(spec.Target.Port) {
						if frame.OutPck == nil {
							return nil, -1, errors.WithStack(ErrAssertFail)
						}
						return frame.OutPck.Payload(), i, nil
					}
				}

				return nil, -1, errors.WithStack(ErrAssertFail)
			})
		}

		return n, nil
	})
}

// NewAssertNode creates a new Assert node
func NewAssertNode(expect func(context.Context, any) (bool, error)) *AssertNode {
	n := &AssertNode{
		expect: expect,
	}

	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// SetTarget sets the target function
func (n *AssertNode) SetTarget(target func(*process.Process, any, int) (any, int, error)) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.target = target
}

func (n *AssertNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	payload, err := types.Cast[any](types.Lookup(inPayload, 0))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	index, err := types.Cast[int](types.Lookup(inPayload, 1))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	if n.target != nil {
		payload, index, err = n.target(proc, payload, index)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
	}

	value, err := types.Marshal(payload)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	if ok, err := n.expect(proc, types.InterfaceOf(value)); err != nil {
		return nil, packet.New(types.NewError(err))
	} else if !ok {
		return nil, packet.New(types.NewError(ErrAssertFail))
	}

	return packet.New(types.NewSlice(value, types.NewInt(index))), nil
}
