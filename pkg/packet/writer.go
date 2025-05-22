package packet

import (
	"slices"
	"sync"
)

// Writer represents a packet writer that sends packets to linked readers.
type Writer struct {
	readers   []*Reader
	receives  [][]*Packet
	in        chan *Packet
	out       chan *Packet
	done      bool
	inbounds  Hooks
	outbounds Hooks
	mu        sync.RWMutex
}

var ClosedWriter *Writer

func init() {
	ClosedWriter = NewWriter()
	ClosedWriter.Close()
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

// NewWriter creates a new Writer instance and starts its processing loop.
func NewWriter() *Writer {
	w := &Writer{
		in:  make(chan *Packet),
		out: make(chan *Packet),
	}

	go func() {
		defer close(w.out)

		buffer := make([]*Packet, 0, 2)
		for pck := range w.in {
			select {
			case w.out <- pck:
			default:
				buffer = append(buffer, pck)
				for len(buffer) > 0 {
					select {
					case pck, ok := <-w.in:
						if !ok {
							return
						}
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

	if w.done {
		return false
	}

	for _, h := range w.inbounds {
		if h == hook {
			return false
		}
	}
	w.inbounds = append(w.inbounds, hook)
	return true
}

// AddOutboundHook adds a handler to process outbound packets.
func (w *Writer) AddOutboundHook(hook Hook) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.done {
		return false
	}

	for _, h := range w.outbounds {
		if h == hook {
			return false
		}
	}
	w.outbounds = append(w.outbounds, hook)
	return true
}

// Links returns a list of readers linked to the writer.
func (w *Writer) Links() []*Reader {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return append([]*Reader(nil), w.readers...)
}

// Link connects a reader to the writer.
func (w *Writer) Link(reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.done {
		return false
	}

	for _, r := range w.readers {
		if r == reader {
			return false
		}
	}
	w.readers = append(w.readers, reader)
	return true
}

// Unlink removes the given reader from the writer, ensuring proper disconnection.
func (w *Writer) Unlink(reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.done {
		return false
	}

	for i, r := range w.readers {
		if r == reader {
			w.readers = append(w.readers[:i], w.readers[i+1:]...)

			for j := range w.receives {
				w.receives[j] = append(w.receives[j][:i], w.receives[j][i+1:]...)
			}

			for len(w.receives) > 0 && !slices.Contains(w.receives[0], nil) {
				pck := New(ErrDroppedPacket)
				if len(w.receives[0]) > 0 {
					pck = Join(w.receives[0]...)
				}

				w.inbounds.Handle(pck)

				w.receives = w.receives[1:]
				w.in <- pck
			}
			return true
		}
	}
	return false
}

// Write writes a packet to all linked readers and returns the count of successful writes.
func (w *Writer) Write(pck *Packet) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.done || len(w.readers) == 0 {
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
		pck := New(ErrDroppedPacket)
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

	if w.done {
		return
	}

	pck := New(ErrDroppedPacket)
	for range w.receives {
		w.inbounds.Handle(pck)
		w.in <- pck
	}

	close(w.in)

	w.done = true
	w.readers = nil
	w.receives = nil
	w.inbounds = nil
	w.outbounds = nil
}

func (w *Writer) receive(pck *Packet, reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.done {
		return false
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
		if slices.Contains(receives, nil) {
			return true
		}

		w.receives = w.receives[1:]

		pck := Join(receives...)
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
