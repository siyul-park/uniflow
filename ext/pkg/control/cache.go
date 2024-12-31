package control

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// CacheNodeSpec represents the specification for a cache node.
type CacheNodeSpec struct {
	spec.Meta `map:",inline"`
	Capacity  int           `map:"capacity,omitempty"`
	Interval  time.Duration `map:"interval,omitempty"`
}

// CacheNode represents a node in the cache.
type CacheNode struct {
	lru     *types.LRU
	tracer  *packet.Tracer
	inPort  *port.InPort
	outPort *port.OutPort
}

const KindCache = "cache"

var _ node.Node = (*CacheNode)(nil)

// NewCacheNodeCodec creates a new codec for CacheNodeSpec.
func NewCacheNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CacheNodeSpec) (node.Node, error) {
		return NewCacheNode(spec.Capacity, spec.Interval), nil
	})
}

// NewCacheNode creates a new CacheNode with the given capacity and Interval.
func NewCacheNode(capacity int, interval time.Duration) *CacheNode {
	n := &CacheNode{
		lru:     types.NewLRU(capacity, interval),
		tracer:  packet.NewTracer(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))

	return n
}

// In returns the input port for the given name.
func (n *CacheNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port for the given name.
func (n *CacheNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	default:
		return nil
	}
}

// Close closes the CacheNode and its ports.
func (n *CacheNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.tracer.Close()
	n.lru.Clear()
	return nil
}

func (n *CacheNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	var outWriter *packet.Writer

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		inPayload := inPck.Payload()
		if outPayload, ok := n.lru.Load(inPayload); ok {
			outPck := packet.New(outPayload)
			n.tracer.Transform(inPck, outPck)
			n.tracer.Reduce(outPck)
		} else {
			n.tracer.AddHook(inPck, packet.HookFunc(func(backPck *packet.Packet) {
				if _, ok := backPck.Payload().(types.Error); !ok {
					n.lru.Store(inPayload, backPck.Payload())
				}
				n.tracer.Transform(inPck, backPck)
				n.tracer.Reduce(backPck)
			}))

			if outWriter == nil {
				outWriter = n.outPort.Open(proc)
			}
			n.tracer.Write(outWriter, inPck)
		}
	}
}

func (n *CacheNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}
