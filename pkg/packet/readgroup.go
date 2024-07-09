package packet

import (
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

	index := r.findReaderIndex(reader)
	if index < 0 {
		return nil
	}

	head := r.findHeadIndex(index)
	if head < 0 {
		r.reads = append(r.reads, make([]*Packet, len(r.readers)))
		head = len(r.reads) - 1
	}

	r.reads[head][index] = pck

	if head == 0 && r.isComplete(r.reads[head]) {
		completeSet := r.reads[0]
		r.reads = r.reads[1:]
		return completeSet
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

// findReaderIndex returns the index of the reader in the readers slice.
func (r *ReadGroup) findReaderIndex(reader *Reader) int {
	for i, r := range r.readers {
		if r == reader {
			return i
		}
	}
	return -1
}

// findHeadIndex returns the index of the first incomplete set of packets for the reader.
func (r *ReadGroup) findHeadIndex(index int) int {
	for i, reads := range r.reads {
		if reads[index] == nil {
			return i
		}
	}
	return -1
}

// isComplete checks if all packets in the set are present.
func (r *ReadGroup) isComplete(pcks []*Packet) bool {
	for _, pck := range pcks {
		if pck == nil {
			return false
		}
	}
	return true
}
