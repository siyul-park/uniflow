package event

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// TriggerNode represents a node that triggers events.
type TriggerNode struct {
	producer *Producer
	consumer *Consumer
	done     chan struct{}
	inPort   *port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// TriggerNodeSpec holds the specifications for creating a TriggerNode.
type TriggerNodeSpec struct {
	spec.Meta `map:",inline"`
	Topic     string `map:"topic"`
}

const KindTrigger = "trigger"

const (
	TopicLoad   = "load"
	TopicUnload = "unload"
)

var _ node.Node = (*TriggerNode)(nil)

// NewTriggerNode creates a new TriggerNode instance.
func NewTriggerNode(producer *Producer, consumer *Consumer) *TriggerNode {
	n := &TriggerNode{
		producer: producer,
		consumer: consumer,
		done:     make(chan struct{}),
		inPort:   port.NewIn(),
		outPort:  port.NewOut(),
		errPort:  port.NewOut(),
	}

	n.inPort.Accept(port.ListenFunc(n.forward))

	return n
}

// In returns the input port with the specified name.
func (n *TriggerNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *TriggerNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Listen listens for incoming events and triggers processing.
func (n *TriggerNode) Listen() {
	n.mu.Lock()
	defer n.mu.Unlock()

	select {
	case <-n.done:
		n.done = make(chan struct{})
	default:
	}

	done := n.done
	go func() {
		for {
			var e *Event
			var ok bool
			select {
			case e, ok = <-n.consumer.Consume():
			case <-done:
			}
			if !ok {
				return
			}

			proc := process.New()

			outWriter := n.outPort.Open(proc)
			errWriter := n.errPort.Open(proc)

			if outPayload, err := object.MarshalText(e.Data()); err != nil {
				errPck := packet.New(object.NewError(err))
				packet.Call(errWriter, errPck)
			} else {
				outPck := packet.New(outPayload)
				packet.Call(outWriter, outPck)
			}

			proc.Wait()
			proc.Exit(nil)
			e.Close()
		}
	}()
}

// Shutdown shuts down the trigger node.
func (n *TriggerNode) Shutdown() {
	n.mu.Lock()
	defer n.mu.Unlock()

	select {
	case <-n.done:
	default:
		close(n.done)
	}
}

// Close closes all ports associated with the node.
func (n *TriggerNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	select {
	case <-n.done:
	default:
		close(n.done)
	}

	n.consumer.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	for e := range n.consumer.Consume() {
		e.Close()
	}

	return nil
}

func (n *TriggerNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()

		e := New(object.InterfaceOf(inPayload))
		n.producer.Produce(e)

		inReader.Receive(packet.None)
	}
}

// NewTriggerNodeCodec creates a new codec for TriggerNodeSpec.
func NewTriggerNodeCodec(upsteam, downsteam *Broker) scheme.Codec {
	return scheme.CodecWithType(func(spec *TriggerNodeSpec) (node.Node, error) {
		p := upsteam.Producer(spec.Topic)
		c := downsteam.Consumer(spec.Topic)

		return NewTriggerNode(p, c), nil
	})
}
