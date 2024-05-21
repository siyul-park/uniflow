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

	for i := 0; i < len(writers); i++ {
		if len(pcks) < i {
			break
		}
		if pcks[i] == nil || writers[i].Write(pcks[i]) == 0 {
			pcks = append(pcks[:i], pcks[i+1:]...)
			writers = append(writers[:i], writers[i+1:]...)
			i--
		}
	}

	if len(writers) == 0 {
		for _, r := range readers {
			r.Receive(packet.None)
		}
		return 0
	}

	receives := make(map[*Writer]*packet.Packet, len(writers))
	for _, w := range writers {
		receives[w] = nil
	}

	b.readers = append(b.readers, readers)
	b.receives = append(b.receives, receives)

	return len(writers)
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

	if index > 0 {
		return true
	}

	for len(b.readers) > 0 {
		readers := b.readers[0]
		receives := b.receives[0]

		for _, pck := range receives {
			if pck == nil {
				return true
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
	return true
}

func (b *Bridge) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.readers = nil
	b.receives = nil
}
