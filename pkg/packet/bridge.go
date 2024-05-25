package packet

import (
	"sync"
)

// Bridge represents a data bridge between readers and writers.
type Bridge struct {
	readers  [][]*Reader
	receives []map[*Writer]*Packet
	mu       sync.Mutex
}

// NewBridge creates a new Bridge instance.
func NewBridge() *Bridge {
	return &Bridge{}
}

// Write writes packets to writers and returns the count of successful writes.
// It also stores the received packets for each writer.
func (b *Bridge) Write(pcks []*Packet, readers []*Reader, writers []*Writer) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	receives := make(map[*Writer]*Packet, len(writers))
	for i := 0; i < len(writers); i++ {
		if len(pcks) <= i {
			break
		}

		writer := writers[i]
		pck := pcks[i]

		if pck == nil {
			continue
		}

		if writer.Write(pck) > 0 {
			receives[writer] = nil
		} else {
			receives[writer] = pck
		}
	}

	b.readers = append(b.readers, readers)
	b.receives = append(b.receives, receives)

	count := 0
	for range receives {
		count++
	}

	if len(b.readers) == 1 {
		b.consume()
	}
	return count
}

// Receive receives a packet from a writer and stores it for further processing.
// It returns true if the packet is successfully received, false otherwise.
func (b *Bridge) Receive(pck *Packet, writer *Writer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, receives := range b.receives {
		if p, ok := receives[writer]; ok && p == nil {
			receives[writer] = pck
			if i == 0 {
				b.consume()
			}
			return true
		}
	}
	return false
}

// Close closes the Bridge by clearing the stored data.
func (b *Bridge) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.readers = nil
	b.receives = nil
}

func (b *Bridge) consume() {
	for len(b.readers) > 0 {
		readers := b.readers[0]
		receives := b.receives[0]

		for _, pck := range receives {
			if pck == nil {
				return
			}
		}

		b.readers = b.readers[1:]
		b.receives = b.receives[1:]

		pcks := make([]*Packet, 0, len(receives))
		for _, pck := range receives {
			pcks = append(pcks, pck)
		}

		pck := Merge(pcks)
		for _, r := range readers {
			r.Receive(pck)
		}
	}
}
