package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

type Bridge struct {
	readers  [][]*Reader
	receives []map[*Writer]*packet.Packet
	mu       sync.Mutex
}

func NewBridge() *Bridge {
	return &Bridge{}
}

func (b *Bridge) Write(pcks []*packet.Packet, readers []*Reader, writers []*Writer) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	receives := make(map[*Writer]*packet.Packet, len(writers))
	for i := 0; i < len(writers); i++ {
		if len(pcks) < i {
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

func (b *Bridge) Receive(pck *packet.Packet, writer *Writer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	index := -1
	for i, receives := range b.receives {
		if pck, ok := receives[writer]; ok && pck == nil {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}

	receives := b.receives[index]
	receives[writer] = pck

	if index == 0 {
		b.consume()
	}
	return true
}

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

		pcks := make([]*packet.Packet, 0, len(receives))
		for _, pck := range receives {
			pcks = append(pcks, pck)
		}

		pck := packet.Merge(pcks)
		for _, r := range readers {
			r.Receive(pck)
		}
	}
}
