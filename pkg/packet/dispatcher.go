package packet

import (
	"sync"
)

// Dispatcher represents a data dispatcher between readers and a route hook.
type Dispatcher struct {
	readers []*Reader
	reads   [][]*Packet
	mu      sync.Mutex
}

// NewDispatcher creates a new Dispatcher instance.
func NewDispatcher(readers []*Reader) *Dispatcher {
	return &Dispatcher{
		readers: readers,
	}
}

// Read records a packet read by a reader and returns a complete set of packets if available.
func (d *Dispatcher) Read(reader *Reader, pck *Packet) []*Packet {
	d.mu.Lock()
	defer d.mu.Unlock()

	index := d.indexOfReader(reader)
	if index < 0 {
		return nil
	}

	head := d.indexOfHead(index)
	if head < 0 {
		d.reads = append(d.reads, make([]*Packet, len(d.readers)))
		head = len(d.reads) - 1
	}

	reads := d.reads[head]
	reads[index] = pck

	if head == 0 && d.isFull(reads) {
		d.reads = d.reads[1:]
		return reads
	}

	return nil
}

// Close closes the Dispatcher by clearing the stored data.
func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.readers = nil
	d.reads = nil
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

func (d *Dispatcher) isFull(packets []*Packet) bool {
	for _, p := range packets {
		if p == nil {
			return false
		}
	}
	return true
}
