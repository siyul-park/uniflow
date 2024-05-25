package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

// Reader represents a packet reader.
type Reader struct {
	writers []*Writer
	in      chan *packet.Packet
	out     chan *packet.Packet
	done    chan struct{}
	mu      sync.Mutex
}

// NewReader creates a new Reader instance and starts its processing loop.
func NewReader() *Reader {
	r := &Reader{
		in:   make(chan *packet.Packet),
		out:  make(chan *packet.Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(r.out)

		buffer := make([]*packet.Packet, 0, 2)
		for {
			pck, ok := <-r.in
			if !ok {
				return
			}

			select {
			case r.out <- pck:
				continue
			default:
			}

			buffer = append(buffer, pck)

			for len(buffer) > 0 {
				select {
				case pck = <-r.in:
					buffer = append(buffer, pck)
				case r.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return r
}

// Read returns the channel for reading packets from the reader.
func (r *Reader) Read() <-chan *packet.Packet {
	return r.out
}

// Receive receives a packet from a writer and forwards it to the reader's input channel.
func (r *Reader) Receive(pck *packet.Packet) bool {
	if w := r.writer(); w == nil {
		return false
	} else {
		return w.receive(pck, r)
	}
}

// Close closes the reader and releases its resources.
func (r *Reader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.done:
	default:
		r.writers = nil
		close(r.done)
		close(r.in)
	}
}

func (r *Reader) write(pck *packet.Packet, writer *Writer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.done:
		return false
	default:
		r.writers = append(r.writers, writer)
		r.in <- pck

		return true
	}
}

func (r *Reader) writer() *Writer {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.writers) == 0 {
		return nil
	}

	writer := r.writers[0]
	r.writers = r.writers[1:]

	return writer
}
