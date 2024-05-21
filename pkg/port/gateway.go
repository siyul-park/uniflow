package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

type Gateway struct {
	readers []*Reader
	reads   [][]*packet.Packet
	forward ForwardHook
	mu      sync.Mutex
}

type ForwardHook interface {
	Forward(pcks []*packet.Packet) bool
}

type ForwardHookFunc func(pcks []*packet.Packet) bool

var _ ForwardHook = (ForwardHookFunc)(nil)

func NewGateway(readers []*Reader, interceptor ForwardHook) *Gateway {
	return &Gateway{
		readers: readers,
		forward: interceptor,
	}
}

func (g *Gateway) Write(pck *packet.Packet, reader *Reader) int {
	g.mu.Lock()
	defer g.mu.Unlock()

	index := -1
	for i, r := range g.readers {
		if r == reader {
			index = i
			break
		}
	}
	if index < 0 {
		return 0
	}

	head := -1
	for i, reads := range g.reads {
		if len(reads) < index {
			continue
		}
		if reads[index] == nil {
			head = i
			break
		}
	}
	if head < 0 {
		g.reads = append(g.reads, make([]*packet.Packet, len(g.readers)))
		head = len(g.reads) - 1
	}

	reads := g.reads[head]
	reads[index] = pck

	if g.forward.Forward(reads) {
		g.reads = append(g.reads[:head], g.reads[head+1:]...)
	}

	count := 0
	for _, pck := range reads {
		if pck != nil {
			count++
		}
	}
	return count
}

func (g *Gateway) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.readers = nil
	g.reads = nil
}

func (h ForwardHookFunc) Forward(pcks []*packet.Packet) bool {
	return h(pcks)
}
