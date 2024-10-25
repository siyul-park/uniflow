package packet

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Writer represents a packet writer that sends packets to linked readers.
type Writer struct {
	readers   []*Reader
	receives  [][]*Packet
	in        chan *Packet
	out       chan *Packet
	done      chan struct{}
	inbounds  Hooks
	outbounds Hooks
	mu        sync.Mutex
}

// Send sends a packet to the writer and returns the received packet or None if to write fails.
func Send(writer *Writer, pck *Packet) *Packet {
	return SendOrFallback(writer, pck, None)
}

// SendOrFallback sends a packet to the writer and returns the received packet or a backup packet if to write fails.
func SendOrFallback(writer *Writer, outPck *Packet, backPck *Packet) *Packet {
	if writer.Write(outPck) == 0 {
		return backPck
	}
	return <-writer.Receive()
}

// Discard discards all packets received by the writer.
func Discard(writer *Writer) {
	go func() {
		for range writer.Receive() {
		}
	}()
}

// NewWriter creates a new Writer instance and starts its processing loop.
func NewWriter() *Writer {
	w := &Writer{
		in:   make(chan *Packet),
		out:  make(chan *Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(w.out)
		defer close(w.in)

		buffer := make([]*Packet, 0, 2)
		for {
			var pck *Packet
			select {
			case pck = <-w.in:
			case <-w.done:
				return
			}

			select {
			case w.out <- pck:
			default:
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
		}
	}()

	return w
}

// AddInboundHook adds a handler to process inbound packets.
func (w *Writer) AddInboundHook(hook Hook) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return false
	default:
		for _, h := range w.inbounds {
			if h == hook {
				return false
			}
		}
		w.inbounds = append(w.inbounds, hook)
		return true
	}
}

// AddOutboundHook adds a handler to process outbound packets.
func (w *Writer) AddOutboundHook(hook Hook) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return false
	default:
		for _, h := range w.outbounds {
			if h == hook {
				return false
			}
		}
		w.outbounds = append(w.outbounds, hook)
		return true
	}
}

// Link connects a reader to the writer.
func (w *Writer) Link(reader *Reader) {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return
	default:
		w.readers = append(w.readers, reader)
	}
}

// Write writes a packet to all linked readers and returns the count of successful writes.
func (w *Writer) Write(pck *Packet) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return 0
	default:
	}

	if len(w.readers) == 0 {
		return 0
	}

	w.outbounds.Handle(pck)

	count := 0
	receives := make([]*Packet, len(w.readers))
	for i, r := range w.readers {
		if r.write(New(pck.Payload()), w) {
			count++
		} else {
			receives[i] = None
		}
	}

	if count > 0 {
		w.receives = append(w.receives, receives)
	} else {
		pck := New(types.NewError(ErrDroppedPacket))
		w.inbounds.Handle(pck)
		w.in <- pck
	}

	return count
}

// Receive returns the channel for receiving packets from the writer.
func (w *Writer) Receive() <-chan *Packet {
	return w.out
}

// Close closes the writer and releases its resources.
func (w *Writer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return
	default:
	}

	pck := New(types.NewError(ErrDroppedPacket))
	for range w.receives {
		w.inbounds.Handle(pck)
		w.in <- pck
	}

	close(w.done)

	w.readers = nil
	w.receives = nil
	w.inbounds = nil
	w.outbounds = nil
}

func (w *Writer) receive(pck *Packet, reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return false
	default:
	}

	index := w.indexOfReader(reader)
	if index < 0 {
		return false
	}

	head := w.indexOfHead(index)
	if head < 0 {
		return false
	}

	receives := w.receives[head]
	receives[index] = pck

	if head == 0 {
		for _, pck := range receives {
			if pck == nil {
				return true
			}
		}

		w.receives = w.receives[1:]

		pck := Merge(receives)
		w.inbounds.Handle(pck)
		w.in <- pck
	}

	return true
}

func (w *Writer) indexOfReader(reader *Reader) int {
	for i, r := range w.readers {
		if r == reader {
			return i
		}
	}
	return -1
}

func (w *Writer) indexOfHead(index int) int {
	for i, receives := range w.receives {
		if len(receives) < index {
			continue
		}
		if receives[index] == nil {
			return i
		}
	}
	return -1
}
