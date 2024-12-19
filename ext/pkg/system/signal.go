package system

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SignalNodeSpec defines the specifications for creating a SignalNode.
type SignalNodeSpec struct {
	spec.Meta `map:",inline"`
	OPCode    string `map:"opcode"`
}

// SignalNode listens to a signal channel and forwards signals as packets.
// It supports restarting after shutdown.
type SignalNode struct {
	outPort *port.OutPort
	signal  <-chan any
	done    chan struct{}
	close   chan struct{}
	mu      sync.RWMutex
}

const KindSignal = "signal"

// NewSignalNodeCodec creates a codec for compiling SignalNodeSpec into SignalNode instances.
func NewSignalNodeCodec(signals map[string]func(context.Context) (<-chan any, error)) scheme.Codec {
	if signals == nil {
		signals = make(map[string]func(context.Context) (<-chan any, error))
	}

	return scheme.CodecWithType[*SignalNodeSpec](func(spec *SignalNodeSpec) (node.Node, error) {
		fn, ok := signals[spec.OPCode]
		if !ok {
			return nil, errors.WithStack(ErrInvalidOperation)
		}

		ctx, cancel := context.WithCancel(context.Background())

		signal, err := fn(ctx)
		if err != nil {
			cancel()
			return nil, err
		}

		n := NewSignalNode(signal)

		go func() {
			<-n.Wait()
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
		done:    nil,
		close:   make(chan struct{}),
	}
}

// Listen starts listening to the signal channel and emits packets based on received signals.
func (n *SignalNode) Listen() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.done != nil {
		return
	}

	done := make(chan struct{})
	n.done = done

	go func() {
		defer func() {
			n.mu.Lock()
			defer n.mu.Unlock()

			if n.done != nil {
				close(n.done)
				n.done = nil
			}
		}()

		for {
			select {
			case sig, ok := <-n.signal:
				if !ok {
					return
				}

				n.emit(sig)
			case <-done:
				return
			}
		}
	}()
}

// Shutdown stops the SignalNode, allowing it to be restarted with Listen().
func (n *SignalNode) Shutdown() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.done == nil {
		return
	}

	close(n.done)
	n.done = nil
}

// Done returns a channel that is closed when the node is shutdown.
func (n *SignalNode) Done() <-chan struct{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.done
}

// Wait returns a channel that is closed when the node is close.
func (n *SignalNode) Wait() <-chan struct{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.close
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
		case <-n.close:
		default:
			close(n.close)
		}
	}()

	n.Shutdown()
	n.outPort.Close()
	return nil
}

func (n *SignalNode) emit(sig any) {
	var err error

	proc := process.New()
	defer proc.Exit(err)

	writer := n.outPort.Open(proc)
	defer writer.Close()

	payload, err := types.Marshal(sig)
	if err != nil {
		return
	}

	outPck := packet.New(payload)
	backPck := packet.Send(writer, outPck)

	if v, ok := backPck.Payload().(types.Error); ok {
		err = v.Unwrap()
	}
}
