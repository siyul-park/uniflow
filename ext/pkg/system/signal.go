package system

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SignalNodeSpec defines the specifications for creating a SignalNode.
type SignalNodeSpec struct {
	spec.Meta `json:",inline"`
	Topic     string `json:"topic" validate:"required"`
}

// SignalNode listens to a signal channel and forwards signals as packets.
// It supports restarting after shutdown.
type SignalNode struct {
	outPort *port.OutPort
	signal  <-chan any
	wait    chan struct{}
	done    chan struct{}
	mu      sync.RWMutex
}

const KindSignal = "signal"

var ErrInvalidTopic = errors.New("topic is invalid")

// NewSignalNodeCodec creates a codec for compiling SignalNodeSpec into SignalNode instances.
func NewSignalNodeCodec(signals map[string]func(context.Context) (<-chan any, error)) scheme.Codec {
	if signals == nil {
		signals = make(map[string]func(context.Context) (<-chan any, error))
	}

	return scheme.CodecWithType[*SignalNodeSpec](func(spec *SignalNodeSpec) (node.Node, error) {
		fn, ok := signals[spec.Topic]
		if !ok {
			return nil, errors.WithStack(ErrInvalidTopic)
		}

		ctx, cancel := context.WithCancel(context.Background())

		signal, err := fn(ctx)
		if err != nil {
			cancel()
			return nil, err
		}

		n := NewSignalNode(signal)

		go func() {
			<-n.Done()
			cancel()
		}()

		return n, nil
	})
}

// NewSignalNode creates a new SignalNode instance.
func NewSignalNode(signal <-chan any) *SignalNode {
	return &SignalNode{
		outPort: port.NewOut(),
		signal:  signal,
		wait:    nil,
		done:    make(chan struct{}),
	}
}

// Listen starts listening to the signal channel and emits packets based on received signals.
func (n *SignalNode) Listen() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.wait != nil {
		return
	}

	wait := make(chan struct{})
	n.wait = wait

	go func() {
		defer func() {
			n.mu.Lock()
			defer n.mu.Unlock()

			if n.wait != nil {
				close(n.wait)
				n.wait = nil
			}
		}()

		for {
			select {
			case sig, ok := <-n.signal:
				if !ok {
					return
				}
				n.emit(sig)
			case <-wait:
				return
			}
		}
	}()
}

// Shutdown stops the SignalNode, allowing it to be restarted with Listen().
func (n *SignalNode) Shutdown() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.wait == nil {
		return
	}

	close(n.wait)
	n.wait = nil
}

// Done returns a channel that is closed when the node is done.
func (n *SignalNode) Done() <-chan struct{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.done
}

// In returns nil as SignalNode does not have input ports.
func (n *SignalNode) In(_ string) *port.InPort {
	return nil
}

// Out returns the output port for the given name.
func (n *SignalNode) Out(name string) *port.OutPort {
	if name == node.PortOut {
		return n.outPort
	}
	return nil
}

// Close stops the node, releases resources, and makes it unusable.
func (n *SignalNode) Close() error {
	func() {
		n.mu.Lock()
		defer n.mu.Unlock()

		select {
		case <-n.done:
		default:
			close(n.done)
		}
	}()

	n.Shutdown()
	n.outPort.Close()
	return nil
}

func (n *SignalNode) emit(sig any) {
	proc := process.New()
	proc.Exit(func() error {
		writer := n.outPort.Open(proc)
		defer writer.Close()

		payload, err := types.Marshal(sig)
		if err != nil {
			return err
		}

		outPck := packet.New(payload)
		backPck := packet.Send(writer, outPck)

		if v, ok := backPck.Payload().(types.Error); ok {
			return v.Unwrap()
		}
		return nil
	}())
}
