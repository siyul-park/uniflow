package control

import (
	"context"
	"time"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// ReduceNodeSpec defines the specifications for creating a ReduceNode.
type ReduceNodeSpec struct {
	spec.Meta `map:",inline"`
	Action    string        `map:"action"`
	Init      any           `map:"init,omitempty"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// ReduceNode performs a reduction operation using the provided action.
type ReduceNode struct {
	action  func(context.Context, any, any, int) (any, error)
	init    any
	tracer  *packet.Tracer
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
}

const KindReduce = "reduce"

// NewReduceNodeCodec creates a codec for decoding ReduceNodeSpec.
func NewReduceNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *ReduceNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.Action)
		if err != nil {
			return nil, err
		}

		return NewReduceNode(func(ctx context.Context, acc, cur any, index int) (any, error) {
			if spec.Timeout != 0 {
				var cancel func()
				ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
				defer cancel()
			}

			res, err := program.Run(ctx, []any{acc, cur, index})
			if err != nil {
				return nil, err
			}
			if len(res) == 0 {
				return nil, nil
			}
			return res[0], nil
		}, spec.Init), nil
	})
}

// NewReduceNode creates a new ReduceNode with the provided action and initial value.
func NewReduceNode(action func(context.Context, any, any, int) (any, error), init any) *ReduceNode {
	n := &ReduceNode{
		action:  action,
		init:    init,
		tracer:  packet.NewTracer(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *ReduceNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output or error port based on the name.
func (n *ReduceNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortError:
		return n.errPort
	default:
		return nil
	}
}

func (n *ReduceNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *ReduceNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	acc := n.init
	for i := 0; ; i++ {
		inPck, ok := <-inReader.Read()
		if !ok {
			break
		}

		n.tracer.Read(inReader, inPck)
		cur := types.InterfaceOf(inPck.Payload())

		if v, err := n.action(proc.Context(), acc, cur, i); err != nil {
			errPck := packet.New(types.NewError(err))
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else if outPayload, err := types.Marshal(v); err != nil {
			errPck := packet.New(types.NewError(err))
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else {
			acc = v
			outPck := packet.New(outPayload)
			n.tracer.Transform(inPck, outPck)
			n.tracer.Write(outWriter, outPck)
		}
	}
}

func (n *ReduceNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *ReduceNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
