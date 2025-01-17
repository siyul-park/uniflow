package packet

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Reader represents a packet reader that manages incoming packets from multiple writers.
type Reader struct {
	writers   []*Writer
	in        chan *Packet
	out       chan *Packet
	done      bool
	inbounds  Hooks
	outbounds Hooks
	mu        sync.Mutex
}

var ClosedReader *Reader

func init() {
	ClosedReader = NewReader()
	ClosedReader.Close()
}

// NewReader creates a new Reader instance and starts its processing loop.
func NewReader() *Reader {
	r := &Reader{
		in:  make(chan *Packet),
		out: make(chan *Packet),
	}

	go func() {
		defer close(r.out)

		buffer := make([]*Packet, 0, 2)
		for pck := range r.in {
			select {
			case r.out <- pck:
			default:
				buffer = append(buffer, pck)
				for len(buffer) > 0 {
					select {
					case pck, ok := <-r.in:
						if !ok {
							return
						}
						buffer = append(buffer, pck)
					case r.out <- buffer[0]:
						buffer = buffer[1:]
					}
				}
			}
		}
	}()

	return r
}

// AddInboundHook adds a handler to process inbound packets.
func (r *Reader) AddInboundHook(hook Hook) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.done {
		return false
	}

	for _, h := range r.inbounds {
		if h == hook {
			return false
		}
	}
	r.inbounds = append(r.inbounds, hook)
	return true
}

// AddOutboundHook adds a handler to process outbound packets.
func (r *Reader) AddOutboundHook(hook Hook) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.done {
		return false
	}

	for _, h := range r.outbounds {
		if h == hook {
			return false
		}
	}
	r.outbounds = append(r.outbounds, hook)
	return true
}

// Read returns the channel for reading packets from the reader.
func (r *Reader) Read() <-chan *Packet {
	return r.out
}

// Receive receives a packet from a writer and forwards it to the reader's input channel.
func (r *Reader) Receive(pck *Packet) bool {
	r.mu.Lock()

	if len(r.writers) == 0 {
		r.mu.Unlock()
		return false
	}

	r.outbounds.Handle(pck)

	w := r.writers[0]
	r.writers = r.writers[1:]

	r.mu.Unlock()

	return w.receive(pck, r)
}

// Close closes the reader and releases its resources, stopping further packet processing.
func (r *Reader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.done {
		return
	}

	pck := New(types.NewError(ErrDroppedPacket))
	for _, w := range r.writers {
		r.outbounds.Handle(pck)
		go w.receive(pck, r)
	}

	close(r.in)

	r.done = true
	r.writers = nil
	r.inbounds = nil
	r.outbounds = nil
}

func (r *Reader) write(pck *Packet, writer *Writer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.done {
		return false
	}

	r.writers = append(r.writers, writer)
	r.inbounds.Handle(pck)
	r.in <- pck
	return true
}
