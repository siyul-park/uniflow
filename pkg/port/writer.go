package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Writer struct {
	readers  []*Reader
	receives [][]*packet.Packet
	in       chan *packet.Packet
	out      chan *packet.Packet
	done     chan struct{}
	mu       sync.Mutex
}

func newWriter() *Writer {
	w := &Writer{
		in:   make(chan *packet.Packet),
		out:  make(chan *packet.Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(w.out)

		buffer := make([]*packet.Packet, 0, 2)
		for {
			pck, ok := <-w.in
			if !ok {
				return
			}

			select {
			case w.out <- pck:
				continue
			default:
			}

			buffer = append(buffer, pck)

			for len(buffer) > 0 {
				select {
				case pck = <-w.in:
					buffer = append(buffer, pck)
				case w.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return w
}

func (w *Writer) Write(pck *packet.Packet) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return 0
	default:
	}

	count := 0
	receives := make([]*packet.Packet, len(w.readers))
	for i, r := range w.readers {
		if r.write(pck, w) {
			count += 1
		} else if len(w.receives) == 0 {
			w.readers = append(w.readers[:i], w.readers[i+1:]...)
			receives = append(receives[:i], receives[i+1:]...)
		} else {
			receives[i] = packet.EOF
		}
	}

	w.receives = append(w.receives, receives)

	return count
}

func (w *Writer) Receive() <-chan *packet.Packet {
	return w.out
}

func (w *Writer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return
	default:
	}

	w.readers = nil
	w.receives = nil

	close(w.done)
	close(w.in)
}

func (w *Writer) link(reader *Reader) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.readers = append(w.readers, reader)
}

func (w *Writer) receive(pck *packet.Packet, reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return false
	default:
	}

	index := -1
	for i, r := range w.readers {
		if r == reader {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}

	head := -1
	for i, receives := range w.receives {
		if len(receives) < index {
			continue
		}
		if receives[index] == nil {
			head = i
			break
		}
	}
	if head < 0 {
		return false
	}

	w.receives[head][index] = pck

	payloads := make([]primitive.Value, 0, len(w.receives[head]))
	for _, pck := range w.receives[head] {
		if pck == nil {
			return true
		}
		if pck != packet.EOF {
			payloads = append(payloads, pck.Payload())
		}
	}

	w.receives = append(w.receives[:head], w.receives[head+1:]...)

	if len(payloads) == 0 {
		w.in <- packet.EOF
	} else if len(payloads) == 1 {
		w.in <- packet.New(payloads[0])
	} else {
		w.in <- packet.New(primitive.NewSlice(payloads...))
	}

	return true
}
