package packet

import (
	"sync"
)

// Bridge represents a data bridge between readers and writers.
type Bridge struct {
	sources  [][]*Reader
	targets  [][]*Writer
	receives [][]*Packet
	mu       sync.Mutex
}

// NewBridge creates a new Bridge instance.
func NewBridge() *Bridge {
	return &Bridge{}
}

// Write writes packets to the specified writers and tracks the success of each write.
// It stores received packets for further processing.
func (b *Bridge) Write(pcks []*Packet, sources []*Reader, targets []*Writer) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(targets) > len(pcks) {
		targets = targets[:len(pcks)]
	}
	if len(pcks) > len(targets) {
		pcks = pcks[:len(targets)]
	}

	for i := 0; i < len(pcks); i++ {
		if pcks[i] == nil {
			pcks = append(pcks[:i], pcks[i+1:]...)
			targets = append(targets[:i], targets[i+1:]...)
			i--
		}
	}

	receives := make([]*Packet, len(targets))
	count := 0
	for i, writer := range targets {
		pck := pcks[i]
		if writer.Write(pck) > 0 {
			receives[i] = nil
			count++
		} else {
			receives[i] = pck
		}
	}

	b.sources = append(b.sources, sources)
	b.targets = append(b.targets, targets)
	b.receives = append(b.receives, receives)

	if len(b.sources) == 1 {
		b.consume()
	}
	return count
}

// Rewrite attempts to rewrite a packet from a source writer to a target writer.
// Returns true if the rewrite is successful, false otherwise.
func (b *Bridge) Rewrite(pck *Packet, source *Writer, target *Writer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, targets := range b.targets {
		for j, writer := range targets {
			if writer == source && b.receives[i][j] == nil {
				if target == nil {
					b.targets[i] = append(b.targets[i][:j], b.targets[i][j+1:]...)
					b.receives[i] = append(b.receives[i][:j], b.receives[i][j+1:]...)
					j--
				} else {
					targets[j] = target

					if target.Write(pck) > 0 {
						b.receives[i][j] = nil
					} else {
						b.receives[i][j] = pck
					}
				}

				if i == 0 {
					b.consume()
				}
				return b.receives[i][j] == nil
			}
		}
	}
	return false
}

// Receive accepts a packet from a writer and stores it for further processing.
// Returns true if the packet is successfully stored, false otherwise.
func (b *Bridge) Receive(pck *Packet, target *Writer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, targets := range b.targets {
		for j, writer := range targets {
			if writer == target && b.receives[i][j] == nil {
				b.receives[i][j] = pck
				if i == 0 {
					b.consume()
				}
				return true
			}
		}
	}
	return false
}

// Close clears all stored data in the Bridge.
func (b *Bridge) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.sources = nil
	b.receives = nil
}

func (b *Bridge) consume() {
	for len(b.sources) > 0 {
		sources := b.sources[0]
		receives := b.receives[0]

		for _, pck := range receives {
			if pck == nil {
				return
			}
		}

		b.sources = b.sources[1:]
		b.targets = b.targets[1:]
		b.receives = b.receives[1:]

		pcks := make([]*Packet, 0, len(receives))
		for _, pck := range receives {
			pcks = append(pcks, pck)
		}

		merged := Merge(pcks)
		for _, reader := range sources {
			reader.Receive(merged)
		}
	}
}
