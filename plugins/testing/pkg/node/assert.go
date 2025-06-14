package node

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
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

// AssertNodeCodec implements scheme.Codec for AssertNode
type AssertNodeCodec struct {
	compiler language.Compiler
	agent    *runtime.Agent
}

const KindAssert = "assert"

var ErrAssertFail = errors.New("assert failed")

var (
	_ node.Node    = (*AssertNode)(nil)
	_ scheme.Codec = (*AssertNodeCodec)(nil)
)

// NewAssertNodeCodec creates a codec for AssertNode
func NewAssertNodeCodec(compiler language.Compiler, agent *runtime.Agent) *AssertNodeCodec {
	return &AssertNodeCodec{
		compiler: compiler,
		agent:    agent,
	}
}

func (c *AssertNodeCodec) Compile(spec spec.Spec) (node.Node, error) {
	converted, ok := spec.(*AssertNodeSpec)
	if !ok {
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	}

	program, err := c.compiler.Compile(converted.Expect)
	if err != nil {
		return nil, err
	}

	n := NewAssertNode(language.Predicate[any](language.Timeout(program, converted.Timeout)))

	if converted.Target != nil {
		n.SetTarget(c.Target(spec.GetNamespace(), converted.Target))
	}

	return n, nil
}

func (c *AssertNodeCodec) Target(namespace string, target *spec.Port) func(proc *process.Process, payload any, index int) (any, int, error) {
	return func(proc *process.Process, payload any, index int) (any, int, error) {
		if index < 0 {
			index = 0
		}

		frames := c.agent.Frames(proc.ID())
		for i := index; i < len(frames); i++ {
			frame := frames[i]
			sym := frame.Symbol
			if sym.Namespace() != namespace || (target.ID != sym.ID() && (target.Name == "" || target.Name != sym.Name())) {
				continue
			}

			if frame.InPort != nil && frame.InPort == sym.In(target.Port) {
				if frame.InPck == nil {
					return nil, 0, errors.WithStack(ErrAssertFail)
				}
				return types.InterfaceOf(frame.InPck.Payload()), i, nil
			}
			if frame.OutPort != nil && frame.OutPort == sym.Out(target.Port) {
				if frame.OutPck == nil {
					return nil, 0, errors.WithStack(ErrAssertFail)
				}
				return types.InterfaceOf(frame.OutPck.Payload()), i, nil
			}
		}
		return nil, 0, errors.WithStack(ErrAssertFail)
	}
}

// NewAssertNode creates a new Assert node
func NewAssertNode(expect func(context.Context, any) (bool, error)) *AssertNode {
	n := &AssertNode{expect: expect}
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

	if ok, err := n.expect(proc, payload); err != nil {
		return nil, packet.New(types.NewError(err))
	} else if !ok {
		return nil, packet.New(types.NewError(ErrAssertFail))
	}

	outPayload, err := types.Marshal([]any{payload, index})
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(outPayload), nil
}
