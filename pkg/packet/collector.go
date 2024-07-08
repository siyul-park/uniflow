package packet

import (
	"sync"
)

// Collector represents a system that collects packets from multiple readers and provides a complete set of packets once all readers have provided their packets.
type Collector struct {
	readers []*Reader
	reads   [][]*Packet
	mu      sync.Mutex
}

// NewCollector creates a new Collector instance initialized with the provided readers.
func NewCollector(readers []*Reader) *Collector {
	return &Collector{
		readers: readers,
	}
}

// Read records a packet read by a specific reader and returns a complete set of packets if all readers have provided their packets.
func (c *Collector) Read(reader *Reader, pck *Packet) []*Packet {
	c.mu.Lock()
	defer c.mu.Unlock()

	index := c.indexOfReader(reader)
	if index < 0 {
		return nil
	}

	head := c.indexOfHead(index)
	if head < 0 {
		c.reads = append(c.reads, make([]*Packet, len(c.readers)))
		head = len(c.reads) - 1
	}

	reads := c.reads[head]
	reads[index] = pck

	if head == 0 && c.isFull(reads) {
		c.reads = c.reads[1:]
		return reads
	}

	return nil
}

// Close clears the stored data in the Collector instance.
func (c *Collector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.readers = nil
	c.reads = nil
}

func (c *Collector) indexOfReader(reader *Reader) int {
	for i, r := range c.readers {
		if r == reader {
			return i
		}
	}
	return -1
}

func (c *Collector) indexOfHead(index int) int {
	for i, reads := range c.reads {
		if reads[index] == nil {
			return i
		}
	}
	return -1
}

func (c *Collector) isFull(pcks []*Packet) bool {
	for _, pck := range pcks {
		if pck == nil {
			return false
		}
	}
	return true
}
