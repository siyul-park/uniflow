package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

// Dispatcher represents a data dispatcher between readers and a route hook.
type Dispatcher struct {
	readers []*Reader
	reads   [][]*packet.Packet
	route   RouteHook
	mu      sync.Mutex
}

// RouteHook represents a function that routes packets.
type RouteHook interface {
	Route(pcks []*packet.Packet) bool
}

// RouteHookFunc is an adapter to allow the use of ordinary functions as RouteHooks.
type RouteHookFunc func(pcks []*packet.Packet) bool

// NewDispatcher creates a new Dispatcher instance.
func NewDispatcher(readers []*Reader, route RouteHook) *Dispatcher {
	return &Dispatcher{
		readers: readers,
		route:   route,
	}
}

// Write writes a packet to a specific reader and returns the count of successful writes.
func (d *Dispatcher) Write(pck *packet.Packet, reader *Reader) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	index := d.indexOfReader(reader)
	if index < 0 {
		return 0
	}

	head := d.indexOfHead(index)
	if head < 0 {
		d.reads = append(d.reads, make([]*packet.Packet, len(d.readers)))
		head = len(d.reads) - 1
	}

	reads := d.reads[head]
	reads[index] = pck

	if head == 0 {
		d.consume()
	}

	count := 0
	for _, pck := range reads {
		if pck != nil {
			count++
		}
	}
	return count
}

// Close closes the Dispatcher by clearing the stored data.
func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.readers = nil
	d.reads = nil
}

// consume processes the stored packets.
func (d *Dispatcher) consume() {
	for len(d.reads) > 0 {
		reads := d.reads[0]

		count := 0
		for _, pck := range reads {
			if pck != nil {
				count++
			}
		}

		if d.route.Route(reads) {
			d.reads = d.reads[1:]
		} else if count == len(reads) {
			d.reads = d.reads[1:]
			for _, r := range d.readers {
				r.Receive(packet.None)
			}
		} else {
			break
		}
	}
}

func (d *Dispatcher) indexOfReader(reader *Reader) int {
	for i, r := range d.readers {
		if r == reader {
			return i
		}
	}
	return -1
}

func (d *Dispatcher) indexOfHead(index int) int {
	for i, reads := range d.reads {
		if reads[index] == nil {
			return i
		}
	}
	return -1
}

// Route forwards packets using a RouteHook function.
func (h RouteHookFunc) Route(pcks []*packet.Packet) bool {
	return h(pcks)
}
