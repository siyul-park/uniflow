package packet

import (
	"slices"
	"sync"
)

// ReadGroup collects packets from multiple readers, providing a complete set once all readers have submitted their packets.
type ReadGroup struct {
	readers []*Reader
	reads   [][]*Packet
	mu      sync.Mutex
}

// NewReadGroup initializes a ReadGroup with the given readers.
func NewReadGroup(readers []*Reader) *ReadGroup {
	return &ReadGroup{
		readers: readers,
	}
}

// Read collects a packet from the given reader. Returns all packets once complete.
func (r *ReadGroup) Read(reader *Reader, pck *Packet) []*Packet {
	r.mu.Lock()
	defer r.mu.Unlock()

	index := -1
	for i, r := range r.readers {
		if r == reader {
			index = i
			break
		}
	}
	if index < 0 {
		return nil
	}

	head := -1
	for i, reads := range r.reads {
		if reads[index] == nil {
			head = i
			break
		}
	}
	if head < 0 {
		r.reads = append(r.reads, make([]*Packet, len(r.readers)))
		head = len(r.reads) - 1
	}

	r.reads[head][index] = pck

	if head == 0 && !slices.Contains(r.reads[head], nil) {
		read := r.reads[0]
		r.reads = r.reads[1:]
		return read
	}
	return nil
}

// Close clears the stored data in the ReadGroup.
func (r *ReadGroup) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.readers = nil
	r.reads = nil
}
